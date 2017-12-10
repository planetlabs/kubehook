package jwt

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/negz/kubehook/auth"
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
		UID: fmt.Sprintf("kubehook/%s", username),
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
			want:    &auth.User{Username: "negz", UID: "kubehook/negz"},
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
			a, _ := NewAuthenticator(tt.secret, tt.opts...)
			got, err := a.Authenticate(tt.token)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Fatalf("a.Authenticate(...): %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("a.Authenticate(...) got %v, want %v", got, tt.want)
			}
		})
	}
}
