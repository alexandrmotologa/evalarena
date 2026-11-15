package server

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:dist
var distFS embed.FS

// FileServer returns an http.Handler serving the embedded React 19 studio SPA.
func FileServer() http.Handler {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Embedded frontend assets not found", http.StatusInternalServerError)
		})
	}

	fsHandler := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		// Check if requested static file exists in embedded dist
		f, err := sub.Open(path)
		if err == nil {
			_ = f.Close()
			fsHandler.ServeHTTP(w, r)
			return
		}

		// Fallback to index.html for client-side routing
		indexF, err := sub.Open("index.html")
		if err == nil {
			_ = indexF.Close()
			r.URL.Path = "/"
			fsHandler.ServeHTTP(w, r)
			return
		}

		http.NotFound(w, r)
	})
}
