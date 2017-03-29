package auth

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/pkg/errors"
)

type noop struct {
	log    *zap.Logger
	groups []string
}

// A Option represents an argument to NewBackend
type Option func(*noop) error

// Logger allows the use of a custom Zap logger.
func Logger(l *zap.Logger) Option {
	return func(f *noop) error {
		f.log = l
		return nil
	}
}

// NewNoopAuthenticator returns an authenticator that authenticates any supplied
// token as the username passed to it as a token.
func NewNoopAuthenticator(groups []string, ao ...Option) (Authenticator, error) {
	l, err := zap.NewProduction()
	if err != nil {
		return nil, errors.Wrap(err, "cannot create default logger")
	}
	a := &noop{log: l, groups: groups}
	for _, o := range ao {
		if err := o(a); err != nil {
			return nil, errors.Wrap(err, "cannot apply noop authenticator option")
		}
	}
	a.log.Debug("granting to all requests", zap.Strings("groups", groups))
	return a, nil
}

func (n *noop) Authenticate(token string) (*User, error) {
	log := n.log.With(zap.String("token", token))
	if token == "" {
		log.Info("authentication", zap.Bool("success", false))
		return nil, errors.New("you must provide a token that is your desired username")
	}

	log.Info("authentication", zap.Bool("success", true))
	return &User{Username: token, UID: fmt.Sprintf("noop/%s", token), Groups: n.groups}, nil
}
