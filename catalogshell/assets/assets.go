// Package assets serves deterministic CSS and JavaScript for catalogshell.
package assets

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"net/http"
	"strings"
)

const defaultPrefix = "/catalogshell/assets/"

var (
	//go:embed shell.css
	stylesheet []byte
	//go:embed shell.js
	script []byte
	stylesheetVersion = contentVersion(stylesheet)
	scriptVersion = contentVersion(script)
)

// Handler serves shell assets at their complete public paths. Consumers should
// mount it without StripPrefix.
func Handler(prefixes ...string) http.Handler {
	prefix := defaultPrefix
	if len(prefixes) > 0 {
		prefix = normalizePrefix(prefixes[0])
	}
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("X-Content-Type-Options", "nosniff")
		writer.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		switch request.URL.Path {
		case prefix + "shell.css":
			writer.Header().Set("Content-Type", "text/css; charset=utf-8")
			_, _ = writer.Write(stylesheet)
		case prefix + "shell.js":
			writer.Header().Set("Content-Type", "text/javascript; charset=utf-8")
			_, _ = writer.Write(script)
		default:
			http.NotFound(writer, request)
		}
	})
}

// StylesheetURL returns the embedded stylesheet URL for an asset prefix.
func StylesheetURL(prefix string) string {
	return normalizePrefix(prefix) + "shell.css?v=" + stylesheetVersion
}

// ScriptURL returns the embedded script URL for an asset prefix.
func ScriptURL(prefix string) string {
	return normalizePrefix(prefix) + "shell.js?v=" + scriptVersion
}

func contentVersion(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:6])
}

func normalizePrefix(prefix string) string {
	if prefix == "" {
		return defaultPrefix
	}
	return strings.TrimSuffix(prefix, "/") + "/"
}
