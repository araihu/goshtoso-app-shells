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

`landingshell` is the public-site frame for product and organization landing
pages. It owns metadata, first-paint color mode, a responsive brand header,
accessible mode and repository icons, the hero boundary, and a structured
linked footer. The consumer owns hero copy, product sections, calls to action,
and art direction.

## Landing shell

```go
import (
  "github.com/a-h/templ"
  "github.com/araihu/goshtoso-app-shells/landingshell"
)

cfg := landingshell.Config{
  Brand: landingshell.Brand{
    Name: "Product", HomeURL: "/", Tagline: "trusted remote assets",
    Logo: productLogo(),
    Badge: &landingshell.BrandBadge{Label: "v1.2.3", Href: "/releases/v1.2.3"},
  },
  Navigation: []landingshell.Link{{Label: "Docs", Href: "/docs", Primary: true}},
  Appearance: landingshell.AppearanceConfig{
    DefaultTheme: "araihu", InitialColorScheme: landingshell.ColorSchemeSystem,
    PersistPreferences: true,
  },
  Footer: landingshell.Footer{
    Meta: []string{"trusted remote assets"},
    Organization: &landingshell.Organization{Name: "Arai Hû", URL: "https://araihu.com"},
    Links: []landingshell.Link{{Label: "Docs", Href: "/docs"}},
  },
  RepositoryURL: "https://github.com/example/product",
}
page := landingshell.Page{
  Title: "Home", Description: "Product description",
  Hero: hero(), Content: content(), Head: templ.Raw(`<link rel="stylesheet" href="/styles/product.css">`),
}
_ = landingshell.Layout(cfg, page).Render(ctx, writer)
```

Set `Footer.HideBrand` when the footer should retain its navigation without
repeating the product logo, name, metadata, or organization.

Mount `landingshell/assets.Handler()` at `/landingshell/assets/` for a server.
Static generators can request `assets.StylesheetURL("")` and
`assets.ScriptURL("")` from the handler at build time, preserving the exact
content-versioned paths emitted by `Layout`. Set `Interactions.LocalRuntime`
when every Goshtoso runtime byte must be served locally. When persistence is
disabled or browser storage is unavailable, color-mode changes remain
session-only and still update the document.

Landing pages can opt into a floating mobile trigger and top Drawer without
moving that responsive policy into product CSS:

```go
cfg.MobileNavigation = &landingshell.MobileNavigationConfig{
  ID: "product-navigation", Title: "Navigation", TriggerLabel: "Menu",
  NavigationLabel: "Primary navigation",
  Position: landingshell.FloatingBottomLeft,
}
```

Custom page frames can compose the same policy directly and retain their own
brand-specific navigation content:

```go
menu := landingshell.MobileNavigation(
  landingshell.MobileNavigationConfig{Title: "Navigation", TriggerLabel: "Menu"},
  productNavigationLinks(),
)
```

The enhanced path uses Goshtoso's top Drawer once Alpine initializes. Until
then, the same slot remains available through a native `<details>` fallback;
JavaScript-disabled pages therefore keep complete navigation. Because the slot
is rendered for both paths, keep its IDs unique or omit them. `ActionGroup`
semantics remain unchanged: the landing shell owns the breakpoint, fixed
trigger placement, and Drawer composition.

## Console shell

```go
import (
  "github.com/araihu/goshtoso/assets"
  "github.com/araihu/goshtoso/components/sidebar"
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
_ = component.Render(request.Context(), writer)
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

### Component docs family navigation

`family navigation` is the global product-family layer. `local navigation` (also
called the `scoped sidebar`) is the route's navigation inside the active family.
`Page.ActiveFamily` identifies the active family and renders
`aria-current="location"`; `Page.Active` identifies the active local page and
renders `aria-current="page"`. Optional `ScopeMetadata` adds module path,
version, and version URL in the scoped sidebar. Consumers provide local
`Items`, `Sections`, `SearchSlot`, and scope metadata for each route. Selecting
or switching a family opens that family's configured overview `Href`; the shell
never maps a local page to an analogous page in another family.
The stable family order and overview routes are Components (`/components`),
Charts (`/charts`), App Shells (`/app-shells`), Icons (`/icons`), LLMs
(`/llms`), and Examples (`/examples`).

The following is one self-contained, copyable example. Every public struct uses
a keyed literal. The module, version, and release URL in this fixture are
illustrative values, not release metadata.

```go
package docs

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-app-shells/componentdocshell"
	"github.com/araihu/goshtoso/components/sidebar"
)

