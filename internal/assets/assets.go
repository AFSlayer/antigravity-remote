// Package assets serves the icons that Antigravity's bundle references but its
// language server does not provide: those paths fall through to the SPA shell, so a
// phone asking for a favicon receives HTML.
//
// The icon files are Antigravity's own, extracted from the desktop app and used
// only to identify Google's application in the browser tab and on the home
// screen. See DISCLAIMER.md.
package assets

import (
	"embed"
	"net/http"
	"strings"
	"time"
)

//go:embed files
var files embed.FS

var modTime = time.Now()

type asset struct {
	path        string
	contentType string
}

var routes = map[string]asset{
	"/favicon.ico":                       {"files/favicon.ico", "image/x-icon"},
	"/apple-touch-icon.png":              {"files/apple-touch-icon.png", "image/png"},
	"/apple-touch-icon-precomposed.png":  {"files/apple-touch-icon.png", "image/png"},
	"/assets/image/antigravity-logo.png": {"files/logo.png", "image/png"},
}

// IsPublicPath reports whether a path may be served without authentication. These
// are static icons, which carry nothing sensitive but must load on the login page.
func IsPublicPath(path string) bool {
	_, ok := routes[path]
	return ok
}

// Paths lists every route this package serves, for mounting on a mux.
func Paths() []string {
	out := make([]string, 0, len(routes))
	for p := range routes {
		out = append(out, p)
	}
	return out
}

// Handler serves the embedded icons.
func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a, ok := routes[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}

		data, err := files.ReadFile(a.path)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", a.contentType)
		w.Header().Set("Cache-Control", "public, max-age=86400")
		http.ServeContent(w, r, a.path, modTime, strings.NewReader(string(data)))
	})
}
