# Component Docs Shell Family Navigation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add responsive, accessible product-family navigation and active-family scope metadata to `componentdocshell` without changing consumers that leave `Navigation.Families` empty.

**Architecture:** Extend the existing shell model with typed family links and scope metadata, validate the full model before rendering, and derive immutable render copies carrying shell-owned HTMX attributes. Render one family-navigation replacement containing wide/medium links and a small-screen semantic disclosure; CSS selects the 64px, 108px, and 64px layouts at the fixed 720px and 1200px boundaries. Expand the standalone example into six family roots, then prove SSR, HTMX identity, responsive behavior, accessibility semantics, storage failure, and public-package consumption.

**Tech Stack:** Go 1.26.5, templ v0.3.1020, Goshtoso v0.1.6, Alpine.js, HTMX, embedded CSS/JavaScript, Playwright-Go v0.5700.1, `net/http`/`httptest`.

## Global Constraints

- Family order is exactly Components, Charts, App Shells, Icons, LLMs, Examples.
- Family overview routes are exactly `/components`, `/charts`, `/app-shells`, `/icons`, `/llms`, `/examples`.
- Small layout is below 720px; medium layout is 720px through 1199px; wide layout begins at 1200px.
- Small header is one 64px row; medium header is a 64px brand/control row plus a 44px family row; wide header is one 64px row.
- Family links stay ordinary anchors and use `aria-current="location"`; exact local sidebar pages retain `aria-current="page"`.
- Configured family URLs and version URLs accept only root-relative or absolute HTTPS values.
- Validation must finish before any document or fragment bytes are written.
- Empty `Navigation.Families` preserves existing rendering and imposes no `Page.ActiveFamily` requirement.
- Consumer-provided family and sidebar link attribute maps must never be mutated.
- Fragment responses include title plus exactly one OOB replacement each for main content, scoped sidebar, and configured family navigation.
- `BrandBadge` remains public; the App Shell package does not remove it.
- `HeaderActions` renders once and remains consumer-owned; the shell does not clone it into the mobile drawer.
- This plan does not change Goshtoso search grouping, footer links, family page content, dependency version, or producer documentation ownership.
- Merge, tag, release, publish, deploy, Goshtoso adoption, and worktree cleanup require separate authorization.

## File Structure

- `componentdocshell/config.go`: public `FamilyLink`, `ScopeMetadata`, `Navigation.Families`, `Navigation.Scope`, and `Page.ActiveFamily` fields plus stable DOM-ID helpers.
- `componentdocshell/validation.go`: family/scope validation and shared safe-navigation-URL validation.
- `componentdocshell/render.go`: immutable family-link derivation, active state, active label lookup, and OOB attribute helpers.
- `componentdocshell/family_navigation.templ`: family link list, small-screen disclosure, active state, and scope metadata.
- `componentdocshell/family_navigation_templ.go`: generated templ output; never hand-edit.
- `componentdocshell/layout.templ`: header grid, sidebar utility region, and family-aware render calls.
- `componentdocshell/layout_templ.go`: generated templ output; never hand-edit.
- `componentdocshell/fragment.templ`: family-navigation OOB replacement.
- `componentdocshell/fragment_templ.go`: generated templ output; never hand-edit.
- `componentdocshell/config_test.go`: public model validation and zero-byte failure coverage.
- `componentdocshell/componentdocshell_test.go`: rendered semantics, immutability, and fragment identity coverage.
- `componentdocshell/assets/shell.css`: three responsive layouts and shared header-height variable.
- `componentdocshell/assets/shell.js`: post-navigation menu closure and existing focus/scroll lifecycle.
- `componentdocshell/assets/assets_test.go`: CSS/JavaScript contract assertions.
- `example/internal/pages/pages.go`: six-family fixture model, active scope, and family pages.
- `example/internal/pages/pages.templ`: reusable family-overview content.
- `example/internal/pages/pages_templ.go`: generated templ output; never hand-edit.
- `example/internal/server/server.go`: six overview routes and family-aware shell rendering.
- `example/internal/server/server_test.go`: direct and fragment route contracts.
- `example/e2e/family_navigation_test.go`: browser layout, navigation, focus, theme, storage, and console coverage.
- `.github/workflows/ci.yml`: Chromium-backed example E2E job.
- `go.mod`, `go.sum`: Playwright-Go test dependency.
- `testdata/external-consumer/go.mod`: unrelated-module fixture requiring the public App Shell module.
- `testdata/external-consumer/main_test.go`: public API render proof.
- `scripts/test-componentdocshell-external-consumer.sh`: clean temporary-module runner with `GOWORK=off`.
- `README.md`: family-navigation API, responsive behavior, example routes, and proof commands.

---

### Task 1: Add and validate the public family model

**Files:**
- Modify: `componentdocshell/config.go`
- Modify: `componentdocshell/validation.go`
- Modify: `componentdocshell/config_test.go`

**Interfaces:**
- Consumes: existing `Config`, `Navigation`, `Page`, and `validate(Config, Page, bool) error`.
- Produces: `FamilyLink`, `ScopeMetadata`, `Navigation.Families`, `Navigation.Scope`, `Page.ActiveFamily`, and `validateFamilyNavigation(Config, Page) error`.

- [ ] **Step 1: Write failing public-model validation tests**

Add fixtures and table tests in `componentdocshell/config_test.go`:

```go
func validFamilyConfig() Config {
	cfg := validConfig()
	cfg.Navigation.Families = []FamilyLink{
		{ID: "components", Label: "Components", Href: "/components"},
		{ID: "charts", Label: "Charts", Href: "https://docs.example/charts"},
	}
	cfg.Navigation.Scope = &ScopeMetadata{
		ModulePath: "github.com/araihu/goshtoso",
		Version:    "v0.1.6",
		VersionURL: "https://github.com/araihu/goshtoso/releases/tag/v0.1.6",
	}
	return cfg
}

func validFamilyPage() Page {
	page := validPage()
	page.ActiveFamily = "components"
	return page
}

func TestValidateAcceptsFamilyNavigation(t *testing.T) {
	t.Parallel()
	if err := validate(validFamilyConfig(), validFamilyPage(), false); err != nil {
		t.Fatalf("validate() error = %v", err)
	}
}

func TestValidateRejectsInvalidFamilyNavigation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		edit func(*Config, *Page)
		want string
	}{
		{"empty ID", func(cfg *Config, _ *Page) { cfg.Navigation.Families[0].ID = " " }, "family ID is required"},
		{"duplicate ID", func(cfg *Config, _ *Page) { cfg.Navigation.Families[1].ID = "components" }, `duplicate family ID "components"`},
		{"empty label", func(cfg *Config, _ *Page) { cfg.Navigation.Families[0].Label = " " }, `family "components" label is required`},
		{"empty URL", func(cfg *Config, _ *Page) { cfg.Navigation.Families[0].Href = "" }, `family "components" URL is required`},
		{"HTTP URL", func(cfg *Config, _ *Page) { cfg.Navigation.Families[0].Href = "http://docs.example/components" }, "must be root-relative or an absolute HTTPS URL"},
		{"scheme-relative URL", func(cfg *Config, _ *Page) { cfg.Navigation.Families[0].Href = "//docs.example/components" }, "must be root-relative or an absolute HTTPS URL"},
		{"unknown active family", func(_ *Config, page *Page) { page.ActiveFamily = "icons" }, `active family ID "icons" is not configured`},
		{"missing active family", func(_ *Config, page *Page) { page.ActiveFamily = "" }, "active family ID is required"},
		{"version URL without version", func(cfg *Config, _ *Page) { cfg.Navigation.Scope.Version = "" }, "scope version URL requires a version"},
		{"HTTP version URL", func(cfg *Config, _ *Page) { cfg.Navigation.Scope.VersionURL = "http://docs.example/v0.1.6" }, "scope version URL must be root-relative or an absolute HTTPS URL"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg, page := validFamilyConfig(), validFamilyPage()
			test.edit(&cfg, &page)
			err := validate(cfg, page, false)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("validate() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestValidateKeepsEmptyFamiliesBackwardCompatible(t *testing.T) {
	t.Parallel()
	page := validPage()
	page.ActiveFamily = ""
	if err := validate(validConfig(), page, false); err != nil {
		t.Fatalf("validate() error = %v", err)
	}
}
```

