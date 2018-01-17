package generate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-test/deep"
	"github.com/negz/kubehook/auth/noop"
	"github.com/negz/kubehook/lifetime"
)

const user = "user"

var noGroups = []string{}

func TestHandler(t *testing.T) {
	cases := []struct {
		name string
		head map[string]string
		req  *req
		rsp  *rsp
	}{
		{
			name: "Success",
			head: map[string]string{DefaultUserHeader: user},
			req:  &req{Lifetime: 10 * lifetime.Minute},
			rsp:  &rsp{Token: user},
		},
		{
			name: "MissingUsernameHeader",
			head: map[string]string{"some-header": "value"},
			req:  &req{Lifetime: 10 * lifetime.Minute},
			rsp:  &rsp{Error: fmt.Sprintf("cannot extract username from header %s", DefaultUserHeader)},
		},
		{
			name: "MissingUsernameHeaderValue",
			head: map[string]string{DefaultUserHeader: ""},
			req:  &req{Lifetime: 10 * lifetime.Minute},
			rsp:  &rsp{Error: fmt.Sprintf("cannot extract username from header %s", DefaultUserHeader)},
		},
		{
			name: "MissingLifetime",
			head: map[string]string{DefaultUserHeader: user},
			req:  &req{},
			rsp:  &rsp{Error: "must specify desired token lifetime"},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			m, err := noop.NewManager(noGroups)
			if err != nil {
				t.Fatalf("auth.NewNoopAuthenticator(%v): %v", noGroups, err)
			}

			w := httptest.NewRecorder()
			body, err := json.Marshal(tt.req)
			if err != nil {
				t.Fatalf("json.Marshal(%+#v): %v", tt.req, err)
			}
			r := httptest.NewRequest("POST", "/", bytes.NewReader(body))
			for k, v := range tt.head {
				r.Header.Set(k, v)
			}

			Handler(m, DefaultUserHeader)(w, r)

			expectedStatus := http.StatusOK
			if tt.rsp.Error != "" {
				expectedStatus = http.StatusBadRequest
			}
			if w.Code != expectedStatus {
				t.Errorf("w.Code: want %v, got %v", expectedStatus, w.Code)
			}

			rsp := &rsp{}
			if err := json.Unmarshal(w.Body.Bytes(), rsp); err != nil {
				t.Fatalf("json.Unmarshal(%v, %s): %v", w.Body, rsp, err)
			}

			if diff := deep.Equal(tt.rsp, rsp); diff != nil {
				t.Errorf("want != got: %v", diff)
			}
		})
	}
}
