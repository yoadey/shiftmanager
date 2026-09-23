package static

import (
	"io/fs"
	"net/http/httptest"
	"strings"
	"testing"
)

// Regression test: without the "all:" prefix on the go:embed directive, Go
// silently drops any file or directory whose name starts with "." or "_"
// from the embedded FS. Vite names the chunk for src/screens/_demo.tsx
// "_demo-<hash>.js"; when it was missing, requests for it silently fell
// through to the SPA fallback (index.html served as text/html) instead of
// 404ing or serving the real file — which browsers reject for a
// <script type="module"> import. This asserts every underscore-prefixed
// asset actually shipped in dist/assets is served as itself, not as HTML.
func TestHandler_ServesUnderscorePrefixedAssets(t *testing.T) {
	root := FS()
	entries, err := fs.ReadDir(root, "assets")
	if err != nil {
		t.Fatalf("read assets dir: %v", err)
	}

	var underscored []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "_") {
			underscored = append(underscored, e.Name())
		}
	}
	if len(underscored) == 0 {
		t.Fatal("expected at least one underscore-prefixed asset in the committed dist/ placeholder (e.g. _demo-<hash>.js); none found")
	}

	h := Handler()
	for _, name := range underscored {
		req := httptest.NewRequest("GET", "/assets/"+name, nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != 200 {
			t.Errorf("%s: status = %d, want 200", name, rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); strings.Contains(ct, "text/html") {
			t.Errorf("%s: Content-Type = %q, want a JS/asset type (got the SPA fallback instead of the real file)", name, ct)
		}
		if strings.HasPrefix(rec.Body.String(), "<!DOCTYPE") {
			t.Errorf("%s: body is index.html (SPA fallback), the file was not actually embedded", name)
		}
	}
}
