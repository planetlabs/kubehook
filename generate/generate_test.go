package generate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/negz/kubehook/auth/noop"
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
			req:  &req{Lifetime: 10 * time.Minute},
			rsp:  &rsp{Token: user},
		},
		{
			name: "MissingUsernameHeader",
			head: map[string]string{"some-header": "value"},
			req:  &req{Lifetime: 10 * time.Minute},
			rsp:  &rsp{Error: fmt.Sprintf("cannot extract username from header %s", DefaultUserHeader)},
		},
		{
			name: "MissingUsernameHeaderValue",
			head: map[string]string{DefaultUserHeader: ""},
			req:  &req{Lifetime: 10 * time.Minute},
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
				t.Fatalf("w.Code: want %v, got %v", expectedStatus, w.Code)
			}

			rsp := &rsp{}
			if err := json.Unmarshal(w.Body.Bytes(), rsp); err != nil {
				t.Fatalf("json.Unmarshal(%v, %s): %v", w.Body, rsp, err)
			}

			if !reflect.DeepEqual(tt.rsp, rsp) {
				t.Errorf("\nwant: %+#v\n got: %+#v", tt.rsp, rsp)
			}
		})
	}
}
