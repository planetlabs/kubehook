package kubecfg

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-test/deep"

	"github.com/negz/kubehook/auth/noop"
	"github.com/negz/kubehook/lifetime"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

const user = "user"

var noGroups = []string{}

func TestHandler(t *testing.T) {
	cases := []struct {
		name     string
		head     map[string]string
		req      *req
		template *api.Config
		status   int
		want     api.Config
	}{
		{
			name: "Success",
			head: map[string]string{DefaultUserHeader: user},
			req:  &req{Lifetime: 10 * lifetime.Minute},
			template: &api.Config{
				Clusters: map[string]*api.Cluster{
					"a": &api.Cluster{Server: "https://example.org", CertificateAuthorityData: []byte("PAM")},
					"b": &api.Cluster{Server: "https://example.net", CertificateAuthorityData: []byte("PAM")},
				},
			},
			status: http.StatusOK,
			want: api.Config{
				Clusters: map[string]*api.Cluster{
					"a": &api.Cluster{Server: "https://example.org", CertificateAuthorityData: []byte("PAM")},
					"b": &api.Cluster{Server: "https://example.net", CertificateAuthorityData: []byte("PAM")},
				},
				Contexts: map[string]*api.Context{
					"a": &api.Context{AuthInfo: templateUser, Cluster: "a"},
					"b": &api.Context{AuthInfo: templateUser, Cluster: "b"},
				},
				AuthInfos: map[string]*api.AuthInfo{templateUser: &api.AuthInfo{Token: user}},
			},
		},
		{
			name:     "MissingUsernameHeader",
			head:     map[string]string{"some-header": "value"},
			req:      &req{Lifetime: 10 * lifetime.Minute},
			template: &api.Config{},
			status:   http.StatusBadRequest,
		},
		{
			name:     "MissingUsernameHeaderValue",
			head:     map[string]string{DefaultUserHeader: ""},
			req:      &req{Lifetime: 10 * lifetime.Minute},
			template: &api.Config{},
			status:   http.StatusBadRequest,
		},
		{
			name:     "MissingLifetime",
			head:     map[string]string{DefaultUserHeader: user},
			req:      &req{},
			template: &api.Config{},
			status:   http.StatusBadRequest,
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

			Handler(m, DefaultUserHeader, tt.template)(w, r)

			if w.Code != tt.status {
				t.Errorf("w.Code: want %v, got %v", tt.status, w.Code)
			}

			if w.Code != http.StatusOK {
				return
			}

			// Test output would be clearer if we diffed these as structs rather
			// than bytes. Unfortunately clientcmd.Load() helpfully allocates
			// memory for all the maps in the config object, making it
			// cumbersome to write a matching config object inline.
			want, _ := clientcmd.Write(tt.want)
			if diff := deep.Equal(string(want), string(w.Body.Bytes())); diff != nil {
				t.Errorf("want != got: %v", diff)
			}
		})
	}
}
