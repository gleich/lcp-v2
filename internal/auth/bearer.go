package auth

import (
	"fmt"
	"net/http"

	"pkg.mattglei.ch/lcp-v2/internal/secrets"
)

func IsAuthorized(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %s", secrets.SECRETS.ValidToken) {
		w.WriteHeader(http.StatusUnauthorized)
		return false
	}
	return true
}