var docsConfig = componentdocshell.Config{
	Brand: componentdocshell.Brand{Name: "Example docs", HomeURL: "/"},
	Navigation: componentdocshell.Navigation{
		Families: []componentdocshell.FamilyLink{
			{ID: "components", Label: "Components", Href: "/components"},
			{ID: "charts", Label: "Charts", Href: "/charts"},
			{ID: "app-shells", Label: "App Shells", Href: "/app-shells"},
			{ID: "icons", Label: "Icons", Href: "/icons"},
			{ID: "llms", Label: "LLMs", Href: "/llms"},
			{ID: "examples", Label: "Examples", Href: "/examples"},
		},
		// Optional illustrative scope; remove it when this route has no module
		// or independently released version.
		Scope: &componentdocshell.ScopeMetadata{
			ModulePath: "example.com/componentdocshell",
			ModuleLabel: "example/componentdocshell",
			ModuleURL:  "https://example.com/componentdocshell",
			Version:    "v0.0.0-example",
			VersionURL: "https://example.com/componentdocshell/releases/v0.0.0-example",
		},
		Items: []sidebar.Item{{ID: "overview", Label: "Overview", Href: "/components"}},
	},
	Appearance: componentdocshell.AppearanceConfig{
		DefaultTheme:       "araihu",
		InitialColorScheme: componentdocshell.ColorSchemeSystem,
		PersistPreferences: true,
	},
	Interactions: componentdocshell.InteractionConfig{EnableHTMX: true},
}

