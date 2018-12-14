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
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-test/deep"
	"github.com/negz/kubehook/auth"
	"github.com/pkg/errors"

	"k8s.io/api/authentication/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

type predictableAuthenticator struct {
	err error
}

func (a *predictableAuthenticator) Authenticate(token string) (*auth.User, error) {
	return testUser(token).Auth(), a.err
}

type tu struct {
	u string
}

func testUser(u string) *tu {
	return &tu{u}
}

func (t *tu) Auth() *auth.User {
	return &auth.User{Username: t.u, UID: "test/" + t.u, Groups: []string{"test"}}
}

func (t *tu) UserInfo() v1beta1.UserInfo {
	return v1beta1.UserInfo{Username: t.u, UID: "test/" + t.u, Groups: []string{"test"}}
}

func TestHandler(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		req        *v1beta1.TokenReview
		trStatus   v1beta1.TokenReviewStatus
		httpStatus int
	}{
		{
			name: "Success",
			req: &v1beta1.TokenReview{
				TypeMeta:   v1.TypeMeta{APIVersion: authv1Beta1, Kind: tokenReview},
				ObjectMeta: v1.ObjectMeta{CreationTimestamp: v1.Now()},
				Spec:       v1beta1.TokenReviewSpec{Token: "token"},
			},
			trStatus: v1beta1.TokenReviewStatus{
				Authenticated: true,
				User:          testUser("token").UserInfo(),
			},
			httpStatus: http.StatusOK,
		},
		{
			name: "AuthFailed",
			err:  errors.New("bad token"),
			req: &v1beta1.TokenReview{
				TypeMeta:   v1.TypeMeta{APIVersion: authv1Beta1, Kind: tokenReview},
				ObjectMeta: v1.ObjectMeta{CreationTimestamp: v1.Now()},
				Spec:       v1beta1.TokenReviewSpec{Token: "badToken"},
			},
			trStatus:   v1beta1.TokenReviewStatus{Error: "bad token"},
			httpStatus: http.StatusForbidden,
		},
		{
			name: "BadAPIVersion",
			req: &v1beta1.TokenReview{
				TypeMeta:   v1.TypeMeta{APIVersion: "auth/v2", Kind: tokenReview},
				ObjectMeta: v1.ObjectMeta{CreationTimestamp: v1.Now()},
				Spec:       v1beta1.TokenReviewSpec{Token: "badToken"},
			},
			trStatus:   v1beta1.TokenReviewStatus{Error: "unsupported API version auth/v2"},
			httpStatus: http.StatusBadRequest,
		},
		{
			name: "BadKind",
			req: &v1beta1.TokenReview{
				TypeMeta:   v1.TypeMeta{APIVersion: authv1Beta1, Kind: "TokenRequest"},
				ObjectMeta: v1.ObjectMeta{CreationTimestamp: v1.Now()},
				Spec:       v1beta1.TokenReviewSpec{Token: "badToken"},
			},
			trStatus:   v1beta1.TokenReviewStatus{Error: "unsupported Kind TokenRequest"},
			httpStatus: http.StatusBadRequest,
		},
		{
			name: "MissingToken",
			req: &v1beta1.TokenReview{
				TypeMeta:   v1.TypeMeta{APIVersion: authv1Beta1, Kind: tokenReview},
				ObjectMeta: v1.ObjectMeta{CreationTimestamp: v1.Now()},
				Spec:       v1beta1.TokenReviewSpec{Token: ""},
			},
			trStatus:   v1beta1.TokenReviewStatus{Error: "missing token"},
			httpStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			a := &predictableAuthenticator{tt.err}

			w := httptest.NewRecorder()
			body, err := json.Marshal(tt.req)
			if err != nil {
				t.Fatalf("json.Marshal(%+#v): %v", tt.req, err)
			}
			Handler(a)(w, httptest.NewRequest("GET", "/", bytes.NewReader(body)))

			if w.Code != tt.httpStatus {
				t.Fatalf("w.Code: want %v, got %v", tt.httpStatus, w.Code)
			}

			rsp := &v1beta1.TokenReview{}
			if err := json.Unmarshal(w.Body.Bytes(), rsp); err != nil {
				t.Errorf("json.Unmarshal(%v, %s): %v", w.Body, rsp, err)
			}

			// Check request status specifically to avoid having to mock out
			// the metadata creation time.
			if diff := deep.Equal(tt.trStatus, rsp.Status); diff != nil {
				t.Errorf("want != got: %v", diff)
			}
		})
	}
}
