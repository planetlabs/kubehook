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
	"fmt"
	"time"

	"github.com/planetlabs/kubehook/auth"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Defaults for JSON Web Tokens.
const (
	DefaultAudience    = "github.com/planetlabs/kubehook"
	DefaultMaxLifetime = 7 * 24 * time.Hour
)

type jwtm struct {
	log         *zap.Logger
	secret      []byte
	audience    string
	maxLifetime time.Duration
}

// An Option represents an optional argument to NewBackend
type Option func(*jwtm) error

// Logger allows the use of a custom Zap logger.
func Logger(l *zap.Logger) Option {
	return func(f *jwtm) error {
		f.log = l
		return nil
	}
}

// Audience required to be set in valid JWTs.
func Audience(a string) Option {
	return func(f *jwtm) error {
		f.audience = a
		return nil
	}
}

// MaxLifetime is the maximum allowed expiry time for generated tokens.
func MaxLifetime(d time.Duration) Option {
	return func(f *jwtm) error {
		f.maxLifetime = d
		return nil
	}
}

// NewManager generates and authenticates JSON Web Tokens (JWTs).
func NewManager(secret []byte, mo ...Option) (auth.Manager, error) {
	l, err := zap.NewProduction()
	if err != nil {
		return nil, errors.Wrap(err, "cannot create default logger")
	}
	m := &jwtm{log: l, secret: secret, audience: DefaultAudience, maxLifetime: DefaultMaxLifetime}
	for _, o := range mo {
		if err := o(m); err != nil {
			return nil, errors.Wrap(err, "cannot apply JWT manager option")
		}
	}
	return m, nil
}

type claims struct {
	Groups []string `json:"grp,omitempty"`
	jwt.StandardClaims
}

func (c *claims) UID() string {
	return fmt.Sprintf("%s/%s", c.Audience, c.Subject)
}

func isHMACSigned(t *jwt.Token) bool {
	_, ok := t.Method.(*jwt.SigningMethodHMAC)
	return ok
}

func (m *jwtm) Authenticate(token string) (*auth.User, error) {
	log := m.log.With(zap.String("jwt", token))

	t, err := jwt.ParseWithClaims(token, &claims{}, func(t *jwt.Token) (interface{}, error) {
		if !isHMACSigned(t) {
			return nil, errors.Errorf("token must be HMAC signed JWT")
		}
		return m.secret, nil
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
	if c.Audience != m.audience {
		log.Info("auth", zap.Bool("success", false))
		return nil, errors.Errorf("invalid JWT audience %s - audience %s is required", c.Audience, m.audience)
	}

	log.Info("auth", zap.Bool("success", true))
	return &auth.User{Username: c.Subject, UID: c.UID(), Groups: c.Groups}, nil
}

func (m *jwtm) Generate(u *auth.User, lifetime time.Duration) (string, error) {
	log := m.log.With(
		zap.String("user", u.Username),
		zap.String("uid", u.UID),
		zap.Strings("groups", u.Groups),
		zap.Duration("lifetime", lifetime))

	if lifetime > m.maxLifetime {
		log.Info("generate", zap.Bool("success", false))
		return "", errors.Errorf("requested JWT lifetime %s is greater than maximum allowed lifetime %s", lifetime, m.maxLifetime)
	}

	c := &claims{
		StandardClaims: jwt.StandardClaims{
			Audience:  m.audience,
			Subject:   u.Username,
			NotBefore: time.Now().UTC().Unix(),
			ExpiresAt: time.Now().UTC().Add(lifetime).Unix(),
		},
		Groups: u.Groups,
	}

	ss, err := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(m.secret)
	if err != nil {
		log.Info("generate", zap.Bool("success", false))
		return "", errors.Wrap(err, "cannot generate JWT")
	}
	log.Info("generate", zap.Bool("success", true))
	return ss, nil
}
