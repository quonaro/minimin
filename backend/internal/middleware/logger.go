package middleware

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"orchestrator/internal/logger"
)

type responseRecorder struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (rr *responseRecorder) WriteHeader(code int) {
	if !rr.wrote {
		rr.status = code
		rr.wrote = true
		rr.ResponseWriter.WriteHeader(code)
	}
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	if !rr.wrote {
		rr.status = http.StatusOK
		rr.wrote = true
	}
	return rr.ResponseWriter.Write(b)
}

func (rr *responseRecorder) Flush() {
	if f, ok := rr.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (rr *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := rr.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("response recorder does not implement http.Hijacker")
	}
	return h.Hijack()
}

// RequestLogger logs each HTTP request with a compact colored format.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		msg := fmt.Sprintf("%s %s %s", r.Method, r.URL.Path, logger.ColorizeStatus(rec.status))
		slog.Info(msg)
	})
}