- [ ] **Step 2: Run focused tests and confirm compile failure**

Run:

```bash
GOWORK=off go test ./componentdocshell -run 'TestValidate(AcceptsFamilyNavigation|RejectsInvalidFamilyNavigation|KeepsEmptyFamiliesBackwardCompatible)' -count=1
```

Expected: FAIL because `FamilyLink`, `ScopeMetadata`, and `Page.ActiveFamily` do not exist.

- [ ] **Step 3: Add exact public fields**

Add to `componentdocshell/config.go`:

```go
type FamilyLink struct {
	ID        string
	Label     string
	Href      string
	LinkAttrs templ.Attributes
}

type ScopeMetadata struct {
	ModulePath string
	Version    string
	VersionURL string
}

type Navigation struct {
	Families          []FamilyLink
	Scope             *ScopeMetadata
	Items             []sidebar.Item
	SectionsTitle     string
	Sections          []sidebar.Section
	SearchPlaceholder string
	DisableSearch     bool
	SearchSlot        templ.Component
}

type Page struct {
	Title        string
	DocumentTitle string
	Description  string
	CanonicalURL string
	ActiveFamily string
	Active       string
	Content      templ.Component
	Head         templ.Component
	EnableTOC    bool
}
```

Preserve existing comments while inserting `ActiveFamily`; run `gofmt` after the edit.

- [ ] **Step 4: Implement family and scope validation**

Call `validateFamilyNavigation(cfg, page)` after base page/content checks and before navigation-item validation. Reuse the existing URL parser through a field-aware helper:

```go
func validateFamilyNavigation(cfg Config, page Page) error {
	if len(cfg.Navigation.Families) == 0 {
		return validateScopeMetadata(cfg.Navigation.Scope)
	}

	ids := make(map[string]struct{}, len(cfg.Navigation.Families))
	for _, family := range cfg.Navigation.Families {
		id := strings.TrimSpace(family.ID)
		if id == "" {
			return fmt.Errorf("component docs shell family ID is required")
		}
		if _, exists := ids[id]; exists {
			return fmt.Errorf("component docs shell duplicate family ID %q", id)
		}
		ids[id] = struct{}{}
		if strings.TrimSpace(family.Label) == "" {
			return fmt.Errorf("component docs shell family %q label is required", id)
		}
		if _, err := validatePresentationURL("family "+strconv.Quote(id)+" URL", family.Href); err != nil {
			return err
		}
	}
	if strings.TrimSpace(page.ActiveFamily) == "" {
		return fmt.Errorf("component docs shell active family ID is required")
	}
	if _, exists := ids[page.ActiveFamily]; !exists {
		return fmt.Errorf("component docs shell active family ID %q is not configured", page.ActiveFamily)
	}
	return validateScopeMetadata(cfg.Navigation.Scope)
}

func validateScopeMetadata(scope *ScopeMetadata) error {
	if scope == nil {
		return nil
	}
	if scope.VersionURL != "" && strings.TrimSpace(scope.Version) == "" {
		return fmt.Errorf("component docs shell scope version URL requires a version")
	}
	if scope.VersionURL != "" {
		if _, err := validatePresentationURL("scope version URL", scope.VersionURL); err != nil {
			return err
		}
	}
	return nil
}
```

Import `strconv`. Keep the existing root-relative/HTTPS URL implementation and its control-character, credentials, fragment, and backslash protections.

- [ ] **Step 5: Prove validation writes zero bytes**

Add a render-level regression test:

```go
func TestFamilyValidationFailsBeforeWritingBytes(t *testing.T) {
	t.Parallel()
	cfg, page := validFamilyConfig(), validFamilyPage()
	cfg.Navigation.Families[0].Href = "javascript:alert(1)"
	for _, component := range []templ.Component{Layout(cfg, page), Fragment(withHTMX(cfg), page)} {
		var buffer bytes.Buffer
		if err := component.Render(context.Background(), &buffer); err == nil {
			t.Fatal("Render() accepted invalid family URL")
		}
		if buffer.Len() != 0 {
			t.Fatalf("Render() wrote %d bytes before validation failure", buffer.Len())
		}
	}
}
```

Add a local `withHTMX` test helper that copies `Config`, sets `Interactions.EnableHTMX = true`, and returns the copy.

- [ ] **Step 6: Run focused tests**

Run:

```bash
gofmt -w componentdocshell/config.go componentdocshell/validation.go componentdocshell/config_test.go
GOWORK=off go test ./componentdocshell -run 'Test(Validate.*Family|FamilyValidation)' -count=1
```

Expected: PASS.

- [ ] **Step 7: Commit the validated public model**

```bash
git add componentdocshell/config.go componentdocshell/validation.go componentdocshell/config_test.go
git commit -m "feat: add component docs family model"
```

### Task 2: Derive immutable family render state

**Files:**
- Modify: `componentdocshell/render.go`
- Modify: `componentdocshell/componentdocshell_test.go`

**Interfaces:**
- Consumes: `[]FamilyLink`, `Page.ActiveFamily`, and `InteractionConfig.EnableHTMX`.
- Produces: `familyLinks(Config, string) []FamilyLink`, `activeFamilyLabel(Config, Page) string`, and `familyNavigationOOBAttributes(bool) templ.Attributes`.

- [ ] **Step 1: Write failing immutability and HTMX tests**

