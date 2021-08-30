package api

import (
	"com.t-systems-mms.cwa/core/security"
	"net/http"
)

func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !security.HasRole(r.Context(), role) {
				WriteError(w, r, security.ErrForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
