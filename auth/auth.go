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

package auth

import (
	"time"
)

// A User represents an authenticated user.
type User struct {
	Username string   // Username is the user's maybe-not-unique username.
	UID      string   // UID is a unique representation of this user.
	Groups   []string // Groups are the groups the user belongs to.
}

// A Generator generates a token for the given user.
type Generator interface {
	Generate(u *User, lifetime time.Duration) (token string, err error)
}

// An Authenticator authenticates a user based on a token.
type Authenticator interface {
	Authenticate(token string) (*User, error)
}

// A Manager both generates and authenticates user tokens.
type Manager interface {
	Generator
	Authenticator
}
