/**
 * Copyright 2018 Planet Labs Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package authenticate

import (
	"encoding/json"
	"io"
	"net/http"

	"k8s.io/api/authentication/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/negz/kubehook/auth"
	"github.com/pkg/errors"
)

const (
	authv1Beta1 = "authentication.k8s.io/v1beta1"
	tokenReview = "TokenReview"
)

// Handler returns an HTTP handler function that handles an authentication
// webhook using the supplied Authenticator.
func Handler(a auth.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		t, err := extractToken(r.Body)
		if err != nil {
			write(w, v1beta1.TokenReviewStatus{Error: err.Error()}, http.StatusBadRequest)
			return
		}

		u, err := a.Authenticate(t)
		if err != nil {
			write(w, v1beta1.TokenReviewStatus{Error: err.Error()}, http.StatusForbidden)
			return
		}

		write(w, tokenReviewStatus(u), http.StatusOK)
	}
}

func extractToken(b io.Reader) (string, error) {
	req := &v1beta1.TokenReview{}
	err := json.NewDecoder(b).Decode(req)
	switch {
	case err != nil:
		return "", errors.Wrap(err, "cannot parse token request")
	case req.APIVersion != authv1Beta1:
		return "", errors.Errorf("unsupported API version %s", req.APIVersion)
	case req.Kind != tokenReview:
		return "", errors.Errorf("unsupported Kind %s", req.Kind)
	case req.Spec.Token == "":
		return "", errors.New("missing token")
	}
	return req.Spec.Token, nil
}

func write(w http.ResponseWriter, trStatus v1beta1.TokenReviewStatus, httpStatus int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(v1beta1.TokenReview{
		TypeMeta:   v1.TypeMeta{APIVersion: authv1Beta1, Kind: tokenReview},
		ObjectMeta: v1.ObjectMeta{CreationTimestamp: v1.Now()},
		Status:     trStatus,
	})
}

func tokenReviewStatus(u *auth.User) v1beta1.TokenReviewStatus {
	return v1beta1.TokenReviewStatus{
		Authenticated: true,
		User: v1beta1.UserInfo{
			Username: u.Username,
			UID:      u.UID,
			Groups:   u.Groups,
		},
	}
}
