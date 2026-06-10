package middleware

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"net/http"
)

// RequestLogger logs each HTTP request with method, path and status.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		slog.Info(fmt.Sprintf("%s %s %d", r.Method, r.URL.Path, rec.status))
	})
}

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.status = code
	rr.ResponseWriter.WriteHeader(code)
}

func (rr *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := rr.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, fmt.Errorf("response recorder: underlying ResponseWriter does not implement http.Hijacker")
}