```go
func TestFamilyLinksCloneAttributesWithoutMutation(t *testing.T) {
	t.Parallel()
	cfg := validFamilyConfig()
	cfg.Interactions.EnableHTMX = true
	cfg.Navigation.Families[0].LinkAttrs = templ.Attributes{
		"data-analytics": "components-family",
		"hx-target":      "#consumer-target",
	}

	links := familyLinks(cfg, "components")
	if links[0].LinkAttrs["data-analytics"] != "components-family" {
		t.Fatal("familyLinks() dropped consumer attribute")
	}
	for key, want := range map[string]any{
		"hx-get":      "/components",
		"hx-target":   "#main-content",
		"hx-push-url": "true",
	} {
		if got := links[0].LinkAttrs[key]; got != want {
			t.Errorf("familyLinks()[0].LinkAttrs[%q] = %#v, want %#v", key, got, want)
		}
	}
	if got := cfg.Navigation.Families[0].LinkAttrs["hx-target"]; got != "#consumer-target" {
		t.Fatalf("familyLinks() mutated caller attribute: %#v", got)
	}
	if _, exists := cfg.Navigation.Families[0].LinkAttrs["hx-get"]; exists {
		t.Fatal("familyLinks() added hx-get to caller map")
	}
	if got := links[0].LinkAttrs["aria-current"]; got != "location" {
		t.Fatalf("familyLinks() active aria-current = %#v, want location", got)
	}
	if _, exists := links[1].LinkAttrs["aria-current"]; exists {
		t.Fatal("familyLinks() marked inactive family current")
	}
}

func TestFamilyLinksStayOrdinaryWithoutHTMX(t *testing.T) {
	t.Parallel()
	cfg := validFamilyConfig()
	links := familyLinks(cfg, "components")
	if _, exists := links[0].LinkAttrs["hx-get"]; exists {
		t.Fatal("familyLinks() added HTMX attributes while HTMX is disabled")
	}
}
```

- [ ] **Step 2: Run focused tests and confirm missing helpers**

```bash
GOWORK=off go test ./componentdocshell -run 'TestFamilyLinks' -count=1
```

Expected: FAIL with `undefined: familyLinks`.

- [ ] **Step 3: Implement immutable render helpers**

Add to `componentdocshell/render.go`:

```go
func familyLinks(cfg Config, active string) []FamilyLink {
	result := make([]FamilyLink, len(cfg.Navigation.Families))
	for index, family := range cfg.Navigation.Families {
		result[index] = family
		if family.LinkAttrs != nil {
			result[index].LinkAttrs = make(templ.Attributes, len(family.LinkAttrs)+4)
			for key, value := range family.LinkAttrs {
				result[index].LinkAttrs[key] = value
			}
		}
		if family.ID == active {
			if result[index].LinkAttrs == nil {
				result[index].LinkAttrs = templ.Attributes{}
			}
			result[index].LinkAttrs["aria-current"] = "location"
		} else if result[index].LinkAttrs != nil {
			delete(result[index].LinkAttrs, "aria-current")
		}
		if cfg.Interactions.EnableHTMX {
			if result[index].LinkAttrs == nil {
				result[index].LinkAttrs = templ.Attributes{}
			}
			result[index].LinkAttrs["hx-get"] = family.Href
			result[index].LinkAttrs["hx-target"] = "#main-content"
			result[index].LinkAttrs["hx-push-url"] = "true"
		}
	}
	return result
}

func activeFamilyLabel(cfg Config, page Page) string {
	for _, family := range cfg.Navigation.Families {
		if family.ID == page.ActiveFamily {
			return family.Label
		}
	}
	return ""
}

func familyNavigationOOBAttributes(enabled bool) templ.Attributes {
	if !enabled {
		return nil
	}
	return templ.Attributes{"hx-swap-oob": "outerHTML:#componentdocshell-family-navigation"}
}
```

- [ ] **Step 4: Run package tests**

```bash
gofmt -w componentdocshell/render.go componentdocshell/componentdocshell_test.go
GOWORK=off go test ./componentdocshell -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit immutable render state**

```bash
git add componentdocshell/render.go componentdocshell/componentdocshell_test.go
git commit -m "feat: derive component docs family links"
```

### Task 3: Render family navigation and scoped sidebar metadata

**Files:**
- Create: `componentdocshell/family_navigation.templ`
- Create: `componentdocshell/family_navigation_templ.go`
- Modify: `componentdocshell/layout.templ`
- Modify: `componentdocshell/layout_templ.go`
- Modify: `componentdocshell/fragment.templ`
- Modify: `componentdocshell/fragment_templ.go`
- Modify: `componentdocshell/config.go`
- Modify: `componentdocshell/componentdocshell_test.go`

**Interfaces:**
- Consumes: `familyLinks(Config, string)`, `activeFamilyLabel(Config, Page)`, `Navigation.Scope`, and existing sidebar config.
- Produces: one `#componentdocshell-family-navigation` wrapper, `familyNavigation(Config, Page, bool)`, `scopeMetadata(Config, Page)`, and `mobileUtilities(Config)` templ components.

- [ ] **Step 1: Write failing semantic render tests**

Add tests covering configured and empty family lists:

```go
func TestLayoutRendersFamilyNavigationAndScope(t *testing.T) {
	t.Parallel()
	cfg, page := validFamilyConfig(), validFamilyPage()
	body := renderLayout(t, cfg, page)
	for _, want := range []string{
		`id="componentdocshell-family-navigation"`,
		`aria-label="Documentation families"`,
		`<details class="component-doc-shell__family-menu"`,
		`aria-current="location"`,
		`github.com/araihu/goshtoso`,
		`href="https://github.com/araihu/goshtoso/releases/tag/v0.1.6"`,
		`v0.1.6`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("layout missing %q", want)
		}
	}
	if got := strings.Count(body, `href="/components"`); got != 2 {
		t.Fatalf("Components family link count = %d, want one desktop and one mobile link", got)
	}
	if got := strings.Count(body, `aria-current="location"`); got != 2 {
		t.Fatalf("active family marker count = %d, want 2", got)
	}
}

func TestLayoutOmitsFamilySurfacesWhenFamiliesEmpty(t *testing.T) {
	t.Parallel()
	body := renderLayout(t, validConfig(), validPage())
	for _, absent := range []string{
		`id="componentdocshell-family-navigation"`,
		`component-doc-shell__family-menu`,
		`component-doc-shell__scope`,
		`component-doc-shell__mobile-utilities`,
	} {
		if strings.Contains(body, absent) {
			t.Errorf("legacy layout unexpectedly contains %q", absent)
		}
	}
}

func TestFragmentRendersAtomicFamilyIdentity(t *testing.T) {
	t.Parallel()
	cfg, page := validFamilyConfig(), validFamilyPage()
	cfg.Interactions.EnableHTMX = true
	body := renderFragment(t, cfg, page)
	for _, target := range []string{
		`outerHTML:#main-content`,
		`outerHTML:#componentdocshell-sidebar-content`,
		`outerHTML:#componentdocshell-family-navigation`,
	} {
		if got := strings.Count(body, target); got != 1 {
			t.Errorf("fragment target %q count = %d, want 1", target, got)
		}
	}
	if !strings.Contains(body, `<title>Line · Reference</title>`) {
		t.Fatal("fragment missing title")
	}
}
```

Use small `renderLayout` and `renderFragment` helpers wrapping templ rendering; keep existing tests intact.

- [ ] **Step 2: Run focused tests and confirm missing markup**

```bash
GOWORK=off go test ./componentdocshell -run 'Test(LayoutRendersFamily|LayoutOmitsFamily|FragmentRendersAtomicFamily)' -count=1
```

Expected: FAIL because family markup does not render.

- [ ] **Step 3: Add stable desktop/mobile control IDs**

Add helpers in `componentdocshell/config.go`:

```go
func (cfg Config) desktopThemeSelectorID() string {
	return cfg.themeSelectorID()
}

