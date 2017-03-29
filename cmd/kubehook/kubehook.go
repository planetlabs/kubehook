package main

import (
	"crypto/tls"
	"net/http"
	"os"
	"path/filepath"

	"github.com/negz/kubehook/auth/noop"
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
		groups = app.Flag("group", "Group to grant all requests.").Default("engineering").Strings()
		tlsCrt = app.Flag("tls-cert", "TLS certificate file.").ExistingFile()
		tlsKey = app.Flag("tls-key", "TLS private key file.").ExistingFile()
	)

	kingpin.MustParse(app.Parse(os.Args[1:]))

	var log *zap.Logger
	log, err := zap.NewProduction()
	if *debug {
		log, err = zap.NewDevelopment()
	}
	kingpin.FatalIfError(err, "cannot create log")

	a, err := noop.NewAuthenticator(*groups, noop.Logger(log))
	kingpin.FatalIfError(err, "cannot create noop authenticator")

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
