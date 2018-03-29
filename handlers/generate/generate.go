package generate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/negz/kubehook/auth"
	"github.com/negz/kubehook/handlers"
	"github.com/negz/kubehook/lifetime"

	"github.com/pkg/errors"
)

type req struct {
	Lifetime lifetime.Duration `json:"lifetime"`
}

type rsp struct {
	Token string `json:"token,omitempty"`
	Error string `json:"error,omitempty"`
}

// Handler returns an HTTP handler function that generates a JSON web token for
// the requesting user.
func Handler(g auth.Generator, h handlers.AuthHeaders) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		u := r.Header.Get(h.User)
		if u == "" {
			write(w, rsp{Error: fmt.Sprintf("cannot extract username from header %s", h.User)}, http.StatusBadRequest)
			return
		}

		req := &req{}
		err := json.NewDecoder(r.Body).Decode(req)
		if err != nil {
			write(w, rsp{Error: errors.Wrap(err, "cannot parse JSON request body").Error()}, http.StatusBadRequest)
			return
		}
		if req.Lifetime == 0 {
			write(w, rsp{Error: "must specify desired token lifetime"}, http.StatusBadRequest)
			return
		}

		// TODO(negz): Extract groups from header?
		t, err := g.Generate(&auth.User{Username: u}, time.Duration(req.Lifetime))
		if err != nil {
			write(w, rsp{Error: errors.Wrap(err, "cannot generate token").Error()}, http.StatusInternalServerError)
			return
		}

		write(w, rsp{Token: t}, http.StatusOK)
	}
}

func write(w http.ResponseWriter, r rsp, httpStatus int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(r)
}
