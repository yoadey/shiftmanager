package static

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

// distFS embeds the built single-page-application assets. The frontend build
// (e.g. `npm run build` in ./frontend) must output its production bundle into
// internal/infrastructure/static/dist before the Go binary is compiled.
//
//go:embed dist
var distFS embed.FS

// FS returns the embedded frontend filesystem rooted at the dist directory.
func FS() fs.FS {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		// dist is always present (it contains at least index.html), so this
		// should never happen. Fall back to the raw embed FS to stay safe.
		return distFS
	}
	return sub
}

// Handler returns an http.Handler that serves the embedded SPA. Requests that
// map to an existing static asset are served directly; everything else falls
// back to index.html so client-side routing works on deep links and reloads.
func Handler() http.Handler {
	root := FS()
	fileServer := http.FileServer(http.FS(root))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upath := strings.TrimPrefix(r.URL.Path, "/")
		if upath == "" {
			upath = "index.html"
		}

		// Serve the asset directly when it exists.
		if f, err := root.Open(upath); err == nil {
			_ = f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// SPA fallback: serve index.html for unknown (non-asset) paths.
		serveIndex(w, r, root)
	})
}

// serveIndex writes the embedded index.html with a 200 status.
func serveIndex(w http.ResponseWriter, r *http.Request, root fs.FS) {
	data, err := fs.ReadFile(root, "index.html")
	if err != nil {
		http.Error(w, "frontend not built", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
