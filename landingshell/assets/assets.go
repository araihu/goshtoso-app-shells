// Package assets serves deterministic CSS and JavaScript for landingshell.
package assets

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"net/http"
	"strings"
)

const defaultPrefix = "/landingshell/assets/"

var (
	//go:embed shell.css
	stylesheet []byte
	//go:embed shell.js
	script            []byte
	stylesheetVersion = contentVersion(stylesheet)
	scriptVersion     = contentVersion(script)
)

// Handler serves shell assets at their complete public paths.
func Handler(prefixes ...string) http.Handler {
	prefix := defaultPrefix
	if len(prefixes) > 0 {
		prefix = normalizePrefix(prefixes[0])
	}
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("X-Content-Type-Options", "nosniff")
		switch request.URL.Path {
		case prefix + "shell.css":
			serve(writer, request, stylesheet, stylesheetVersion, "text/css; charset=utf-8")
		case prefix + "shell.js":
			serve(writer, request, script, scriptVersion, "text/javascript; charset=utf-8")
		default:
			http.NotFound(writer, request)
		}
	})
}

func serve(writer http.ResponseWriter, request *http.Request, content []byte, version, contentType string) {
	requestedVersion := request.URL.Query().Get("v")
	switch {
	case requestedVersion == "":
		writer.Header().Set("Cache-Control", "public, max-age=0, must-revalidate")
	case requestedVersion != version:
		writer.Header().Set("Cache-Control", "no-store")
		http.NotFound(writer, request)
		return
	default:
		writer.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	}
	writer.Header().Set("Content-Type", contentType)
	_, _ = writer.Write(content)
}

// StylesheetURL returns the content-versioned shell stylesheet URL.
func StylesheetURL(prefix string) string {
	return normalizePrefix(prefix) + "shell.css?v=" + stylesheetVersion
}

// ScriptURL returns the content-versioned shell script URL.
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
