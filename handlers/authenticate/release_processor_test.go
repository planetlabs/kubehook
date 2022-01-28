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

package authenticate

import (
	"strings"
	"testing"
)

func TestReleaseProcessorExtractToken(t *testing.T) {
	sut := newReleaseProcessor(mockTimeProvider)
	cases := map[string]struct {
		input      string
		wantResult string
		wantError  bool
	}{
		"success": {
			input: `{
  "kind": "TokenReview",
  "apiVersion": "authentication.k8s.io/v1",
  "spec": {
    "token": "release:password"
  }
}`,
			wantResult: "release:password",
		},
		"wrong DTO": {
			input: `{
  "error": "marshal failure"
}`,
			wantError: true,
		},
		"wrong kind": {
			input: `{
  "kind": "NotATokenReview",
  "apiVersion": "authentication.k8s.io/v1",
  "spec": {
    "token": "release:password"
  }
}`,
			wantError: true,
		},
		"wrong API version": {
			input: `{
  "kind": "TokenReview",
  "apiVersion": "authentication.k8s.io/v1beta1",
  "spec": {
    "token": "release:password"
  }
}`,
			wantError: true,
		},
		"empty token": {
			input: `{
  "kind": "TokenReview",
  "apiVersion": "authentication.k8s.io/v1",
  "spec": {
    "token": ""
  }
}`,
			wantError: true,
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			input := strings.NewReader(tt.input)
			token, err := sut.ExtractToken(input)

			if err != nil {
				if tt.wantError {
					return
				}

				t.Fatalf("p.ExtractToken(...): %v", err)
			}

			if token != tt.wantResult {
				t.Errorf("got = %v; want = %v", token, tt.wantResult)
			}
		})
	}
}