func renderDocs(w http.ResponseWriter, request *http.Request) {
	page := componentdocshell.Page{
		Title:        "Components",
		Description:  "Component documentation overview.",
		ActiveFamily: "components",
		Active:       "overview",
		Content:      templ.Raw(`<h1>Components</h1>`),
	}
	view := componentdocshell.Layout(docsConfig, page)
	if request.Header.Get("HX-Request") == "true" {
		view = componentdocshell.Fragment(docsConfig, page)
	}
	if err := view.Render(request.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
```

Families are optional. An empty `Navigation.Families` slice preserves legacy
behavior, emits no family surfaces, and does not require `Page.ActiveFamily`.
When families are non-empty, IDs must be unique and have no leading or trailing
whitespace; labels and overview URLs are required; overview and version URLs
must be root-relative or absolute HTTPS; and `Page.ActiveFamily` must match a
configured ID. `ScopeMetadata.ModuleLabel` and `ModuleURL` require `ModulePath`;
`ModuleLabel` defaults to the canonical module path. `VersionURL` requires
`Version`. Validation
finishes before any layout or fragment bytes are written. The public model is
additive for behavior, zero values, and keyed literals; adding exported fields
is not source-compatible with external positional literals, so supported
examples and consumers should use keyed literals.

`FamilyLink.LinkAttrs` is copied without mutating the caller's map; unrelated
attributes are retained. The shell always owns `aria-current` on both
responsive surfaces. When HTMX enhancement is enabled, it also owns `hx-get`,
`hx-target`, and `hx-push-url`. `LinkAttrs` must not contain an `id` attribute,
case-insensitively, in any configuration: each family link appears as
responsive duplicate anchors and the shell owns their IDs. The ordinary `Href`
anchor remains in every mode, so no-JavaScript/full-page navigation works
whether HTMX is enabled or disabled. With HTMX enabled, `Fragment` returns the
title plus exactly one out-of-band replacement
for `#main-content`, `#componentdocshell-sidebar-content` (the scoped sidebar),
and `#componentdocshell-family-navigation`. Desktop family links use
`aria-current="location"`; the small-screen Goshtoso Select exposes the same
state with `aria-selected="true"`; the active local page stays
`aria-current="page"`. The preserved Select is synchronized after HTMX swaps.
Without JavaScript, a six-link navigation fallback remains available.

Responsive family navigation has three exact ranges: small `<720px` uses a
64px row with the brand, current-family Goshtoso Select, and dark-mode control;
medium `720px–1439px` uses one uninterrupted shared header surface
containing a 64px brand/control row plus a 44px family row (108px total); wide
`>=1440px` uses one 64px row with inline family links.
The local sidebar remains a drawer with its menu trigger below `1024px` and
becomes persistent at `>=1024px`.
At medium and wide widths, family links do not shrink and the navigation region
scrolls horizontally when localization, long labels, or additional families
exceed its available width, so every configured destination remains reachable.
On small layouts, the built-in theme selector and repository link move to the
drawer utilities. Set `Brand.CompactLogo` to a purpose-built small mark; when it
is empty, the shell uses the first rune of `Brand.Name`. `HeaderActions` is
rendered once and is never cloned or
moved; consumers own its responsive reachability, IDs, and state. `BrandBadge`
remains supported. Goshtoso may omit a global version badge because families
release independently, but that is a future consumer configuration choice, not
a removal from this public API.

This package change does not claim Goshtoso adoption, complete Charts or App
Shells catalogs, a release, deployment, or accessibility certification. Those
adoption, catalog, release, and lifecycle steps require separate authorization.

`Layout` is a complete SSR document. `Interactions.LocalRuntime` opts into a
local HTMX runtime; otherwise Goshtoso's CDN-first loader is used.
`Interactions.RuntimeScripts` appends ordered scripts after eager local HTMX for
application-required extensions. `Navigation.SearchSlot` replaces the default
filter, while `BodyEnd` hosts application-owned modals, consent, or overlays.

Each route may provide `Page.DocumentTitle`, `Description`, an absolute HTTPS
`CanonicalURL`, `SiteName`, `Locale`, and a typed `SocialImage`. Complete social
metadata is emitted in the initial SSR document for Open Graph and X. When a
social image is configured, its URL must be absolute HTTPS and its MIME type,
positive pixel dimensions, and descriptive alt text are required. `SiteName`
defaults to `Brand.Name`; zero-value metadata keeps existing consumers working.

The shell owns header, responsive navigation, grouped sidebar search, theme and
dark controls, scroll regions, optional TOC, focus handling, and embedded shell
assets. Applications retain routes, content, metadata values, authentication,
storage consent, analytics, and domain state.

Maintainers refreshing embedded theme or brand fallbacks should follow the
[immutable Arai Hu asset update contract](docs/ARAIHU_ASSETS.md).

## Local CI with Dagger

CI uses Dagger 0.21.8 for the same Go 1.26.5 and templ 0.3.1020 workload locally
and on GitHub Actions:

```bash
dagger call ci --source=. --cache-namespace=local --run-nonce=local
dagger call browser --source=. --cache-namespace=local --run-nonce=local
dagger call benchmark --source=. --cache-namespace=local --run-nonce=local
```

`ci` generates templ files and rejects drift, then runs every test, `go vet`,
and `go build`. `browser` preserves the dedicated Playwright Chromium gate for
the component documentation shell. `benchmark` preserves the existing
cold/warm marker and exact workload. GitHub-hosted jobs install Dagger 0.21.8
through the commit-pinned installer action; self-hosted jobs require the
embedded CLI to report exactly v0.21.8. Both invoke the verified CLI directly.

Fallback updates use `assets-update`. Provide the provider-owned event JSON as
a `File`, its event name, and the read-only GitHub token as a Dagger `Secret`:

```bash
dagger call assets-update \
  --source=. \
  --provider-event=.dagger-input/assets-provider-event.json \
  --event-name=repository_dispatch \
  --github-token=env://GH_TOKEN \
  --cache-namespace=trusted \
  --run-nonce=local \
  export --path=.dagger-output/assets
```

The function extracts exactly six allowed identity fields and validates the
provider event before admitting the secret, verifies tag and
archive identities, rejects unsafe archive members, runs the updater twice to
prove idempotence, and returns only allowlisted files. GitHub Actions owns App
token creation, label discovery, and creation or update of the non-auto-merged
pull request. Runner-host steps require only Bash, Git, Dagger, and
commit-pinned JavaScript actions; `jq` remains pinned inside Dagger.

Every pull request mounts persistent Go module, build, and Playwright caches in
stable namespace `pr`. Protected `main` pushes and asset-update jobs use
`trusted`; non-`main` pushes run on GitHub-hosted runners with `branch-hosted`.
GitHub-hosted benchmark and local runs retain separate efficiency namespaces.
Only dependencies, build output, and browser tooling are cached. Function
results remain uncached.

Cache namespace is an efficiency hint, not an authorization boundary. Pull
requests run only on `hostinger-vps-pr`; protected `main` push, asset-update,
and self-hosted benchmark jobs use `hostinger-vps-trusted`. Other branch pushes
run on `ubuntu-24.04`. Isolated Engine
socket/data and host ACLs prevent PR workloads from reaching trusted cache
storage even if PR-owned code requests another cache name. Workflow arguments
do not establish isolation or authorization.

## Presentation channels

Presentation channels are opt-in. The shell only renders declared integration
hooks; it never fetches a channel, chooses a campaign, or changes campaign
policy. Configure a fixed-size, channel-managed logo and an integrity-pinned
deferred runtime explicitly:

```go
cfg := componentdocshell.Config{
	Brand: componentdocshell.Brand{
		Name: "My reference", HomeURL: "/",
		ManagedLogo: &componentdocshell.ManagedBrandAsset{
			URL: "/brand/logo.svg", Alt: "My reference", Width: 120, Height: 32,
		},
	},
	Interactions: componentdocshell.InteractionConfig{
		PresentationChannel: &componentdocshell.PresentationChannelConfig{
			RuntimeURL:       "/campaign/v1.js",
			ChannelURL:       "/releases/current",
			Integrity:        "sha384-<base64 digest of exact runtime bytes>",
			UseCampaignLabel: "Use seasonal appearance",
			UseBaselineLabel: "Use standard appearance",
		},
	},
}
```

`ManagedLogo.Width` and `Height` are required positive values; shell layout
reserves that box before the image loads. Runtime and channel URLs must both be
root-relative, or both be same-origin HTTPS URLs. The runtime is deferred after
the first-paint bootstrap and receives the channel URL, SHA-384 SRI, and
anonymous cross-origin mode.

Before the deferred runtime can execute, the root has
`data-theme-source="default"` unless an application-owned saved theme was read,
when it has `data-theme-source="preference"`. A presentation runtime must trust
that root marker instead of rereading browser storage. Applications own storage
consent, existing preference keys, and clearing preferences; App Shells owns no
campaign opt-out storage. If storage is unavailable, the configured default and
`default` marker remain. If the runtime or channel fails integrity or loading,
the managed baseline remains and the campaign toggle stays hidden.

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
intro, optional controls, state-labelled preview, usage code, and repeated
variant sections. Consumers retain every example component and copy string.
Set `Example.PreviewLabel` when the rendered state needs a label other than
`Default` for the primary example or the secondary section title. An unnamed
secondary example falls back to `Preview`.
`componentpage.Section` renders the same secondary-example contract when a
consumer composes variants incrementally instead of passing `Page.Sections`.

## Example

```bash
go run ./example/cmd/server
```

Open `http://localhost:8092`. The example demonstrates full-page SSR, ordinary
links, HTMX fragments, all six family overview routes, the `<720px`,
`720px–1439px`, and `>=1440px` layouts, mobile drawer, themes, and an optional
table-of-contents rail.

Run the browser and unrelated-consumer proofs from the repository root:

```bash
COMPONENTDOCSHELL_E2E=1 GOWORK=off go test ./example/e2e -count=1
./scripts/test-componentdocshell-external-consumer.sh
GOWORK=off go test ./... -count=1
GOWORK=off go vet ./...
GOWORK=off go build ./...
go mod verify
git diff --check
```

## Development

```bash
templ generate
go test ./...
go vet ./...
go build ./...
git diff --exit-code
```

## Deferred test debt

- Remove unused `consoleshell` shell-runtime persistence code.
- Manual VoiceOver/Safari and NVDA/Chrome review remains a separate release gate;
  this package does not claim accessibility certification.
- Replace substring/index markup checks with parsed-HTML assertions for exactly-one and attribute ownership.