func (cfg Config) mobileThemeSelectorID() string {
	return cfg.themeSelectorID() + "-mobile"
}
```

The existing header selector keeps its current ID. Only the new mobile copy receives the suffix.

- [ ] **Step 4: Create semantic family and scope templates**

Create `componentdocshell/family_navigation.templ` with this structure:

```templ
package componentdocshell

templ familyNavigation(cfg Config, page Page, oob bool) {
	if len(cfg.Navigation.Families) > 0 {
		<div id="componentdocshell-family-navigation" class="component-doc-shell__family-navigation" { familyNavigationOOBAttributes(oob)... }>
			<nav class="component-doc-shell__family-links" aria-label="Documentation families">
				for _, family := range familyLinks(cfg, page.ActiveFamily) {
					<a class="component-doc-shell__family-link" href={ templ.SafeURL(family.Href) } { family.LinkAttrs... }>{ family.Label }</a>
				}
			</nav>
			<details class="component-doc-shell__family-menu" data-componentdocshell-family-menu x-on:click.outside="$el.open = false" x-on:keydown.escape.window="$el.open = false; $nextTick(() => $refs.familySummary.focus())">
				<summary x-ref="familySummary" aria-label={ "Current documentation family: " + activeFamilyLabel(cfg, page) }>
					<span>{ activeFamilyLabel(cfg, page) }</span>
					@chevronDownIcon()
				</summary>
				<nav class="component-doc-shell__family-menu-links" aria-label="Documentation families">
					for _, family := range familyLinks(cfg, page.ActiveFamily) {
						<a href={ templ.SafeURL(family.Href) } { family.LinkAttrs... }>{ family.Label }</a>
					}
				</nav>
			</details>
		</div>
	}
}

templ scopeMetadata(cfg Config, page Page) {
	if len(cfg.Navigation.Families) > 0 {
		<section class="component-doc-shell__scope" aria-label="Current documentation family">
			<p class="component-doc-shell__scope-family">{ activeFamilyLabel(cfg, page) }</p>
			if cfg.Navigation.Scope != nil && cfg.Navigation.Scope.ModulePath != "" {
				<code class="component-doc-shell__scope-module">{ cfg.Navigation.Scope.ModulePath }</code>
			}
			if cfg.Navigation.Scope != nil && cfg.Navigation.Scope.Version != "" {
				if cfg.Navigation.Scope.VersionURL != "" {
					<a class="component-doc-shell__scope-version" href={ templ.SafeURL(cfg.Navigation.Scope.VersionURL) }>{ cfg.Navigation.Scope.Version }</a>
				} else {
					<span class="component-doc-shell__scope-version">{ cfg.Navigation.Scope.Version }</span>
				}
			}
		</section>
	}
}
```

Add `chevronDownIcon()` in the same file. Keep both navigations as ordinary anchors; Alpine attributes only enhance disclosure closure.

- [ ] **Step 5: Integrate one family wrapper and scoped sidebar**

In `componentdocshell/layout.templ`:

- Add `data-family-navigation={ boolText(len(cfg.Navigation.Families) > 0) }` to `<body>`.
- Keep brand group first, render `@familyNavigation(cfg, page, false)` second, and controls third inside `.component-doc-shell__header-inner`.
- Wrap the existing header theme selector in `.component-doc-shell__header-theme` and keep `cfg.desktopThemeSelectorID()`.
- Change `sidebarContent` to accept `(cfg Config, page Page, nav sidebar.Config, oob bool)`.
- Render `@scopeMetadata(cfg, page)` before `@sidebar.Sidebar(nav)`.
- Render `.component-doc-shell__mobile-utilities` after the sidebar only when families exist. Put a second theme selector with `cfg.mobileThemeSelectorID()` and a labelled repository link there. Do not render `HeaderActions` in this region.
- Pass `cfg` and `page` at every `sidebarContent` call.

Use this exact utility component so disabled or absent controls do not leave empty destinations:

```templ
templ mobileUtilities(cfg Config) {
	if len(cfg.Navigation.Families) > 0 && (!cfg.Appearance.DisableThemeSelector || cfg.RepositoryURL != "") {
		<div class="component-doc-shell__mobile-utilities">
			if !cfg.Appearance.DisableThemeSelector {
				@selectfield.Select(selectfield.Config{
					ID: cfg.mobileThemeSelectorID(),
					Name: "theme-mobile",
					Options: cfg.themes(),
					Alpine: &selectfield.AlpineConfig{Model: "theme"},
					TriggerAttrs: templ.Attributes{"aria-label": "Theme"},
				})
			}
			if cfg.RepositoryURL != "" {
				<a class="component-doc-shell__mobile-repository" href={ templ.SafeURL(cfg.RepositoryURL) } target="_blank" rel="noreferrer">Source repository</a>
			}
		</div>
	}
}
```

In `componentdocshell/fragment.templ`, append `@familyNavigation(cfg, page, true)` after scoped sidebar replacement.

- [ ] **Step 6: Generate templ output and run semantic tests**

```bash
templ generate
gofmt -w componentdocshell/config.go componentdocshell/componentdocshell_test.go
GOWORK=off go test ./componentdocshell -run 'Test(LayoutRendersFamily|LayoutOmitsFamily|FragmentRendersAtomicFamily|FragmentRendersMain)' -count=1
```

Expected: PASS. Inspect generated changes; hand-written `.templ` files remain source authority.

- [ ] **Step 7: Commit semantic rendering**

```bash
git add componentdocshell/config.go componentdocshell/family_navigation.templ componentdocshell/family_navigation_templ.go componentdocshell/layout.templ componentdocshell/layout_templ.go componentdocshell/fragment.templ componentdocshell/fragment_templ.go componentdocshell/componentdocshell_test.go
git commit -m "feat: render component docs family navigation"
```

### Task 4: Implement deterministic responsive layout

**Files:**
- Modify: `componentdocshell/assets/shell.css`
- Modify: `componentdocshell/assets/assets_test.go`

**Interfaces:**
- Consumes: template classes and `body[data-family-navigation]` from Task 3.
- Produces: `--component-doc-shell-header-height`, small/medium/wide family layouts, drawer utilities, and no-gap frame/sidebar/backdrop/TOC offsets.

- [ ] **Step 1: Write failing CSS contract tests**

Add to `componentdocshell/assets/assets_test.go`:

```go
func TestShellStylesDefineFamilyNavigationBreakpoints(t *testing.T) {
	t.Parallel()
	body := servedAsset(t, "/componentdocshell/assets/shell.css")
	for _, want := range []string{
		`--component-doc-shell-header-height: 4rem`,
		`height: calc(100vh - var(--component-doc-shell-header-height))`,
		`inset: var(--component-doc-shell-header-height) auto 0 0`,
		`inset: var(--component-doc-shell-header-height) 0 0`,
		`top: var(--component-doc-shell-header-height)`,
		`@media (min-width: 720px) and (max-width: 1199px)`,
		`--component-doc-shell-header-height: 6.75rem`,
		`@media (min-width: 1200px)`,
		`.component-doc-shell__family-menu`,
		`.component-doc-shell__family-links`,
		`.component-doc-shell__mobile-utilities`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("shell stylesheet missing family layout contract %q", want)
		}
	}
}
```

Extract the existing recorder setup into `servedAsset(t, path) string` so all asset tests can reuse it.

- [ ] **Step 2: Run focused test and confirm failure**

```bash
GOWORK=off go test ./componentdocshell/assets -run TestShellStylesDefineFamilyNavigationBreakpoints -count=1
```

Expected: FAIL with missing header-height and family selectors.

- [ ] **Step 3: Replace hard-coded header offsets with one variable**

In `componentdocshell/assets/shell.css`, define the variable on `.component-doc-shell` and use it for every shell boundary:

```css
.component-doc-shell {
  --component-doc-shell-header-height: 4rem;
  height: 100vh;
  overflow: hidden;
}

