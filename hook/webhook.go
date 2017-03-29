package hook

import (
	"encoding/json"
	"net/http"

	"k8s.io/client-go/pkg/apis/authentication/v1beta1"

	"github.com/negz/kubehook/auth"
)

// StatusFailed is returned for failed authentication requests.
var StatusFailed v1beta1.TokenReviewStatus = v1beta1.TokenReviewStatus{Authenticated: false}

// Handler returns an HTTP handler function that handles an authentication
// webhook using the supplied Authenticator.
func Handler(a auth.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		tr := &v1beta1.TokenReview{}
		err := json.NewDecoder(r.Body).Decode(tr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		u, err := a.Authenticate(tr.Spec.Token)
		if err != nil {
			tr.Status = StatusFailed
			j, jerr := json.Marshal(tr)
			if jerr != nil {
				http.Error(w, jerr.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(j)
			return
		}

		tr.Status = userToStatus(u)
		j, err := json.Marshal(tr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(j)
	}
}

func userToStatus(u *auth.User) v1beta1.TokenReviewStatus {
	return v1beta1.TokenReviewStatus{
		Authenticated: true,
		User: v1beta1.UserInfo{
			Username: u.Username,
			UID:      u.UID,
			Groups:   u.Groups,
		},
	}
}
