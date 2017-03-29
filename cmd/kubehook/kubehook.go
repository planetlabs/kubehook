package main

import (
	"crypto/tls"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"

	"github.com/negz/kubehook/auth/dynamo"
	"github.com/negz/kubehook/hook"

	"github.com/facebookgo/httpdown"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func logReq(fn http.HandlerFunc, log *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("request",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.String("addr", r.RemoteAddr))
		fn(w, r)
	}
}

func main() {
	var (
		app    = kingpin.New(filepath.Base(os.Args[0]), "Authenticates Kubernetes users.").DefaultEnvars()
		listen = app.Flag("listen", "Address at which to expose HTTP webhook.").Default(":10003").String()
		debug  = app.Flag("debug", "Run with debug logging.").Short('d').Bool()
		stop   = app.Flag("close-after", "Wait this long at shutdown before closing HTTP connections.").Default("1m").Duration()
		kill   = app.Flag("kill-after", "Wait this long at shutdown before exiting.").Default("2m").Duration()
		tlsCrt = app.Flag("tls-cert", "TLS certificate file.").ExistingFile()
		tlsKey = app.Flag("tls-key", "TLS private key file.").ExistingFile()

		dynamoEndpoint = app.Flag("dynamo-endpoint", "DynamoDb endpoint.").String()
		dynamoRegion   = app.Flag("dynamo-region", "DynamoDb region.").Default("us-east-1").String()
		dynamoTable    = app.Flag("dynamo-table", "DynamoDb user table.").Default(dynamo.DefaultUserTable).String()
	)

	kingpin.MustParse(app.Parse(os.Args[1:]))

	var log *zap.Logger
	log, err := zap.NewProduction()
	if *debug {
		log, err = zap.NewDevelopment()
	}
	kingpin.FatalIfError(err, "cannot create log")

	cfg := aws.NewConfig().WithRegion(*dynamoRegion)
	if *dynamoEndpoint != "" {
		log.Info("explicit DynamoDb index endpoint", zap.String("endpoint", *dynamoEndpoint))
		cfg = cfg.WithEndpoint(*dynamoEndpoint)
	}
	session, err := session.NewSession(cfg)
	kingpin.FatalIfError(err, "cannot connect to AWS")

	a, err := dynamo.NewAuthenticator(
		dynamodb.New(session),
		dynamo.Logger(log),
		dynamo.UserTable(*dynamoTable))
	kingpin.FatalIfError(err, "cannot connect to DynamoDb")

	r := httprouter.New()
	r.HandlerFunc("GET", "/authenticate", logReq(hook.Handler(a), log))
	r.HandlerFunc("GET", "/quitquitquit", logReq(func(_ http.ResponseWriter, _ *http.Request) { os.Exit(0) }, log))

	hd := &httpdown.HTTP{StopTimeout: *stop, KillTimeout: *kill}
	http := &http.Server{Addr: *listen, Handler: r}
	if *tlsCrt != "" && *tlsKey != "" {
		crt, err := tls.LoadX509KeyPair(*tlsCrt, *tlsKey)
		kingpin.FatalIfError(err, "cannot parse TLS certificate and private key")
		http.TLSConfig = &tls.Config{Certificates: []tls.Certificate{crt}}
	}

	kingpin.FatalIfError(httpdown.ListenAndServe(http, hd), "HTTP server error")
}
