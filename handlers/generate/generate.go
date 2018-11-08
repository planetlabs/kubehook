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

package generate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/planetlabs/kubehook/auth"
	"github.com/planetlabs/kubehook/handlers"
	"github.com/planetlabs/kubehook/lifetime"

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

		u := r.Header.Get(h.User)
		if u == "" {
			write(w, rsp{Error: fmt.Sprintf("cannot extract username from header %s", h.User)}, http.StatusBadRequest)
			return
		}
		gs := strings.Split(r.Header.Get(h.Group), h.GroupDelimiter)
		t, err := g.Generate(&auth.User{Username: u, Groups: gs}, time.Duration(req.Lifetime))
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
