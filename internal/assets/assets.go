// Package assets serves the icons and web manifest that Antigravity's bundle
// references but its language server does not provide: those paths fall through
// to the SPA shell, so a phone asking for a favicon receives HTML.
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
	"/__agy/icon-192.png":                {"files/icon-192.png", "image/png"},
	"/__agy/icon-512.png":                {"files/icon-512.png", "image/png"},
}

// IsPublicPath reports whether a path may be served without authentication.
// These are static icons and the manifest, which carry nothing sensitive but must
// load on the login page and the home screen.
func IsPublicPath(path string) bool {
	_, ok := routes[path]
	return ok || path == ManifestPath
}

// Paths lists every route this package serves, for mounting on a mux.
func Paths() []string {
	out := make([]string, 0, len(routes)+1)
	for p := range routes {
		out = append(out, p)
	}
	out = append(out, ManifestPath)
	return out
}

// ManifestPath is the web app manifest that makes Add to Home Screen open
// fullscreen.
const ManifestPath = "/__agy/manifest.webmanifest"

const manifestJSON = `{
  "name": "Antigravity",
  "short_name": "Antigravity",
  "description": "Google Antigravity, on your phone.",
  "start_url": "/",
  "scope": "/",
  "display": "standalone",
  "orientation": "portrait",
  "background_color": "#1e1e1e",
  "theme_color": "#1e1e1e",
  "icons": [
    { "src": "/__agy/icon-192.png", "sizes": "192x192", "type": "image/png", "purpose": "any" },
    { "src": "/__agy/icon-512.png", "sizes": "512x512", "type": "image/png", "purpose": "any" }
  ]
}
`

// Handler serves the embedded icons and manifest.
func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == ManifestPath {
			w.Header().Set("Content-Type", "application/manifest+json")
			w.Header().Set("Cache-Control", "public, max-age=3600")
			_, _ = w.Write([]byte(manifestJSON))
			return
		}

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
