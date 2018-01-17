package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/negz/kubehook/auth/jwt"
	"github.com/negz/kubehook/handlers/authenticate"
	"github.com/negz/kubehook/handlers/generate"
	"github.com/negz/kubehook/handlers/kubecfg"
	"github.com/negz/kubehook/handlers/util"

	_ "github.com/negz/kubehook/statik"

	"github.com/julienschmidt/httprouter"
	"github.com/rakyll/statik/fs"
	"go.uber.org/zap"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const indexPath = "/index.html"

func logRequests(h http.Handler, log *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("request",
			zap.String("host", r.Host),
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.String("agent", r.UserAgent()),
			zap.String("addr", r.RemoteAddr))
		log.Debug("request", zap.Any("headers", r.Header))
		h.ServeHTTP(w, r)
	})
}

// envVarName returns DefaultEnvars style env var names. It can be used for
// args, which DefaultEnvars does not seem to setup.
func envVarName(app, arg string) string {
	return strings.Replace(strings.ToUpper(fmt.Sprintf("%s_%s", app, arg)), "-", "_", -1)
}

func main() {
	var (
		app      = kingpin.New(filepath.Base(os.Args[0]), "Authenticates Kubernetes users via JWT tokens.").DefaultEnvars()
		listen   = app.Flag("listen", "Address at which to expose HTTP webhook.").Default(":10003").String()
		debug    = app.Flag("debug", "Run with debug logging.").Short('d').Bool()
		grace    = app.Flag("shutdown-grace-period", "Wait this long for sessions to end before shutting down.").Default("1m").Duration()
		audience = app.Flag("audience", "Audience for JWT HMAC creation and verification.").Default(jwt.DefaultAudience).String()
		header   = app.Flag("user-header", "HTTP header specifying the authenticated user sending a token generation request.").Default(generate.DefaultUserHeader).String()
		maxlife  = app.Flag("max-lifetime", "Maximum allowed JWT lifetime, in Go's time.ParseDuration format.").Default(jwt.DefaultMaxLifetime.String()).Duration()
		template = app.Flag("kubecfg-template", "A kubecfg file containing clusters to populate with a user and contexts.").ExistingFile()

		secret = app.Arg("secret", "Secret for JWT HMAC signature and verification.").Required().Envar(envVarName(app.Name, "secret")).String()
	)

	kingpin.MustParse(app.Parse(os.Args[1:]))

	var log *zap.Logger
	log, err := zap.NewProduction()
	if *debug {
		log, err = zap.NewDevelopment()
	}
	kingpin.FatalIfError(err, "cannot create log")

	m, err := jwt.NewManager([]byte(*secret), jwt.Audience(*audience), jwt.MaxLifetime(*maxlife), jwt.Logger(log))
	kingpin.FatalIfError(err, "cannot create JWT authenticator")

	r := httprouter.New()
	s := &http.Server{Addr: *listen, Handler: logRequests(r, log)}

	ctx, cancel := context.WithTimeout(context.Background(), *grace)
	done := make(chan struct{})
	shutdown := func() {
		log.Info("shutdown", zap.Error(s.Shutdown(ctx)))
		close(done)
	}

	go func() {
		sigterm := make(chan os.Signal, 1)
		signal.Notify(sigterm, syscall.SIGTERM)
		<-sigterm
		shutdown()
	}()

	frontend, err := fs.New()
	kingpin.FatalIfError(err, "cannot load frontend")

	index, err := frontend.Open(indexPath)
	kingpin.FatalIfError(err, "cannot open frontend index %s", indexPath)

	r.ServeFiles("/dist/*filepath", frontend)
	r.HandlerFunc("GET", "/", util.Content(index, filepath.Base(indexPath)))
	r.HandlerFunc("POST", "/generate", generate.Handler(m, *header))
	r.HandlerFunc("POST", "/authenticate", authenticate.Handler(m))
	r.HandlerFunc("GET", "/quitquitquit", util.Run(shutdown))
	r.HandlerFunc("GET", "/healthz", util.Ping())
	r.HandlerFunc("GET", "/kubecfg", util.NotImplemented())

	if *template != "" {
		tmpl, err := kubecfg.LoadTemplate(*template)
		kingpin.FatalIfError(err, "cannot load kubeconfig template")
		r.HandlerFunc("GET", "/kubecfg", kubecfg.Handler(m, *header, tmpl))
	}

	log.Info("shutdown", zap.Error(s.ListenAndServe()))
	<-done
	cancel()
}
