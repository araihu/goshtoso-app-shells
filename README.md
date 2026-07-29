# Goshtoso App Shells

Reusable server-rendered application shell patterns built from
[Goshtoso](https://github.com/araihu/goshtoso) primitives.

The first package is `componentdocshell`, the shared frame for component documentation,
API references, design systems, and product documentation.
Catalog/browse experiences are a separate shell pattern and are not aliases of this package.

`consoleshell` is the companion application frame for operations consoles and
server-rendered HTMX products. It has a persistent header/sidebar/mobile drawer,
stable main fragment target, title/focus/scroll lifecycle, optional OOB navigation,
and first-paint theme state. It intentionally has no documentation, catalog, or TOC API.

## Console shell

```go
import (
  "github.com/araihu/goshtoso/assets"
  "github.com/araihu/goshtoso-app-shells/consoleshell"
  shellassets "github.com/araihu/goshtoso-app-shells/consoleshell/assets"
)

mux.Handle("GET /assets/", assets.Handler())
mux.Handle("GET /consoleshell/assets/", shellassets.Handler())

cfg := consoleshell.Config{
  Brand: consoleshell.Brand{Name: "Ops", HomeURL: "/"},
  Navigation: consoleshell.Navigation{Items: []sidebar.Item{
    {ID: "runs", Label: "Runs", Href: "/runs"},
  }},
  Appearance: consoleshell.AppearanceConfig{PersistPreferences: true},
  Interactions: consoleshell.InteractionConfig{
    EnableHTMX: true, NavigationOOB: true,
    // LocalRuntime: true, // explicit offline/no-CDN option
  },
}

page := consoleshell.Page{Title: "Runs", Active: "runs", Content: runsPage()}
component := consoleshell.Layout(cfg, page)
if request.Header.Get("HX-Request") == "true" { component = consoleshell.Fragment(cfg, page) }
```

Normal links remain normal `href`s. With HTMX enabled, shell-owned attributes
target `#main-content`, push history, preserve sidebar scroll, reset main scroll,
focus an explicit `[data-autofocus]` or page heading after settle, and close the
mobile drawer. Fragments contain one `<main>` only; use `NavigationOOB` when the
active sidebar must update in the same response. HTMX/Alpine swapped nodes are
left to framework lifecycle: the shell does not manually initialize them.

## Install

```bash
go get github.com/araihu/goshtoso-app-shells/componentdocshell
```

Mount both Goshtoso and component-doc-shell assets. Both handlers receive their full
public paths; do not wrap them in `http.StripPrefix`.

```go
import (
	"net/http"

	"github.com/araihu/goshtoso/assets"
	shellassets "github.com/araihu/goshtoso-app-shells/componentdocshell/assets"
)

mux.Handle("GET /assets/", assets.Handler())
mux.Handle("GET /componentdocshell/assets/", shellassets.Handler())
```

Define shell-wide presentation once, then supply route-specific pages:

```go
cfg := componentdocshell.Config{
	Brand: componentdocshell.Brand{Name: "My reference", HomeURL: "/"},
	Navigation: componentdocshell.Navigation{
		Items: []sidebar.Item{{ID: "overview", Label: "Overview", Href: "/"}},
		Sections: []sidebar.Section{{Title: "Components", Items: []sidebar.Item{
			{ID: "button", Label: "Button", Href: "/components/button"},
		}}},
	},
	Appearance: componentdocshell.AppearanceConfig{
		DefaultTheme: "araihu",
		InitialColorScheme: componentdocshell.ColorSchemeSystem,
		PersistPreferences: true,
	},
	Interactions: componentdocshell.InteractionConfig{EnableHTMX: true},
}

page := componentdocshell.Page{
	Title:   "Button",
	Active:  "button",
	Content: buttonReference(),
}

component := componentdocshell.Layout(cfg, page)
if request.Header.Get("HX-Request") == "true" {
	component = componentdocshell.Fragment(cfg, page)
}
_ = component.Render(request.Context(), writer)
```

`Layout` is a complete SSR document. Normal links work with JavaScript
disabled. When `Interactions.EnableHTMX` is true, `Fragment` updates the stable main-content
and sidebar contracts without giving page rendering to the browser.
Set `Interactions.LocalRuntime` when page content needs the embedded HTMX global
before body parsing completes; the default keeps Goshtoso's CDN-first loader.
`Interactions.RuntimeScripts` appends ordered scripts after eager local HTMX for
application-required extensions. `Navigation.SearchSlot` replaces the default
filter, while `BodyEnd` hosts application-owned modals, consent, or overlays.

The shell owns header, responsive navigation, grouped sidebar search, theme and
dark controls, scroll regions, optional TOC, focus handling, and embedded shell
assets. Applications retain routes, content, metadata values, authentication,
storage consent, analytics, and domain state.

Set `Page.DocumentTitle` when an existing site must preserve an exact
browser/SEO title. Otherwise the shell emits `Page.Title · Brand.Name`.

The default appearance includes canonical Arai Hû plus every theme compiled
into Goshtoso and selects Arai Hû. `AppearanceConfig` can replace or reorder
that list, choose the default and initial color scheme, hide either appearance
control, add consumer theme stylesheets, or disable the bundled Arai Hû theme.
Set `Appearance.PersistPreferences` only when the application permits browser
storage; otherwise selection stays in memory for the current document.

Existing applications can preserve a public dark-mode DOM/store contract with
`Appearance.DarkModeBinding`. Supply its button ID, Alpine state expression,
and toggle expression; empty fields retain shell defaults. Load any application
store registration through `Page.Head`, which renders before the shell runtime.
Set `Appearance.ThemeSelectorID` when existing automation or integrations depend
on the theme select's established DOM ID.
`Config.TOC` similarly preserves established rail/list IDs; shell behavior binds
through semantic data hooks and keeps `data-toc-link` on generated entries.

`componentpage.Page` renders the shared component-reference pattern: page
intro, optional controls, framed preview, usage code, and repeated variant
sections. Consumers retain every example component and copy string.
`componentpage.Section` renders the same secondary-example contract when a
consumer composes variants incrementally instead of passing `Page.Sections`.

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
