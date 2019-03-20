/*
Copyright 2018 Planet Labs Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
implied. See the License for the specific language governing permissions
and limitations under the License.
*/

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/planetlabs/kubehook/auth/jwt"
	"github.com/planetlabs/kubehook/handlers"
	"github.com/planetlabs/kubehook/handlers/authenticate"
	"github.com/planetlabs/kubehook/handlers/generate"
	"github.com/planetlabs/kubehook/handlers/kubecfg"
	_ "github.com/planetlabs/kubehook/statik"

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

func listenAndServe(s *http.Server, tlsCert, tlsKey string) error {
	var err error

	if tlsCert != "" && tlsKey != "" {
		err = s.ListenAndServeTLS(tlsCert, tlsKey)
	} else {
		err = s.ListenAndServe()
	}

	return err
}

func makeTLSConfig(clientCA, clientCASubject string) (*tls.Config, error) {
	tlsConfig := &tls.Config{}

	if clientCASubject != "" {
		tlsConfig.VerifyPeerCertificate = func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			err := fmt.Errorf("No verified certificates")

			for _, chain := range verifiedChains {
				certificate := chain[0]
				err = certificate.VerifyHostname(clientCASubject)
				if err == nil {
					return nil
				}
			}

			return err
		}
	}

	if clientCA != "" {
		var clientCACert []byte

		// Suppress linter warning related to file inclusion since we are
		// attempting to load a CA file specified by the user.
		clientCACert, err := ioutil.ReadFile(clientCA) // nolint: gosec
		if err != nil {
			return nil, err
		}

		clientCertPool := x509.NewCertPool()
		clientCertPool.AppendCertsFromPEM(clientCACert)

		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		tlsConfig.ClientCAs = clientCertPool

		tlsConfig.BuildNameToCertificate()
	}

	return tlsConfig, nil
}

func main() {
	var (
		app              = kingpin.New(filepath.Base(os.Args[0]), "Authenticates Kubernetes users via JWT tokens.").DefaultEnvars()
		listen           = app.Flag("listen", "Address at which to expose HTTP webhook.").Default(":10003").String()
		debug            = app.Flag("debug", "Run with debug logging.").Short('d').Bool()
		grace            = app.Flag("shutdown-grace-period", "Wait this long for sessions to end before shutting down.").Default("1m").Duration()
		audience         = app.Flag("audience", "Audience for JWT HMAC creation and verification.").Default(jwt.DefaultAudience).String()
		userHeader       = app.Flag("user-header", "HTTP header specifying the authenticated user sending a token generation request.").Default(handlers.DefaultUserHeader).String()
		groupHeader      = app.Flag("group-header", "HTTP header specifying the authenticated user's groups.").Default(handlers.DefaultGroupHeader).String()
		groupHeaderDelim = app.Flag("group-header-delimiter", "Delimiter separating group names in the group-header.").Default(handlers.DefaultGroupHeaderDelimiter).String()
		maxlife          = app.Flag("max-lifetime", "Maximum allowed JWT lifetime, in Go's time.ParseDuration format.").Default(jwt.DefaultMaxLifetime.String()).Duration()
		template         = app.Flag("kubecfg-template", "A kubecfg file containing clusters to populate with a user and contexts.").ExistingFile()
		clientCA         = app.Flag("client-ca", "If set, enables mutual TLS and specifies the path to CA file to use when validating client connections.").String()
		clientCASubject  = app.Flag("client-ca-subject", "If set, requires that the client CA matches the provided subject (requires --client-ca).").String()
		tlsCert          = app.Flag("tls-cert", "If set, enables TLS and specifies the path to TLS certificate to use for HTTPS server (requires --tls-key).").String()
		tlsKey           = app.Flag("tls-key", "Path to TLS key to use for HTTPS server (requires --tls-cert).").String()

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

	tlsConfig, err := makeTLSConfig(*clientCA, *clientCASubject)
	kingpin.FatalIfError(err, "initializing TLS configuration")

	s := &http.Server{Addr: *listen, Handler: logRequests(r, log), TLSConfig: tlsConfig,}

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

	h := handlers.AuthHeaders{
		User:           *userHeader,
		Group:          *groupHeader,
		GroupDelimiter: *groupHeaderDelim,
	}

	r.ServeFiles("/dist/*filepath", frontend)
	r.HandlerFunc("GET", "/", handlers.Content(index, filepath.Base(indexPath)))
	r.HandlerFunc("POST", "/generate", generate.Handler(m, h))
	r.HandlerFunc("POST", "/authenticate", authenticate.Handler(m))
	r.HandlerFunc("GET", "/quitquitquit", handlers.Run(shutdown))
	r.HandlerFunc("GET", "/healthz", handlers.Ping())

	if *template != "" {
		t, err := kubecfg.LoadTemplate(*template)
		kingpin.FatalIfError(err, "cannot load kubeconfig template")
		r.HandlerFunc("GET", "/kubecfg", kubecfg.Handler(m, t, h))
	} else {
		r.HandlerFunc("GET", "/kubecfg", handlers.NotImplemented())
	}

	log.Info("shutdown", zap.Error(listenAndServe(s, *tlsCert, *tlsKey)))

	<-done
	cancel()
}
