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

package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-test/deep"
	"k8s.io/client-go/tools/clientcmd/api"
)


const (
	testCluster = "test"
)

var (
	testDuration, _ = time.ParseDuration("10h")
	testTemplate = &api.Config {
		Clusters: map[string]*api.Cluster {
			testCluster: &api.Cluster {
			},
		},
	}
)

func TestHandler(t *testing.T) {
	cases := map[string]struct {
		lifetime time.Duration
		template *api.Config
		rsp  		 *rsp
	}{
		"Default Cluster": {
			lifetime: testDuration,
			rsp:      &rsp{
				MaxLifetime: testDuration.Hours(),
				ClusterID:   defaultClusterId,
			},
		},
		"Template Cluster": {
			lifetime: testDuration,
			template: testTemplate,
			rsp:      &rsp{
				MaxLifetime: testDuration.Hours(),
				ClusterID:   testCluster,
			},
		},
	}
	for testName, tt := range cases {
		t.Run(testName, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			Handler(tt.lifetime, tt.template)(w, r)

			expectedStatus := http.StatusOK
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
