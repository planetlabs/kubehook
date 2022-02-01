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
	"time"

	"k8s.io/client-go/tools/clientcmd/api"
)


const (
	defaultClusterId = "radcluster"
)

type rsp struct {
	ClusterID   string `json:"cluster_id,omitempty"`
	MaxLifetime float64 `json:"max_lifetime,omitempty"`
}

// Handler returns an HTTP handler function that provides the client config.
func Handler(lifetime time.Duration, template *api.Config) http.HandlerFunc {
	data := rsp {
		ClusterID: defaultClusterId,
		MaxLifetime: lifetime.Hours(),
	}

	if template != nil {
		for name, _ := range template.Clusters {
			data.ClusterID = name
			break
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		write(w, data, http.StatusOK)
	}
}

func write(w http.ResponseWriter, data interface{}, httpStatus int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(data) // nolint: gosec
}
