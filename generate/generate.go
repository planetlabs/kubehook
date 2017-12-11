package generate

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/negz/kubehook/auth"
	"github.com/negz/kubehook/generate/lifetime"

	"github.com/pkg/errors"
)

// DefaultUserHeader specifies the default header used to determine the
// currently authenticated user.
const DefaultUserHeader = "X-Forwarded-User"

type req struct {
	Lifetime lifetime.Duration `json:"lifetime"`
}

type rsp struct {
	Token string `json:"token,omitempty"`
	Error string `json:"error,omitempty"`
}

func badRequest(w http.ResponseWriter, err error) {
	j, jerr := json.Marshal(&rsp{Error: err.Error()})
	if jerr != nil {
		http.Error(w, jerr.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	w.Write(j)
}

// Handler returns an HTTP handler function that generates a JSON web token for
// the requesting user.
func Handler(g auth.Generator, userHeader string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		u := r.Header.Get(userHeader)
		if u == "" {
			badRequest(w, errors.Errorf("cannot extract username from header %s", userHeader))
			return
		}

		req := &req{}
		err := json.NewDecoder(r.Body).Decode(req)
		if err != nil {
			badRequest(w, errors.Wrap(err, "cannot parse JSON request body"))
			return
		}
		if req.Lifetime == 0 {
			badRequest(w, errors.New("must specify desired token lifetime"))
			return
		}

		// TODO(negz): Extract groups from header?
		t, err := g.Generate(&auth.User{Username: u}, time.Duration(req.Lifetime))
		if err != nil {
			badRequest(w, errors.Wrap(err, "cannot generate JSON Web Token"))
			return
		}

		j, jerr := json.Marshal(&rsp{Token: t})
		if jerr != nil {
			http.Error(w, jerr.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(j)
		return
	}
}
