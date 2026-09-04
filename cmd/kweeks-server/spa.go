package main

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// spaHandler serves the built React app (web/dist) and falls back to
// index.html for client-side routes so deep links like /instructor/dashboard
// work on refresh. API + websocket paths are registered on the mux BEFORE
// this handler and therefore take precedence.
type spaHandler struct {
	root    string
	files   http.Handler
	index   []byte
	indexCT string
}

func newSPAHandler(webRoot string) (http.Handler, error) {
	abs, err := filepath.Abs(webRoot)
	if err != nil {
		return nil, err
	}
	indexPath := filepath.Join(abs, "index.html")
	index, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, err
	}
	return &spaHandler{
		root:    abs,
		files:   http.FileServer(http.Dir(abs)),
		index:   index,
		indexCT: "text/html; charset=utf-8",
	}, nil
}

func (h *spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clean := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if clean == "" {
		clean = "index.html"
	}
	filePath := filepath.Join(h.root, filepath.FromSlash(clean))
	info, err := os.Stat(filePath)
	if err == nil && !info.IsDir() {
		// Real asset (JS/CSS/images): let FileServer serve it with the right
		// content type. FileServer never redirects here because the file exists.
		h.files.ServeHTTP(w, r)
		return
	}
	if strings.Contains(clean, ".") && !strings.HasSuffix(clean, ".html") {
		// A dotted path that is not a real asset -> 404, never the SPA shell.
		http.NotFound(w, r)
		return
	}
	// SPA route: serve index.html bytes directly (no FileServer, no redirects).
	w.Header().Set("Content-Type", h.indexCT)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(h.index)
}
