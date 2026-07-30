// Package server exposes the component docs shell example application.
package server

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-app-shells/componentdocshell"
	shellassets "github.com/araihu/goshtoso-app-shells/componentdocshell/assets"
	"github.com/araihu/goshtoso-app-shells/example/internal/pages"
	"github.com/araihu/goshtoso/assets"
)

const (
	fixtureRuntime = "/* fixture campaign runtime: intentionally inert */\n"
	fixtureChannel = `{"version":1,"campaigns":[]}`
	fixtureLogo    = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 120 32" role="img" aria-label="Component docs shell"><rect width="120" height="32" rx="4" fill="#1e293b"/><text x="12" y="21" fill="#fff" font-family="sans-serif" font-size="14">Goshtoso</text></svg>`
)

// New returns the standalone example handler.
func New() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /assets/", assets.Handler())
	mux.Handle("GET /componentdocshell/assets/", shellassets.Handler())
	mux.HandleFunc("GET /fixtures/campaign/v1.js", fixture("text/javascript; charset=utf-8", fixtureRuntime))
	mux.HandleFunc("GET /fixtures/releases/current", fixture("application/json; charset=utf-8", fixtureChannel))
	mux.HandleFunc("GET /fixtures/brand/logo.svg", fixture("image/svg+xml", fixtureLogo))
	mux.HandleFunc("GET /", func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/" {
			http.NotFound(writer, request)
			return
		}
		render(writer, request, pages.Overview())
	})
	mux.HandleFunc("GET /components/button", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, pages.Button())
	})
	return mux
}

func fixture(contentType, body string) http.HandlerFunc {
	return func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", contentType)
		_, _ = writer.Write([]byte(body))
	}
}

func render(writer http.ResponseWriter, request *http.Request, page componentdocshell.Page) {
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	var component templ.Component = componentdocshell.Layout(pages.ShellConfig(), page)
	if request.Header.Get("HX-Request") == "true" {
		component = componentdocshell.Fragment(pages.ShellConfig(), page)
	}
	if err := component.Render(request.Context(), writer); err != nil {
		http.Error(writer, "render component docs", http.StatusInternalServerError)
	}
}
