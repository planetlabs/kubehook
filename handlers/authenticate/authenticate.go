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

package authenticate

import (
	"encoding/json"
	"net/http"

	"k8s.io/api/authentication/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/planetlabs/kubehook/auth"
)

const (
	tokenReview = "TokenReview"
)

type timeProvider func() v1.Time

// Handler returns an HTTP handler function that handles an authentication
// webhook using the supplied Authenticator.
func Handler(a auth.Authenticator, tokenVersion string) http.HandlerFunc {
	var proc processor

	if tokenVersion == v1beta1.SchemeGroupVersion.Version {
		proc = newBeta1Processor(v1.Now)
	} else {
		proc = newReleaseProcessor(v1.Now)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		t, err := proc.ExtractToken(r.Body)
		if err != nil {
			write(w, proc.CreateErrorStatus(err), http.StatusBadRequest)
			return
		}

		u, err := a.Authenticate(t)
		if err != nil {
			write(w, proc.CreateErrorStatus(err), http.StatusForbidden)
			return
		}

		write(w, proc.CreateReviewStatus(u), http.StatusOK)
	}
}

func write(w http.ResponseWriter, data interface{}, httpStatus int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(data) // nolint: gosec
}
