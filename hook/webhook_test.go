package hook

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/negz/kubehook/auth"

	"k8s.io/client-go/pkg/apis/authentication/v1beta1"
)

var webhookTests = []struct {
	req v1beta1.TokenReview
	rsp v1beta1.TokenReview
}{
	{
		req: v1beta1.TokenReview{Spec: v1beta1.TokenReviewSpec{Token: "token"}},
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
		req: v1beta1.TokenReview{Spec: v1beta1.TokenReviewSpec{Token: ""}},
		rsp: v1beta1.TokenReview{
			Spec:   v1beta1.TokenReviewSpec{Token: ""},
			Status: v1beta1.TokenReviewStatus{Authenticated: false},
		},
	},
}

func TestHandler(t *testing.T) {
	for _, tt := range webhookTests {
		a, err := auth.NewNoopAuthenticator(tt.rsp.Status.User.Groups)
		if err != nil {
			t.Errorf("auth.NewNoopAuthenticator(%v): %v", tt.rsp.Status.User.Groups, err)
			continue
		}

		w := httptest.NewRecorder()
		body, err := json.Marshal(tt.req)
		if err != nil {
			t.Errorf("json.Marshal(%+#v): %v", tt.req, err)
			continue
		}
		Handler(a)(w, httptest.NewRequest("GET", "/", bytes.NewReader(body)))
		expectedStatus := http.StatusOK
		if !tt.rsp.Status.Authenticated {
			expectedStatus = http.StatusUnauthorized
		}

		if w.Code != expectedStatus {
			t.Errorf("w.Code: want %v, got %v", expectedStatus, w.Code)
			continue
		}

		rsp := &v1beta1.TokenReview{}
		if err := json.Unmarshal(w.Body.Bytes(), rsp); err != nil {
			t.Errorf("json.Unmarshal(%v, %s): %v", w.Body, rsp, err)
			continue
		}

		if !reflect.DeepEqual(tt.rsp, *rsp) {
			t.Errorf("want:\n %+#v\n\n got:\n %+#v", tt.rsp, rsp)
		}
	}
}
