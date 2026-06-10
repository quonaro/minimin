package middleware

import (
	"net/http"
	"strings"
)

// WithAuth returns a handler that validates the static API key before calling next.
func WithAuth(apiKey string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := ""
		if auth := r.Header.Get("Authorization"); auth != "" {
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				token = parts[1]
			}
		} else if c, err := r.Cookie("auth_token"); err == nil {
			token = c.Value
		} else {
			token = r.URL.Query().Get("token")
		}

		if token != apiKey {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}
		next(w, r)
	}
}
