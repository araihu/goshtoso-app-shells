// Package server exposes the catalog shell example application.
package server

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/assets"
	"github.com/araihu/goshtoso-app-shells/catalogshell"
	shellassets "github.com/araihu/goshtoso-app-shells/catalogshell/assets"
	"github.com/araihu/goshtoso-app-shells/example/internal/pages"
)

// New returns the standalone example handler.
func New() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /assets/", assets.Handler())
	mux.Handle("GET /catalogshell/assets/", shellassets.Handler())
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

func render(writer http.ResponseWriter, request *http.Request, page catalogshell.Page) {
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	var component templ.Component = catalogshell.Layout(pages.ShellConfig(), page)
	if request.Header.Get("HX-Request") == "true" {
		component = catalogshell.Fragment(pages.ShellConfig(), page)
	}
	if err := component.Render(request.Context(), writer); err != nil {
		http.Error(writer, "render catalog", http.StatusInternalServerError)
	}
}