.component-doc-shell__header {
  height: var(--component-doc-shell-header-height);
}

.component-doc-shell__frame {
  height: calc(100vh - var(--component-doc-shell-header-height));
}

.component-doc-shell__sidebar {
  inset: var(--component-doc-shell-header-height) auto 0 0;
}

.component-doc-shell__backdrop {
  inset: var(--component-doc-shell-header-height) 0 0;
}

.component-doc-shell__toc-inner {
  top: var(--component-doc-shell-header-height);
}
```

Retain `100vh` for current compatibility; browser tests cover actual gaps and overflow.

- [ ] **Step 4: Add small, medium, and wide family layouts**

Keep the existing flex header rules for empty-family consumers. Add this family-scoped grid and visibility contract:

```css
.component-doc-shell[data-family-navigation="true"] .component-doc-shell__header-inner {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  height: 4rem;
}

.component-doc-shell[data-family-navigation="true"] .component-doc-shell__brand-group {
  grid-column: 1;
  grid-row: 1;
}

.component-doc-shell[data-family-navigation="true"] .component-doc-shell__controls {
  grid-column: 3;
  grid-row: 1;
}

.component-doc-shell__family-navigation {
  min-width: 0;
}

.component-doc-shell__family-links {
  display: none;
}

.component-doc-shell__family-menu {
  position: relative;
  min-width: 0;
}

.component-doc-shell__mobile-utilities {
  display: none;
}

@media (width < 720px) {
  .component-doc-shell[data-family-navigation="true"] .component-doc-shell__header-theme,
  .component-doc-shell[data-family-navigation="true"] .component-doc-shell__repository {
    display: none;
  }

  .component-doc-shell[data-family-navigation="true"] .component-doc-shell__mobile-utilities {
    display: grid;
  }
}

@media (min-width: 720px) and (max-width: 1199px) {
  .component-doc-shell[data-family-navigation="true"] {
    --component-doc-shell-header-height: 6.75rem;
  }

  .component-doc-shell[data-family-navigation="true"] .component-doc-shell__header-inner {
    grid-template-columns: minmax(0, 1fr) auto;
    grid-template-rows: 4rem 2.75rem;
    height: 6.75rem;
  }

  .component-doc-shell[data-family-navigation="true"] .component-doc-shell__family-navigation {
    grid-column: 1 / -1;
    grid-row: 2;
  }

  .component-doc-shell[data-family-navigation="true"] .component-doc-shell__brand-group {
    grid-column: 1;
    grid-row: 1;
  }

  .component-doc-shell[data-family-navigation="true"] .component-doc-shell__controls {
    grid-column: 2;
    grid-row: 1;
  }
}

@media (min-width: 720px) {
  .component-doc-shell__family-links {
    display: flex;
  }

  .component-doc-shell__family-menu {
    display: none;
  }
}

@media (min-width: 1200px) {
  .component-doc-shell[data-family-navigation="true"] .component-doc-shell__header-inner {
    grid-template-columns: auto minmax(0, 1fr) auto;
    height: 4rem;
  }

  .component-doc-shell[data-family-navigation="true"] .component-doc-shell__family-navigation {
    grid-column: 2;
    grid-row: 1;
  }

  .component-doc-shell[data-family-navigation="true"] .component-doc-shell__controls {
    grid-column: 3;
    grid-row: 1;
  }
}
```

Use these exact interaction and overflow properties, then add dark selectors by substituting existing `*-dark` semantic tokens:

```css
.component-doc-shell__family-links {
  align-items: center;
  gap: 0.25rem;
  overflow: hidden;
  white-space: nowrap;
}

.component-doc-shell__family-link,
.component-doc-shell__family-menu-links a {
  border-radius: var(--radius);
  color: var(--color-on-surface-muted);
  font-size: 0.875rem;
  font-weight: 600;
  text-decoration: none;
}

.component-doc-shell__family-link {
  padding: 0.5rem 0.75rem;
}

.component-doc-shell__family-link[aria-current="location"],
.component-doc-shell__family-menu-links a[aria-current="location"] {
  background: var(--color-surface-alt);
  color: var(--color-on-surface-strong);
}

.component-doc-shell__family-link:focus-visible,
.component-doc-shell__family-menu summary:focus-visible,
.component-doc-shell__family-menu-links a:focus-visible,
.component-doc-shell__mobile-repository:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}

.component-doc-shell__family-menu summary {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  cursor: pointer;
  list-style: none;
  overflow: hidden;
  padding: 0.5rem 0.75rem;
}

.component-doc-shell__family-menu summary span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.component-doc-shell__family-menu-links {
  position: absolute;
  top: calc(100% + 0.5rem);
  left: 0;
  z-index: 60;
  display: grid;
  min-width: 12rem;
  border: 1px solid var(--color-outline);
  border-radius: var(--radius-radius);
  background: var(--color-surface);
  padding: 0.5rem;
  box-shadow: 0 1rem 2rem rgb(0 0 0 / 0.16);
}

.component-doc-shell__family-menu-links a,
.component-doc-shell__mobile-repository {
  padding: 0.625rem 0.75rem;
}

