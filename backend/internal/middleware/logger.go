package middleware

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
)

type responseRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
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
	n, err := rr.ResponseWriter.Write(b)
	rr.bytes += n
	return n, err
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

// RequestLogger logs each HTTP request with method, path, status and duration.
// Statuses >= 400 are logged as errors so they stand out in the logs.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		if rec.status >= 400 {
			slog.Error("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"bytes", rec.bytes,
				"duration", time.Since(start),
			)
		} else {
			slog.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"bytes", rec.bytes,
				"duration", time.Since(start),
			)
		}
	})
}
