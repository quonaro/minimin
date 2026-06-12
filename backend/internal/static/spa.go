package static

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:web/*
var webFS embed.FS

// SPAHandler returns an http.Handler that serves the embedded SPA assets.
// It falls back to index.html for any missing file so client-side routing works.
// Returns nil when the SPA is not embedded (e.g. dev mode where web/ only
// contains .gitkeep).
func SPAHandler() http.Handler {
	fsys, err := fs.Sub(webFS, "web")
	if err != nil {
		return nil
	}

	if _, err := fs.Stat(fsys, "index.html"); err != nil {
		return nil
	}

	fileServer := http.FileServer(http.FS(fsys))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cleanPath := strings.TrimPrefix(r.URL.Path, "/")
		if cleanPath == "" {
			cleanPath = "index.html"
		}

		if _, err := fs.Stat(fsys, cleanPath); err != nil {
			r.URL.Path = "/"
		}

		fileServer.ServeHTTP(w, r)
	})
}
