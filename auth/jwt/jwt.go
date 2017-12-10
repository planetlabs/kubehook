package jwt

import (
	"github.com/negz/kubehook/auth"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// DefaultAudience for JWT token verification.
const DefaultAudience = "github.com/negz/kubehook"

type jwta struct {
	log      *zap.Logger
	secret   []byte
	audience string
}

// A Option represents an argument to NewBackend
type Option func(*jwta) error

// Logger allows the use of a custom Zap logger.
func Logger(l *zap.Logger) Option {
	return func(f *jwta) error {
		f.log = l
		return nil
	}
}

// Audience required to be set in valid JWTs.
func Audience(a string) Option {
	return func(f *jwta) error {
		f.audience = a
		return nil
	}
}

// NewAuthenticator returns an authenticator that authenticates JWTs.
func NewAuthenticator(secret []byte, ao ...Option) (auth.Authenticator, error) {
	l, err := zap.NewProduction()
	if err != nil {
		return nil, errors.Wrap(err, "cannot create default logger")
	}
	a := &jwta{log: l, secret: secret, audience: DefaultAudience}
	for _, o := range ao {
		if err := o(a); err != nil {
			return nil, errors.Wrap(err, "cannot apply JWT authenticator option")
		}
	}
	return a, nil
}

type claims struct {
	UID    string   `json:"uid,omitempty"`
	Groups []string `json:"grp,omitempty"`
	jwt.StandardClaims
}

func isHMACSigned(t *jwt.Token) bool {
	_, ok := t.Method.(*jwt.SigningMethodHMAC)
	return ok
}

func (a *jwta) Authenticate(token string) (*auth.User, error) {
	log := a.log.With(zap.String("jwt", token))

	t, err := jwt.ParseWithClaims(token, &claims{}, func(t *jwt.Token) (interface{}, error) {
		if !isHMACSigned(t) {
			return nil, errors.Errorf("token must be HMAC signed JWT")
		}
		return a.secret, nil
	})
	if err != nil {
		log.Info("auth", zap.Bool("success", false))
		return nil, errors.Wrap(err, "invalid JWT token")
	}

	c, ok := t.Claims.(*claims)
	if !ok {
		log.Info("auth", zap.Bool("success", false))
		return nil, errors.New("cannot parse JWT claims")
	}
	if c.Audience != a.audience {
		log.Info("auth", zap.Bool("success", false))
		return nil, errors.Errorf("invalid JWT audience %s - audience %s is required", c.Audience, a.audience)
	}

	log.Info("auth", zap.Bool("success", true))
	return &auth.User{Username: c.Subject, UID: c.UID, Groups: c.Groups}, nil
}
