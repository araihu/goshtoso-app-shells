// Package assets serves deterministic CSS and JavaScript for catalogshell.
package assets

import (
	_ "embed"
	"net/http"
	"strings"
)

const defaultPrefix = "/catalogshell/assets/"

var (
	//go:embed shell.css
	stylesheet []byte
	//go:embed shell.js
	script []byte
)

// Handler serves shell assets at their complete public paths. Consumers should
// mount it without StripPrefix.
func Handler() http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("X-Content-Type-Options", "nosniff")
		writer.Header().Set("Cache-Control", "public, max-age=3600")
		switch request.URL.Path {
		case defaultPrefix + "shell.css":
			writer.Header().Set("Content-Type", "text/css; charset=utf-8")
			_, _ = writer.Write(stylesheet)
		case defaultPrefix + "shell.js":
			writer.Header().Set("Content-Type", "text/javascript; charset=utf-8")
			_, _ = writer.Write(script)
		default:
			http.NotFound(writer, request)
		}
	})
}

// StylesheetURL returns the embedded stylesheet URL for an asset prefix.
func StylesheetURL(prefix string) string {
	return normalizePrefix(prefix) + "shell.css"
}

// ScriptURL returns the embedded script URL for an asset prefix.
func ScriptURL(prefix string) string {
	return normalizePrefix(prefix) + "shell.js"
}

func normalizePrefix(prefix string) string {
	if prefix == "" {
		return defaultPrefix
	}
	return strings.TrimSuffix(prefix, "/") + "/"
}
