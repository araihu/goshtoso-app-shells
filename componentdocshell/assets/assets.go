// Package assets serves deterministic CSS and JavaScript for componentdocshell.
package assets

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"net/http"
	"strings"
)

const defaultPrefix = "/componentdocshell/assets/"

var (
	//go:embed shell.css
	stylesheet []byte
	//go:embed shell.js
	script []byte
	//go:embed araihu.css
	araihuTheme []byte
	//go:embed goshtoso-logo.svg
	goshtosoLogo []byte
	//go:embed goshtoso-mark.svg
	goshtosoMark []byte
	//go:embed goshtoso-mark-reverse.svg
	goshtosoMarkReverse []byte
	//go:embed goshtoso-favicon.svg
	goshtosoFavicon            []byte
	stylesheetVersion          = contentVersion(stylesheet)
	scriptVersion              = contentVersion(script)
	araihuThemeVersion         = contentVersion(araihuTheme)
	goshtosoLogoVersion        = contentVersion(goshtosoLogo)
	goshtosoMarkVersion        = contentVersion(goshtosoMark)
	goshtosoMarkReverseVersion = contentVersion(goshtosoMarkReverse)
	goshtosoFaviconVersion     = contentVersion(goshtosoFavicon)
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
		case prefix + "araihu.css":
			writer.Header().Set("Content-Type", "text/css; charset=utf-8")
			_, _ = writer.Write(araihuTheme)
		case prefix + "goshtoso-logo.svg":
			serveSVG(writer, goshtosoLogo)
		case prefix + "goshtoso-mark.svg":
			serveSVG(writer, goshtosoMark)
		case prefix + "goshtoso-mark-reverse.svg":
			serveSVG(writer, goshtosoMarkReverse)
		case prefix + "goshtoso-favicon.svg":
			serveSVG(writer, goshtosoFavicon)
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

// AraiHuThemeURL returns the canonical Arai Hû theme stylesheet URL.
func AraiHuThemeURL(prefix string) string {
	return normalizePrefix(prefix) + "araihu.css?v=" + araihuThemeVersion
}

func GoshtosoLogoURL(prefix string) string {
	return versionedAssetURL(prefix, "goshtoso-logo.svg", goshtosoLogoVersion)
}
func GoshtosoMarkURL(prefix string) string {
	return versionedAssetURL(prefix, "goshtoso-mark.svg", goshtosoMarkVersion)
}
func GoshtosoMarkReverseURL(prefix string) string {
	return versionedAssetURL(prefix, "goshtoso-mark-reverse.svg", goshtosoMarkReverseVersion)
}
func GoshtosoFaviconURL(prefix string) string {
	return versionedAssetURL(prefix, "goshtoso-favicon.svg", goshtosoFaviconVersion)
}

func versionedAssetURL(prefix, name, version string) string {
	return normalizePrefix(prefix) + name + "?v=" + version
}

func serveSVG(writer http.ResponseWriter, content []byte) {
	writer.Header().Set("Content-Type", "image/svg+xml")
	_, _ = writer.Write(content)
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