.component-doc-shell__mobile-utilities {
  gap: 0.75rem;
  border-top: 1px solid var(--color-outline);
  padding: 1rem 2rem;
}
```

Family links stay on one line at medium/wide widths. Empty-family header remains on the original flex path.

- [ ] **Step 5: Run asset and package tests**

```bash
gofmt -w componentdocshell/assets/assets_test.go
GOWORK=off go test ./componentdocshell/assets ./componentdocshell -count=1
```

Expected: PASS. `assets.go` derives cache versions from embedded bytes at startup, so it does not change.

- [ ] **Step 6: Commit responsive styles**

```bash
git add componentdocshell/assets/shell.css componentdocshell/assets/assets_test.go
git commit -m "feat: add responsive family header layouts"
```

### Task 5: Close navigation surfaces through the shell lifecycle

**Files:**
- Modify: `componentdocshell/assets/shell.js`
- Modify: `componentdocshell/assets/assets_test.go`
- Modify: `componentdocshell/layout.templ`
- Modify: `componentdocshell/layout_templ.go`

**Interfaces:**
- Consumes: existing `htmx:afterSwap`, `componentdocshell:navigated`, `sidebarOpen`, and `[data-componentdocshell-family-menu]`.
- Produces: `closeFamilyMenu()`, explicit sidebar/family closure after main swaps, preserved scroll reset, TOC rebuild, and main-heading focus.

- [ ] **Step 1: Write failing runtime contract test**

```go
func TestShellRuntimeClosesFamilyMenuAfterNavigation(t *testing.T) {
	t.Parallel()
	body := servedAsset(t, "/componentdocshell/assets/shell.js")
	for _, want := range []string{
		`function closeFamilyMenu()`,
		`[data-componentdocshell-family-menu]`,
		`menu.open = false`,
		`closeFamilyMenu();`,
		`window.dispatchEvent(new CustomEvent("componentdocshell:navigated"))`,
		`focusMain();`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("shell runtime missing family lifecycle contract %q", want)
		}
	}
}
```

- [ ] **Step 2: Run focused test and confirm failure**

```bash
GOWORK=off go test ./componentdocshell/assets -run TestShellRuntimeClosesFamilyMenuAfterNavigation -count=1
```

Expected: FAIL with missing `closeFamilyMenu`.

- [ ] **Step 3: Add menu closure without replacing anchor semantics**

Add to `shell.js`:

```js
function closeFamilyMenu() {
  var menu = document.querySelector("[data-componentdocshell-family-menu]");
  if (menu) menu.open = false;
}
```

Call `closeFamilyMenu()` inside the existing main-target `htmx:afterSwap` branch before dispatching `componentdocshell:navigated`. Export it beside `buildTOC` and `focusMain` for deterministic browser assertions. Keep the existing scroll reset, TOC rebuild, and focus order.

Update the root Alpine event in `layout.templ` to close only `sidebarOpen`; native family disclosure is owned by `closeFamilyMenu()` and its inline Escape/outside-click enhancement.

- [ ] **Step 4: Generate templates and run runtime tests**

```bash
templ generate
gofmt -w componentdocshell/assets/assets_test.go
GOWORK=off go test ./componentdocshell/assets ./componentdocshell -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit navigation lifecycle**

```bash
git add componentdocshell/assets/shell.js componentdocshell/assets/assets_test.go componentdocshell/layout.templ componentdocshell/layout_templ.go
git commit -m "feat: synchronize family navigation lifecycle"
```

### Task 6: Expand the standalone example to six family roots

**Files:**
- Modify: `example/internal/pages/pages.go`
- Modify: `example/internal/pages/pages.templ`
- Modify: `example/internal/pages/pages_templ.go`
- Modify: `example/internal/server/server.go`
- Modify: `example/internal/server/server_test.go`

**Interfaces:**
- Consumes: public family-navigation API from Tasks 1–5.
- Produces: `ShellConfig(activeFamily string) componentdocshell.Config`, `FamilyOverview(familyID string) (componentdocshell.Page, bool)`, six family roots, local runtime assets, and family-aware fragment responses.

- [ ] **Step 1: Write failing example route tests**

Replace the two-page-only route expectations with all family roots:

```go
func TestFamilyRoutesRenderScopedShells(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		path   string
		family string
	}{
		{"/components", "Components"},
		{"/charts", "Charts"},
		{"/app-shells", "App Shells"},
		{"/icons", "Icons"},
		{"/llms", "LLMs"},
		{"/examples", "Examples"},
	} {
		t.Run(test.family, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))
			if recorder.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want 200", test.path, recorder.Code)
			}
			body := recorder.Body.String()
			if !strings.Contains(body, `aria-current="location">`+test.family+`</a>`) {
				t.Errorf("GET %s missing active family %q", test.path, test.family)
			}
			if !strings.Contains(body, `<h1` ) || !strings.Contains(body, test.family) {
				t.Errorf("GET %s missing family overview content", test.path)
			}
		})
	}
}

func TestFamilyHTMXResponseIsAtomic(t *testing.T) {
	t.Parallel()
	request := httptest.NewRequest(http.MethodGet, "/charts", nil)
	request.Header.Set("HX-Request", "true")
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, request)
	body := recorder.Body.String()
	for _, target := range []string{"#main-content", "#componentdocshell-sidebar-content", "#componentdocshell-family-navigation"} {
		if got := strings.Count(body, "outerHTML:"+target); got != 1 {
			t.Errorf("HTMX target %s count = %d, want 1", target, got)
		}
	}
}
```

- [ ] **Step 2: Run focused tests and confirm missing routes**

```bash
GOWORK=off go test ./example/internal/server -run 'TestFamily' -count=1
```

Expected: FAIL because `/charts` and other family roots return 404.

- [ ] **Step 3: Define exact family fixture data**

In `example/internal/pages/pages.go`, define the ordered links once:

```go
var families = []componentdocshell.FamilyLink{
	{ID: "components", Label: "Components", Href: "/components"},
	{ID: "charts", Label: "Charts", Href: "/charts"},
	{ID: "app-shells", Label: "App Shells", Href: "/app-shells"},
	{ID: "icons", Label: "Icons", Href: "/icons"},
	{ID: "llms", Label: "LLMs", Href: "/llms"},
	{ID: "examples", Label: "Examples", Href: "/examples"},
}

type familyFixture struct {
	ID         string
	Label      string
	ModulePath string
	Version    string
	VersionURL string
}
```

Create this exact six-entry map; the App Shell checkout omits a version because the family-navigation commit is unreleased:

```go
var familyFixtures = map[string]familyFixture{
	"components": {ID: "components", Label: "Components", ModulePath: "github.com/araihu/goshtoso", Version: "v0.1.6", VersionURL: "https://github.com/araihu/goshtoso/releases/tag/v0.1.6"},
	"charts": {ID: "charts", Label: "Charts", ModulePath: "github.com/araihu/goshtoso-charts", Version: "v0.0.1", VersionURL: "https://github.com/araihu/goshtoso-charts/releases/tag/v0.0.1"},
	"app-shells": {ID: "app-shells", Label: "App Shells", ModulePath: "github.com/araihu/goshtoso-app-shells"},
	"icons": {ID: "icons", Label: "Icons"},
	"llms": {ID: "llms", Label: "LLMs"},
	"examples": {ID: "examples", Label: "Examples"},
}
```

Return fresh family and sidebar slices from `ShellConfig` so tests can detect mutation. Set `Interactions.LocalRuntime = true` for deterministic E2E assets.

- [ ] **Step 4: Add family pages and routes**

Implement:

```go
func FamilyOverview(familyID string) (componentdocshell.Page, bool) {
	fixture, ok := familyFixtures[familyID]
	if !ok {
		return componentdocshell.Page{}, false
	}
	return componentdocshell.Page{
		Title:        fixture.Label,
		Description:  fixture.Label + " documentation family overview.",
		ActiveFamily: fixture.ID,
		Active:       "overview",
		Content:      familyOverviewContent(fixture.Label),
	}, true
}
```

