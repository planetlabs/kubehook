package hook

import (
	"encoding/json"
	"net/http"

	"k8s.io/api/authentication/v1beta1"

	"github.com/negz/kubehook/auth"
)

// Handler returns an HTTP handler function that handles an authentication
// webhook using the supplied Authenticator.
func Handler(a auth.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		req := &v1beta1.TokenReview{}
		err := json.NewDecoder(r.Body).Decode(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		rsp := v1beta1.TokenReview{}
		u, err := a.Authenticate(req.Spec.Token)
		if err != nil {
			rsp.Status = v1beta1.TokenReviewStatus{Authenticated: false}
			j, jerr := json.Marshal(rsp)
			if jerr != nil {
				http.Error(w, jerr.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusForbidden)
			w.Write(j)
			return
		}

		rsp.Status = userToStatus(u)
		j, err := json.Marshal(rsp)
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
