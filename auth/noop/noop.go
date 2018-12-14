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

package noop

import (
	"fmt"
	"time"

	"github.com/negz/kubehook/auth"

	"github.com/pkg/errors"
	"go.uber.org/zap"
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

// NewManager returns a no-op token manager.
func NewManager(groups []string, ao ...Option) (auth.Manager, error) {
	l, err := zap.NewProduction()
	if err != nil {
		return nil, errors.Wrap(err, "cannot create default logger")
	}
	a := &noop{log: l, groups: groups}
	for _, o := range ao {
		if err := o(a); err != nil {
			return nil, errors.Wrap(err, "cannot apply noop manager option")
		}
	}
	a.log.Debug("granting to all requests", zap.Strings("groups", groups))
	return a, nil
}

func (n *noop) Authenticate(token string) (*auth.User, error) {
	log := n.log.With(zap.String("token", token))
	if token == "" {
		log.Info("authentication", zap.Bool("success", false))
		return nil, errors.New("you must provide a token that is your desired username")
	}

	log.Info("authentication", zap.Bool("success", true))
	return &auth.User{Username: token, UID: fmt.Sprintf("noop/%s", token), Groups: n.groups}, nil
}

func (n *noop) Generate(u *auth.User, _ time.Duration) (string, error) {
	n.log.Info("generate", zap.String("uid", u.UID), zap.String("token", u.Username))
	return u.Username, nil
}
