package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/negz/kubehook/auth/jwt"
	"github.com/negz/kubehook/generate"
	"github.com/negz/kubehook/hook"

	"github.com/facebookgo/httpdown"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func logReq(fn http.HandlerFunc, log *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("request",
			zap.String("host", r.Host),
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.String("agent", r.UserAgent()),
			zap.String("addr", r.RemoteAddr))
		fn(w, r)
	}
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
		stop     = app.Flag("close-after", "Wait this long at shutdown before closing HTTP connections.").Default("1m").Duration()
		kill     = app.Flag("kill-after", "Wait this long at shutdown before exiting.").Default("2m").Duration()
		audience = app.Flag("audience", "Audience for JWT HMAC creation and verification.").Default(jwt.DefaultAudience).String()
		header   = app.Flag("user-header", "HTTP header specifying the authenticated user sending a token generation request.").Default(generate.DefaultUserHeader).String()
		maxlife  = app.Flag("max-lifetime", "Maximum allowed JWT lifetime, in Go's time.ParseDuration format.").Default(jwt.DefaultMaxLifetime.String()).Duration()

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
	r.HandlerFunc("POST", "/generate", logReq(generate.Handler(m, *header), log))
	r.HandlerFunc("GET", "/authenticate", logReq(hook.Handler(m), log))
	r.HandlerFunc("GET", "/quitquitquit", logReq(func(_ http.ResponseWriter, _ *http.Request) { os.Exit(0) }, log))
	r.HandlerFunc("GET", "/healthz", logReq(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.WriteHeader(http.StatusOK)
	}, log))

	hd := &httpdown.HTTP{StopTimeout: *stop, KillTimeout: *kill}
	http := &http.Server{Addr: *listen, Handler: r}

	kingpin.FatalIfError(httpdown.ListenAndServe(http, hd), "HTTP server error")
}
