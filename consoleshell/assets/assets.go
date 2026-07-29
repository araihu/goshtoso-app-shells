// Package assets serves deterministic console-shell CSS and runtime JavaScript.
package assets

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"net/http"
	"strings"
)

const defaultPrefix = "/consoleshell/assets/"

var (
	//go:embed shell.css
	stylesheet []byte
	//go:embed shell.js
	script            []byte
	stylesheetVersion = version(stylesheet)
	scriptVersion     = version(script)
)

func Handler(prefixes ...string) http.Handler {
	prefix := defaultPrefix
	if len(prefixes) > 0 {
		prefix = normalizePrefix(prefixes[0])
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		switch r.URL.Path {
		case prefix + "shell.css":
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
			_, _ = w.Write(stylesheet)
		case prefix + "shell.js":
			w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
			_, _ = w.Write(script)
		default:
			http.NotFound(w, r)
		}
	})
}
func StylesheetURL(prefix string) string {
	return normalizePrefix(prefix) + "shell.css?v=" + stylesheetVersion
}
func ScriptURL(prefix string) string { return normalizePrefix(prefix) + "shell.js?v=" + scriptVersion }
func version(b []byte) string        { s := sha256.Sum256(b); return hex.EncodeToString(s[:6]) }
func normalizePrefix(prefix string) string {
	if prefix == "" {
		return defaultPrefix
	}
	return strings.TrimSuffix(prefix, "/") + "/"
}