Add `familyOverviewContent(label string)` in `pages.templ`. Register the six exact paths explicitly in `server.New`; do not use prefix matching. Change `render` to call `pages.ShellConfig(page.ActiveFamily)`. Preserve `/`, `/components/button`, assets, and fixture routes; make `/` render the Components-scoped overview instead of redirecting.

- [ ] **Step 5: Generate and run example tests**

```bash
templ generate
gofmt -w example/internal/pages/pages.go example/internal/server/server.go example/internal/server/server_test.go
GOWORK=off go test ./example/internal/pages ./example/internal/server -count=1
```

Expected: PASS, including existing presentation-channel and asset tests.

- [ ] **Step 6: Commit the six-family example**

```bash
git add example/internal/pages/pages.go example/internal/pages/pages.templ example/internal/pages/pages_templ.go example/internal/server/server.go example/internal/server/server_test.go
git commit -m "feat: demonstrate six documentation families"
```

### Task 7: Add browser acceptance for responsive and navigation behavior

**Files:**
- Create: `example/e2e/family_navigation_test.go`
- Modify: `go.mod`
- Modify: `go.sum`
- Modify: `.github/workflows/ci.yml`

**Interfaces:**
- Consumes: `example/internal/server.New()`, local Goshtoso runtime assets, six family routes, and rendered data hooks.
- Produces: `TestFamilyNavigationVisualMatrix`, `TestFamilyNavigationHTMXHistoryAndFocus`, `TestFamilyNavigationWithoutJavaScript`, and a CI browser gate.

- [ ] **Step 1: Add the pinned Playwright-Go test dependency**

```bash
GOWORK=off go get github.com/playwright-community/playwright-go@v0.5700.1
```

Expected: `go.mod` and `go.sum` record Playwright-Go plus its transitive test runtime.

- [ ] **Step 2: Write the browser harness and responsive matrix**

Create `example/e2e/family_navigation_test.go`. Skip unless `COMPONENTDOCSHELL_E2E=1`, serve `server.New()` through `httptest.NewServer`, and run Chromium. The matrix must use exact widths and three approved themes:

```go
var familyWidths = []int{390, 719, 720, 841, 1199, 1200, 1280, 1440}
var familyThemes = []string{"araihu", "goshtoso", "minimal"}

func requireE2E(t *testing.T) {
	t.Helper()
	if os.Getenv("COMPONENTDOCSHELL_E2E") != "1" {
		t.Skip("set COMPONENTDOCSHELL_E2E=1 to run browser acceptance")
	}
}
```

For each width, theme, and light/dark state:

- Seed `theme` and `darkMode` with `AddInitScript`.
- Visit `/components` and wait for `#main-content`.
- Record page errors and `console.error` messages.
- Evaluate computed header height (`64` except `108` at 720–1199), body/document horizontal overflow, sidebar/backdrop top bounds, family label clipping, visible family surface, desktop/mobile theme-selector visibility, and active `aria-current` values.
- Assert small widths show disclosure and local menu trigger; medium/wide widths show inline families; sidebar becomes persistent at 720px; only one theme selector is visible.
- Assert all configured controls have a visible labelled access path.

Use one returned metrics object with boolean fields and report the full object on failure. This follows the existing AraiHu browser-test pattern and avoids screenshot-only verdicts.

- [ ] **Step 3: Add interaction, history, focus, and failure-path tests**

Implement three focused flows. Use `page.SetViewportSize(841, 900)`, click `a[href="/charts"]` from the visible inline nav, wait for URL `**/charts`, and assert this identity object after navigation, `GoBack`, and `GoForward`:

```js
() => ({
  title: document.title,
  heading: document.querySelector("#main-content h1")?.textContent.trim(),
  family: document.querySelector('.component-doc-shell__family-links [aria-current="location"]')?.textContent.trim(),
  scope: document.querySelector(".component-doc-shell__scope-family")?.textContent.trim(),
  focus: document.activeElement?.textContent.trim(),
  path: location.pathname,
})
```

Expected tuples are Components/`/components`, Charts/`/charts`, Components/`/components`, then Charts/`/charts`; `title`, `heading`, `family`, `scope`, and focused heading must match the active label.

At 390px, click `.component-doc-shell__family-menu summary`, assert the `<details>` has `open`, press Escape, then assert `open == false` and the summary equals `document.activeElement`. Open `.component-doc-shell__menu-button`, click the visible local `/components/button` link, and assert `.component-doc-shell__sidebar` lacks `is-open` after `#main-content` becomes Button.

Create a JavaScript-disabled context with:

```go
context, err := browser.NewContext(playwright.BrowserNewContextOptions{
	JavaScriptEnabled: playwright.Bool(false),
})
```

Load `/components`, assert the family disclosure contains the six exact hrefs, click its native summary to open it, click the `/charts` anchor, and assert a full document with Charts title, heading, scope, and active family.

Add a throwing-storage subtest with this init script before navigation:

```js
Storage.prototype.getItem = function () { throw new Error("storage disabled"); };
Storage.prototype.setItem = function () { throw new Error("storage disabled"); };
```

Assert page content, family links, theme control, navigation, and captured console/page errors remain healthy.

- [ ] **Step 4: Install Chromium and run browser acceptance**

```bash
go run github.com/playwright-community/playwright-go/cmd/playwright@v0.5700.1 install chromium
COMPONENTDOCSHELL_E2E=1 GOWORK=off go test ./example/e2e -count=1 -v
```

Expected: PASS for all widths, themes, color schemes, interaction flows, history, no-JS, and throwing storage.

- [ ] **Step 5: Add a dedicated CI browser job**

Extend `.github/workflows/ci.yml` with a `browser` job using the same Go version and templ version as `verify`. Install Chromium with system dependencies, generate templates, and run only the E2E package:

```yaml
      - name: Install Chromium
        run: go run github.com/playwright-community/playwright-go/cmd/playwright@v0.5700.1 install --with-deps chromium
      - name: Test component docs shell browser behavior
        env:
          COMPONENTDOCSHELL_E2E: "1"
        run: GOWORK=off go test ./example/e2e -count=1 -v
```

- [ ] **Step 6: Run normal tests to prove E2E skips cleanly without browser setup**

```bash
GOWORK=off go test ./... -count=1
```

Expected: PASS; `example/e2e` reports a skip unless the opt-in environment variable is set.

- [ ] **Step 7: Commit browser acceptance**

```bash
git add go.mod go.sum example/e2e/family_navigation_test.go .github/workflows/ci.yml
git commit -m "test: cover component docs family navigation"
```

### Task 8: Document the API and prove an unrelated consumer

**Files:**
- Modify: `README.md`
- Create: `testdata/external-consumer/go.mod`
- Create: `testdata/external-consumer/main_test.go`
- Create: `scripts/test-componentdocshell-external-consumer.sh`

**Interfaces:**
- Consumes: only exported `componentdocshell` and `componentdocshell/assets` APIs.
- Produces: copyable README configuration and `scripts/test-componentdocshell-external-consumer.sh` clean-module gate.

- [ ] **Step 1: Add a failing unrelated-module fixture**

