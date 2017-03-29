package auth

// A User represents an authenticated user.
type User struct {
	Username string   // Username is the user's maybe-not-unique username.
	UID      string   // UID is a unique representation of this user.
	Groups   []string // Groups are the groups the user belongs to.
}

// An Authenticator authenticates a user based on a token.
type Authenticator interface {
	Authenticate(token string) (*User, error)
}
