// Package web serves the statically exported Next.js site (apps/web) from the API binary,
// so the site and /v1 share one origin. The Docker build copies apps/web/out into dist/.
package web

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed all:dist
var dist embed.FS

// Handler serves the site; it reports false when the binary was built without a site.
func Handler() (http.Handler, bool) {
	root, _ := fs.Sub(dist, "dist")
	if _, err := fs.Stat(root, "index.html"); err != nil {
		return nil, false
	}
	files := http.FileServer(http.FS(root))
	notFound, _ := fs.ReadFile(root, "404.html")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := path.Clean(r.URL.Path)
		if strings.HasPrefix(p, "/_next/static/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable") // content-hashed names
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		name := strings.TrimPrefix(p, "/")
		if name != "" && name != "." {
			if _, err := fs.Stat(root, name); err != nil {
				if _, err := fs.Stat(root, name+"/index.html"); err != nil {
					w.Header().Set("Content-Type", "text/html; charset=utf-8")
					w.WriteHeader(http.StatusNotFound)
					w.Write(notFound)
					return
				}
			}
		}
		files.ServeHTTP(w, r)
	}), true
}
