package http

import (
	"net/http"
	"os"
	"path/filepath"
)

// spaFileServer serves the built React frontend out of dir. Requests for
// an actual file (JS/CSS/images) are served directly; anything else (the
// SPA's own routes, or just "/") falls back to index.html so client-side
// routing works, and so a single deep link still loads the app instead of
// 404ing.
//
// If dir doesn't exist — the normal case in local dev, where the frontend
// is served separately by Vite — every request just falls through to
// http.ServeFile's own file-not-found handling. It never panics on a
// missing directory, so this is always safe to wire up unconditionally.
func spaFileServer(dir string) http.Handler {
	fileServer := http.FileServer(http.Dir(dir))
	index := filepath.Join(dir, "index.html")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requested := filepath.Join(dir, filepath.Clean(r.URL.Path))
		if info, err := os.Stat(requested); err != nil || info.IsDir() {
			http.ServeFile(w, r, index)
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}
