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
	"github.com/negz/kubehook/handlers"
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
			head: map[string]string{handlers.DefaultUserHeader: user},
			req:  &req{Lifetime: 10 * lifetime.Minute},
			rsp:  &rsp{Token: user},
		},
		{
			name: "MissingUsernameHeader",
			head: map[string]string{"some-header": "value"},
			req:  &req{Lifetime: 10 * lifetime.Minute},
			rsp:  &rsp{Error: fmt.Sprintf("cannot extract username from header %s", handlers.DefaultUserHeader)},
		},
		{
			name: "MissingUsernameHeaderValue",
			head: map[string]string{handlers.DefaultUserHeader: ""},
			req:  &req{Lifetime: 10 * lifetime.Minute},
			rsp:  &rsp{Error: fmt.Sprintf("cannot extract username from header %s", handlers.DefaultUserHeader)},
		},
		{
			name: "MissingLifetime",
			head: map[string]string{handlers.DefaultUserHeader: user},
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

			h := handlers.AuthHeaders{
				User:           handlers.DefaultUserHeader,
				Group:          handlers.DefaultGroupHeader,
				GroupDelimiter: handlers.DefaultGroupHeaderDelimiter,
			}
			Handler(m, h)(w, r)

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
