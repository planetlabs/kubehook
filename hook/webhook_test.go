package hook

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/negz/kubehook/auth/noop"

	"k8s.io/api/authentication/v1beta1"
)

func TestHandler(t *testing.T) {
	cases := []struct {
		name string
		req  v1beta1.TokenReview
		rsp  v1beta1.TokenReview
	}{
		{
			name: "Success",
			req:  v1beta1.TokenReview{Spec: v1beta1.TokenReviewSpec{Token: "token"}},
			rsp: v1beta1.TokenReview{
				Spec: v1beta1.TokenReviewSpec{Token: "token"},
				Status: v1beta1.TokenReviewStatus{
					Authenticated: true,
					User: v1beta1.UserInfo{
						Username: "token",
						UID:      "noop/token",
						Groups:   []string{"engineering"},
					},
				},
			},
		},
		{
			name: "Failure",
			req:  v1beta1.TokenReview{Spec: v1beta1.TokenReviewSpec{Token: ""}},
			rsp: v1beta1.TokenReview{
				Spec:   v1beta1.TokenReviewSpec{Token: ""},
				Status: v1beta1.TokenReviewStatus{Authenticated: false},
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			a, err := noop.NewAuthenticator(tt.rsp.Status.User.Groups)
			if err != nil {
				t.Fatalf("auth.NewNoopAuthenticator(%v): %v", tt.rsp.Status.User.Groups, err)
			}

			w := httptest.NewRecorder()
			body, err := json.Marshal(tt.req)
			if err != nil {
				t.Fatalf("json.Marshal(%+#v): %v", tt.req, err)
			}
			Handler(a)(w, httptest.NewRequest("GET", "/", bytes.NewReader(body)))
			expectedStatus := http.StatusOK
			if !tt.rsp.Status.Authenticated {
				expectedStatus = http.StatusForbidden
			}

			if w.Code != expectedStatus {
				t.Fatalf("w.Code: want %v, got %v", expectedStatus, w.Code)
			}

			rsp := &v1beta1.TokenReview{}
			if err := json.Unmarshal(w.Body.Bytes(), rsp); err != nil {
				t.Errorf("json.Unmarshal(%v, %s): %v", w.Body, rsp, err)
			}

			if !reflect.DeepEqual(tt.rsp, *rsp) {
				t.Errorf("want:\n %+#v\n\n got:\n %+#v", tt.rsp, rsp)
			}
		})
	}
}
