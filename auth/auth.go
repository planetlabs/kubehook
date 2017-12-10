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