Use keyed literals for every exported public struct in this fixture. Public-model
changes preserve behavior and zero values for keyed literals, but adding exported
fields is not source-compatible with external positional literals.

Create `testdata/external-consumer/go.mod`:

```go
module example.com/componentdocshell-consumer

go 1.26.5

require github.com/araihu/goshtoso-app-shells v0.0.0
```

Create `testdata/external-consumer/main_test.go` using only public imports:

```go
package consumer_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-app-shells/componentdocshell"
	"github.com/araihu/goshtoso/components/sidebar"
)

func TestPublicFamilyNavigationAPI(t *testing.T) {
	cfg := componentdocshell.Config{
		Brand: componentdocshell.Brand{Name: "External docs", HomeURL: "/"},
		Navigation: componentdocshell.Navigation{
			Families: []componentdocshell.FamilyLink{
				{ID: "components", Label: "Components", Href: "/components"},
				{ID: "charts", Label: "Charts", Href: "/charts"},
			},
			Scope: &componentdocshell.ScopeMetadata{ModulePath: "example.com/widgets", Version: "v1.2.3"},
			Items: []sidebar.Item{{ID: "overview", Label: "Overview", Href: "/components"}},
		},
		Interactions: componentdocshell.InteractionConfig{EnableHTMX: true},
	}
	page := componentdocshell.Page{
		Title: "Components", ActiveFamily: "components", Active: "overview", Content: templ.NopComponent,
	}
	var output bytes.Buffer
	if err := componentdocshell.Layout(cfg, page).Render(context.Background(), &output); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`aria-current="location"`, `example.com/widgets`, `v1.2.3`} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("external render missing %q", want)
		}
	}
}
```

- [ ] **Step 2: Add the clean temporary-module runner**

Create `scripts/test-componentdocshell-external-consumer.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail

repo_root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
fixture_source="$repo_root/testdata/external-consumer"
fixture_work="$(mktemp -d "${TMPDIR:-/tmp}/componentdocshell-consumer.XXXXXX")"
trap 'rm -rf "$fixture_work"' EXIT

cp "$fixture_source/go.mod" "$fixture_source/main_test.go" "$fixture_work/"
cd "$fixture_work"
GOWORK=off go mod edit -replace="github.com/araihu/goshtoso-app-shells=$repo_root"
GOWORK=off go mod tidy
GOWORK=off go test ./... -count=1
```

Mark it executable. Cleanup targets only the validated `mktemp` directory.

- [ ] **Step 3: Run external proof**

```bash
chmod +x scripts/test-componentdocshell-external-consumer.sh
./scripts/test-componentdocshell-external-consumer.sh
```

Expected: PASS from module `example.com/componentdocshell-consumer` with `GOWORK=off`; no `internal` package import succeeds or appears.

- [ ] **Step 4: Document family configuration and delivery boundary**

Add README sections showing:

```go
Navigation: componentdocshell.Navigation{
	Families: []componentdocshell.FamilyLink{
		{ID: "components", Label: "Components", Href: "/components"},
		{ID: "charts", Label: "Charts", Href: "/charts"},
		{ID: "app-shells", Label: "App Shells", Href: "/app-shells"},
		{ID: "icons", Label: "Icons", Href: "/icons"},
		{ID: "llms", Label: "LLMs", Href: "/llms"},
		{ID: "examples", Label: "Examples", Href: "/examples"},
	},
	Scope: &componentdocshell.ScopeMetadata{
		ModulePath: "github.com/araihu/goshtoso",
		Version: "v0.1.6",
		VersionURL: "https://github.com/araihu/goshtoso/releases/tag/v0.1.6",
	},
},
```

Explain `Page.ActiveFamily`, overview-route switching, empty-family compatibility, responsive breakpoints, ordinary-link fallback, HTMX OOB identity, responsive `HeaderActions` ownership, example routes, browser command, and external-consumer command. State that public-model changes are additive only for behavior, zero values, and keyed literals: adding exported fields is not source-compatible with external positional literals, and every supported fixture/example must use keyed literals. State that Goshtoso adoption and full Charts/App Shells content need separate release and contribution work.

- [ ] **Step 5: Run documentation and consumer checks**

```bash
./scripts/test-componentdocshell-external-consumer.sh
GOWORK=off go test ./... -count=1
git diff --check
```

Expected: all commands PASS.

- [ ] **Step 6: Commit docs and consumer proof**

```bash
git add README.md testdata/external-consumer/go.mod testdata/external-consumer/main_test.go scripts/test-componentdocshell-external-consumer.sh
git commit -m "docs: explain component docs family navigation"
```

### Task 9: Run final package gates and freeze review identity

**Files:**
- Verify only: all changed files from Tasks 1–8

**Interfaces:**
- Consumes: completed implementation commits.
- Produces: generated-drift proof, root test/vet/build proof, browser proof, external-consumer proof, and exact review identity. No merge, release, or cleanup.

- [ ] **Step 1: Regenerate and prove no generated drift**

```bash
templ generate
git diff --exit-code
```

Expected: PASS with no diff. If generated files move, inspect and commit the source plus generated correction before continuing.

- [ ] **Step 2: Run full Go gates outside any workspace**

```bash
GOWORK=off go mod tidy -diff
GOWORK=off go test ./... -count=1
GOWORK=off go test -race ./... -count=1
GOWORK=off go vet ./...
GOWORK=off go build ./...
```

Expected: all commands PASS.

- [ ] **Step 3: Run browser and external-consumer gates**

```bash
COMPONENTDOCSHELL_E2E=1 GOWORK=off go test ./example/e2e -count=1 -v
./scripts/test-componentdocshell-external-consumer.sh
```

Expected: PASS with no page errors, console errors, identity drift, overflow, clipping, control loss, breakpoint gap, or public-package violation.

- [ ] **Step 4: Run source-quality checks**

```bash
git diff --check origin/main..HEAD
git diff origin/main..HEAD -- README.md componentdocshell example testdata scripts .github | rg -n 'TBD|TODO|FIXME|(^|[^X])XXX([^X]|$)|implement later|fill in details|similar to Task'
git status --short --branch
```

Expected: diff check PASS; placeholder scan returns no new implementation placeholders; worktree contains no uncommitted files.

- [ ] **Step 5: Record immutable review identity**

```bash
git rev-parse HEAD
git rev-parse HEAD^{tree}
git log --oneline --decorate origin/main..HEAD
git diff --stat origin/main..HEAD
```

Record commit SHA, tree SHA, exact test commands, and results in the review request. Stop here. Merge, tag, release, Goshtoso dependency bump, deployment, and worktree removal remain separate approval gates.

## Deferred Consumer Project

After an authorized App Shell release exists, write a separate Goshtoso adoption plan from then-current fetched `origin/main`. That plan must use the released module version, add the six real overview routes and family-scoped sidebars, group global search results by family, move Attributions/License/Privacy to the structured footer, remove Goshtoso's header-wide version badge, run root and `site/` tests, run the Goshtoso accessibility scan and full visual matrix, and prove standalone pinned-dependency deployment with `GOWORK=off`. Full Charts/App Shells catalogs remain a third project governed by a versioned producer contribution contract.
