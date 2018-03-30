package kubecfg

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/negz/kubehook/auth"
	"github.com/negz/kubehook/handlers"
	"github.com/negz/kubehook/lifetime"

	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

const (
	templateUser       = "kubehook"
	queryParamLifetime = "lifetime"
)

// LoadTemplate loads a kubeconfig template from a file.
func LoadTemplate(filename string) (*api.Config, error) {
	c, err := clientcmd.LoadFromFile(filename)
	return c, errors.Wrapf(err, "cannot load template from %v", filename)
}

// Handler returns an HTTP handler function that generates a kubeconfig file
// preconfigured with a set of clusters and a JSON Web Token for the requesting
// user.
func Handler(g auth.Generator, template *api.Config, h handlers.AuthHeaders) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		l, err := lifetime.ParseDuration(r.URL.Query().Get(queryParamLifetime))
		if err != nil {
			http.Error(w, errors.Wrapf(err, "cannot parse query parameter %v", queryParamLifetime).Error(), http.StatusBadRequest)
			return
		}

		u := r.Header.Get(h.User)
		if u == "" {
			http.Error(w, fmt.Sprintf("cannot extract username from header %s", h.User), http.StatusBadRequest)
			return
		}
		gs := strings.Split(r.Header.Get(h.Group), h.GroupDelimiter)
		t, err := g.Generate(&auth.User{Username: u, Groups: gs}, time.Duration(l))
		if err != nil {
			http.Error(w, errors.Wrap(err, "cannot generate token").Error(), http.StatusInternalServerError)
			return
		}

		y, err := clientcmd.Write(populateUser(template, templateUser, t))
		if err != nil {
			http.Error(w, errors.Wrap(err, "cannot marshal template to YAML").Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/x-yaml; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment")
		w.Write(y)
	}
}

func populateUser(cfg *api.Config, username, token string) api.Config {
	c := api.Config{}
	c.AuthInfos = make(map[string]*api.AuthInfo)
	c.Clusters = make(map[string]*api.Cluster)
	c.Contexts = make(map[string]*api.Context)
	c.AuthInfos[username] = &api.AuthInfo{
		Token: token,
	}
	for name, cluster := range cfg.Clusters {
		c.Clusters[name] = cluster
		c.Contexts[name] = &api.Context{Cluster: name, AuthInfo: templateUser}
	}
	return c
}
