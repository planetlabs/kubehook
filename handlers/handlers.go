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

package handlers

import (
	"io"
	"net/http"
	"time"
)

// Default headers used to determine the currently authenticated user and their
// groups.
const (
	DefaultUserHeader           = "X-Forwarded-User"
	DefaultGroupHeader          = "X-Forwarded-Groups"
	DefaultGroupHeaderDelimiter = ";"
)

// AuthHeaders from which the authenticated user and their groups are extracted.
type AuthHeaders struct {
	User  string // The authenticated user.
	Group string // The authenticated user's groups.

	GroupDelimiter string // The delimiter of the list of groups.
}

// Run the provided function in a goroutine.
func Run(fn func()) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		go fn()
		w.WriteHeader(http.StatusOK)
		r.Body.Close()
	}
}

// Ping always returns HTTP 200.
func Ping() http.HandlerFunc {
	// TODO(negz): Check kubehook health?
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		r.Body.Close()
	}
}

// Content serves the supplied content at the supplied path.
func Content(c io.ReadSeeker, filename string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, filename, time.Unix(0, 0), c)
		r.Body.Close()
	}
}

// NotImplemented always returns HTTP 501
func NotImplemented() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		r.Body.Close()
	}
}
