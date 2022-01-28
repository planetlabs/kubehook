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
	"encoding/json"
	"io"

	"github.com/pkg/errors"
	"k8s.io/api/authentication/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/planetlabs/kubehook/auth"
)

var beta1APIVersion = v1beta1.SchemeGroupVersion.String()

type beta1Processor struct {
	APIVersion string
	Now        timeProvider
}

func (p *beta1Processor) ExtractToken(b io.Reader) (string, error) {
	req := &v1beta1.TokenReview{}
	err := json.NewDecoder(b).Decode(req)

	switch {
	case err != nil:
		return "", errors.Wrap(err, "cannot parse token request")
	case req.APIVersion != p.APIVersion:
		return "", errors.Errorf("unsupported API version %s", req.APIVersion)
	case req.Kind != tokenReview:
		return "", errors.Errorf("unsupported Kind %s", req.Kind)
	case req.Spec.Token == "":
		return "", errors.New("missing token")
	}

	return req.Spec.Token, nil
}

func (p *beta1Processor) CreateErrorStatus(err error) interface{} {
	review := p.newTokenReview()

	review.Status = v1beta1.TokenReviewStatus{Error: err.Error()}

	return review
}

func (p *beta1Processor) CreateReviewStatus(u *auth.User) interface{} {
	review := p.newTokenReview()

	review.Status = v1beta1.TokenReviewStatus{
		Authenticated: true,
		User: v1beta1.UserInfo{
			Username: u.Username,
			UID:      u.UID,
			Groups:   u.Groups,
		},
	}

	return review
}

func (p *beta1Processor) newTokenReview() v1beta1.TokenReview {
	return v1beta1.TokenReview{
		TypeMeta: v1.TypeMeta{
			APIVersion: p.APIVersion,
			Kind:       tokenReview,
		},
		ObjectMeta: v1.ObjectMeta{CreationTimestamp: p.Now()},
	}
}

func newBeta1Processor(now timeProvider) *beta1Processor {
	result := &beta1Processor{
		APIVersion: beta1APIVersion,
		Now:        now,
	}

	return result
}
