package middleware

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"strings"
)

// responseWriter wraps http.ResponseWriter and optionally preserves http.Flusher.
type responseWriter struct {
	http.ResponseWriter
	flusher http.Flusher
}

func newResponseWriter(w http.ResponseWriter) http.ResponseWriter {
	rw := &responseWriter{ResponseWriter: w}
	if f, ok := w.(http.Flusher); ok {
		rw.flusher = f
	}
	return rw
}

func (w *responseWriter) Flush() {
	if w.flusher != nil {
		w.flusher.Flush()
	}
}

func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, fmt.Errorf("response writer: underlying ResponseWriter does not implement http.Hijacker")
}

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
		next(newResponseWriter(w), r)
	}
}
