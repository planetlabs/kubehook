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

package jwt

import (
	"reflect"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-test/deep"
	"github.com/planetlabs/kubehook/auth"
)

var secret = []byte("secret!")

var tenMinsAgo = time.Now().UTC().Add(-10 * time.Minute).Unix()
var tenMinsFromNow = time.Now().UTC().Add(10 * time.Minute).Unix()

func tokenWithMethod(m jwt.SigningMethod, secret []byte, audience, username string, nbf, exp int64) string {
	c := &claims{
		StandardClaims: jwt.StandardClaims{
			Audience:  audience,
			Subject:   username,
			NotBefore: nbf,
			ExpiresAt: exp,
		},
	}

	t := jwt.NewWithClaims(m, c)
	ss, _ := t.SignedString(secret)
	return ss
}

func token(secret []byte, audience, username string, nbf, exp int64) string {
	return tokenWithMethod(jwt.SigningMethodHS256, secret, audience, username, nbf, exp)
}

func TestAuthenticate(t *testing.T) {
	cases := []struct {
		name    string
		secret  []byte
		opts    []Option
		token   string
		want    *auth.User
		wantErr bool
	}{
		{
			name:    "ValidToken",
			secret:  secret,
			token:   token(secret, DefaultAudience, "negz", tenMinsAgo, tenMinsFromNow),
			want:    &auth.User{Username: "negz", UID: "github.com/planetlabs/kubehook/negz"},
			wantErr: false,
		},
		{
			name:    "InvalidSignature",
			secret:  secret,
			token:   token([]byte("notsecret"), DefaultAudience, "negz", tenMinsAgo, tenMinsFromNow),
			wantErr: true,
		},
		{
			name:    "InvalidAudience",
			opts:    []Option{Audience("audience!")},
			secret:  secret,
			token:   token(secret, DefaultAudience, "negz", tenMinsAgo, tenMinsFromNow),
			wantErr: true,
		},
		{
			name:    "NotYetValid",
			secret:  secret,
			token:   token(secret, DefaultAudience, "negz", tenMinsFromNow, tenMinsFromNow),
			wantErr: true,
		},
		{
			name:    "Expired",
			secret:  secret,
			token:   token(secret, DefaultAudience, "negz", tenMinsAgo, tenMinsAgo),
			wantErr: true,
		},
		{
			name:    "NotHMAC",
			secret:  secret,
			token:   tokenWithMethod(jwt.SigningMethodNone, secret, DefaultAudience, "negz", tenMinsAgo, tenMinsFromNow),
			wantErr: true,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			m, _ := NewManager(tt.secret, tt.opts...)
			got, err := m.Authenticate(tt.token)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Fatalf("m.Authenticate(...): %v", err)
			}
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("m.Authenticate(...): got != want: %v", diff)
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	cases := []struct {
		name     string
		secret   []byte
		lifetime time.Duration
		opts     []Option
		user     *auth.User
		wantErr  bool
	}{
		{
			name:     "ValidToken",
			secret:   secret,
			user:     &auth.User{Username: "negz", UID: "github.com/planetlabs/kubehook/negz"},
			lifetime: DefaultMaxLifetime,
			wantErr:  false,
		},
		{
			name:     "LifetimeTooLong",
			secret:   secret,
			opts:     []Option{MaxLifetime(DefaultMaxLifetime - 1*time.Hour)},
			user:     &auth.User{Username: "negz", UID: "github.com/planetlabs/kubehook/negz"},
			lifetime: DefaultMaxLifetime,
			wantErr:  true,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			m, _ := NewManager(tt.secret, tt.opts...)
			token, err := m.Generate(tt.user, tt.lifetime)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Fatalf("m.Generate(...): %v", err)
			}

			got, err := m.Authenticate(token)
			if err != nil {
				t.Fatalf("m.Authenticate(...): %v", err)
			}

			if !reflect.DeepEqual(got, tt.user) {
				t.Errorf("m.Generate(...): got %v, want %v", got, tt.user)
			}
		})
	}
}
