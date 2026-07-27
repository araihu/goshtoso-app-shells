# Goshtoso App Shells

Reusable server-rendered application shell patterns built from
[Goshtoso](https://github.com/araihu/goshtoso) primitives.

The first package is `catalogshell`, the shared frame for component catalogs,
API references, design systems, and product documentation.

## Install

```bash
go get github.com/araihu/goshtoso-app-shells/catalogshell
```

Mount both Goshtoso and catalog-shell assets. Both handlers receive their full
public paths; do not wrap them in `http.StripPrefix`.

```go
import (
	"net/http"

	"github.com/araihu/goshtoso/assets"
	shellassets "github.com/araihu/goshtoso-app-shells/catalogshell/assets"
)

mux.Handle("GET /assets/", assets.Handler())
mux.Handle("GET /catalogshell/assets/", shellassets.Handler())
```

Define shell-wide presentation once, then supply route-specific pages:

```go
cfg := catalogshell.Config{
	Brand: catalogshell.Brand{Name: "My reference", HomeURL: "/"},
	Navigation: catalogshell.Navigation{
		Items: []sidebar.Item{{ID: "overview", Label: "Overview", Href: "/"}},
		Sections: []sidebar.Section{{Title: "Components", Items: []sidebar.Item{
			{ID: "button", Label: "Button", Href: "/components/button"},
		}}},
	},
	EnableHTMX: true,
}

page := catalogshell.Page{
	Title:   "Button",
	Active:  "button",
	Content: buttonReference(),
}

component := catalogshell.Layout(cfg, page)
if request.Header.Get("HX-Request") == "true" {
	component = catalogshell.Fragment(cfg, page)
}
_ = component.Render(request.Context(), writer)
```

`Layout` is a complete SSR document. Normal links work with JavaScript
disabled. When `EnableHTMX` is true, `Fragment` updates the stable main-content
and sidebar contracts without giving page rendering to the browser.

The shell owns header, responsive navigation, grouped sidebar search, theme and
dark controls, scroll regions, optional TOC, focus handling, and embedded shell
assets. Applications retain routes, content, metadata values, authentication,
storage consent, analytics, and domain state.

Set `Page.DocumentTitle` when an existing site must preserve an exact
browser/SEO title. Otherwise the shell emits `Page.Title · Brand.Name`.

Set `PersistPreferences` only when the application permits browser storage.
The default keeps theme selection in memory for the current document.

## Example

```bash
go run ./example/cmd/server
```

Open `http://localhost:8092`. The example demonstrates full-page SSR, HTMX
fragments, a 720px persistent-sidebar breakpoint, mobile drawer, themes, and an
optional table-of-contents rail.

## Development

```bash
templ generate
go test ./...
go vet ./...
go build ./...
git diff --exit-code
```
