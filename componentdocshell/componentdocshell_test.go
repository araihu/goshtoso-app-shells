package componentdocshell

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func TestLayoutRendersComponentDocsShellContract(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.RepositoryURL = "https://github.com/araihu/reference"
	cfg.Interactions.EnableHTMX = true
	page := validPage()
	page.Description = "Reference components"
	page.Content = templ.Raw(`<h1 id="line" data-toc-heading>Line</h1><h2 id="usage" data-toc-heading>Usage</h2>`)
	page.EnableTOC = true

	var buffer bytes.Buffer
	if err := Layout(cfg, page).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	body := buffer.String()
	for _, want := range []string{
		"<!doctype html>",
		`<html lang="en" class="component-doc-shell-root"`,
		`href="#main-content"`,
		`class="component-doc-shell__header"`,
		`aria-label="Open navigation"`,
		`aria-controls="componentdocshell-sidebar"`,
		`x-bind:aria-expanded="sidebarOpen"`,
		`aria-label="Theme"`,
		`aria-label="Switch to dark mode"`,
		`Switch to light mode`,
		`x-on:componentdocshell:navigated.window="sidebarOpen = false"`,
		`aria-current="page"`,
		`id="componentdocshell-sidebar-content"`,
		`id="main-content"`,
		`id="componentdocshell-toc"`,
		`/componentdocshell/assets/shell.css`,
		`/componentdocshell/assets/shell.js`,
		`localStorage.getItem("theme")`,
		`github.com/araihu/reference`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("layout missing %q", want)
		}
	}
	if shellIndex := strings.Index(body, `/componentdocshell/assets/shell.js`); shellIndex < 0 || shellIndex > strings.Index(body, `/assets/js/dependency-loader.js`) {
		t.Fatal("shell registration script must run before the Goshtoso dependency loader")
	}
}

func TestLayoutBootstrapsPersistedAppearanceBeforeRuntime(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Appearance.PersistPreferences = true
	cfg.Appearance.DefaultTheme = "minimal"
	cfg.Appearance.InitialColorScheme = ColorSchemeDark
	var buffer bytes.Buffer
	if err := Layout(cfg, validPage()).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	body := buffer.String()
	for _, want := range []string{`"persist":true`, `"persistTheme":true`, `"theme":"minimal"`, `"colorScheme":"dark"`, `document.documentElement.setAttribute("data-theme",theme)`} {
		if !strings.Contains(body, want) {
			t.Errorf("layout appearance bootstrap missing %q", want)
		}
	}
}

func TestLayoutRendersManagedPresentationChannel(t *testing.T) {
	t.Parallel()
	zeroValueHTML := renderValid(t, validConfig())
	for _, hook := range []string{"data-asset-brand", "data-campaign-toggle", "data-use-campaign-label", "data-channel"} {
		if strings.Contains(zeroValueHTML, hook) {
			t.Fatalf("zero-value configuration rendered %q", hook)
		}
	}

	cfg := validConfig()
	cfg.Brand.ManagedLogo = &ManagedBrandAsset{URL: "/assets/brand/logo.svg", Alt: "Reference", Width: 120, Height: 32}
	cfg.Brand.ManageFavicon = true
	cfg.Brand.FaviconURL = "/assets/brand/icon.svg"
	cfg.Interactions.PresentationChannel = &PresentationChannelConfig{
		RuntimeURL:       "/assets/campaign/v1.js",
		ChannelURL:       "/assets/releases/current",
		Integrity:        "sha384-campaign",
		UseCampaignLabel: "Use campaign",
		UseBaselineLabel: "Use baseline",
	}
	html := renderValid(t, cfg)
	for _, want := range []string{
		`data-asset-brand="logo"`, `width="120"`, `height="32"`,
		`data-asset-brand="icon"`, `data-campaign-toggle`, `data-campaign-toggle-icon`,
		`data-use-campaign-label="Use campaign"`, `data-use-baseline-label="Use baseline"`,
		`data-channel="/assets/releases/current"`, `integrity="sha384-campaign"`, `crossorigin="anonymous"`, `defer`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("layout missing %q", want)
		}
	}
	if strings.Count(html, "data-campaign-toggle") != 2 {
		t.Fatalf("campaign toggle hooks = %d, want button and icon only", strings.Count(html, "data-campaign-toggle"))
	}
	for _, want := range []string{`type="button"`, `hidden`, `aria-pressed="false"`, `class="sr-only"`} {
		if !strings.Contains(html, want) {
			t.Errorf("campaign toggle missing %q", want)
		}
	}
	if bootstrap, runtime := strings.Index(html, "dataset.themeSource"), strings.Index(html, `src="/assets/campaign/v1.js"`); bootstrap < 0 || runtime < 0 || bootstrap > runtime {
		t.Fatal("first-paint bootstrap must precede campaign runtime")
	}
	if stylesheet, documentEnd := strings.Index(html, `/componentdocshell/assets/shell.css`), strings.Index(html, "</html>"); stylesheet < 0 || documentEnd < 0 || stylesheet > documentEnd {
		t.Fatal("baseline stylesheet must precede document end")
	}
	if !strings.Contains(html, `class="component-doc-shell__brand-name">Reference</span>`) {
		t.Fatal("managed logo must preserve brand name by default")
	}

	cfg.Brand.HideName = true
	html = renderValid(t, cfg)
	if strings.Contains(html, `class="component-doc-shell__brand-name">Reference</span>`) {
		t.Fatal("hidden managed brand name rendered")
	}
}

func TestAppearanceBootstrapMarksThemeSource(t *testing.T) {
	t.Parallel()
	disabled := appearanceBootstrapScript(validConfig())
	if !strings.Contains(disabled, `"persist":false`) || !strings.Contains(disabled, `"persistTheme":false`) {
		t.Error("disabled persistence bootstrap enables a saved theme")
	}
	cfg := validConfig()
	cfg.Appearance.PersistPreferences = true
	script := appearanceBootstrapScript(cfg)
	for _, want := range []string{
		`var source="default"`,
		`var savedTheme=localStorage.getItem("theme");if(savedTheme){theme=savedTheme;source="preference"}`, // missing or empty key stays default; non-empty key is preference
		`catch(_){}`, // storage exceptions retain configured theme and default source
		`var saved=localStorage.getItem("darkMode");if(saved!==null)dark=saved==="true"`,
		`document.documentElement.dataset.themeSource=source`,
	} {
		if !strings.Contains(script, want) {
			t.Errorf("appearance bootstrap missing %q", want)
		}
	}
}

func TestLayoutLocksThemePersistenceWhenSelectorIsDisabled(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Appearance.PersistPreferences = true
	cfg.Appearance.DefaultTheme = "araihu"
	cfg.Appearance.DisableThemeSelector = true
	var buffer bytes.Buffer
	if err := Layout(cfg, validPage()).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	body := buffer.String()
	for _, want := range []string{`"persist":true`, `"persistTheme":false`, `"theme":"araihu"`} {
		if !strings.Contains(body, want) {
			t.Errorf("locked theme layout missing %q", want)
		}
	}
	if strings.Contains(body, `aria-label="Theme"`) {
		t.Error("locked theme layout rendered theme selector")
	}
}

func TestLayoutCanBindDarkModeControlToApplicationStore(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Appearance.DarkModeBinding = &DarkModeBinding{
		ButtonID:         "darkModeToggleBtn",
		StateExpression:  "$store.darkMode.on",
		ToggleExpression: "$store.darkMode.toggle()",
	}
	var buffer bytes.Buffer
	if err := Layout(cfg, validPage()).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	body := buffer.String()
	for _, want := range []string{
		`id="darkModeToggleBtn"`,
		`x-bind:aria-label="$store.darkMode.on ? &#39;Switch to light mode&#39; : &#39;Switch to dark mode&#39;"`,
		`x-on:click="$store.darkMode.toggle()"`,
		`x-show="!($store.darkMode.on)"`,
		`x-show="$store.darkMode.on"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("layout application dark-mode binding missing %q", want)
		}
	}
}

func TestLayoutCanUseWordmarkWithoutDuplicateBrandName(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Brand.Logo = templ.Raw(`<img src="/wordmark.svg" alt="">`)
	cfg.Brand.HideName = true
	var buffer bytes.Buffer
	if err := Layout(cfg, validPage()).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	body := buffer.String()
	if !strings.Contains(body, `/wordmark.svg`) {
		t.Error("wordmark layout missing logo")
	}
	if strings.Contains(body, `component-doc-shell__brand-name`) {
		t.Error("wordmark layout rendered duplicate brand name")
	}
}

func TestLayoutBrandBadgeIsOptional(t *testing.T) {
	t.Parallel()
	var buffer bytes.Buffer
	if err := Layout(validConfig(), validPage()).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	if strings.Contains(buffer.String(), `component-doc-shell__brand-badge`) {
		t.Fatal("layout rendered an unconfigured brand badge")
	}
}

func TestLayoutRendersLinkedBrandBadge(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Brand.Badge = &BrandBadge{
		Label:     "v1.2.3",
		AriaLabel: "Goshtoso release v1.2.3",
		Href:      "https://github.com/araihu/goshtoso/releases/tag/v1.2.3",
	}
	var buffer bytes.Buffer
	if err := Layout(cfg, validPage()).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	body := buffer.String()
	for _, want := range []string{
		`class="component-doc-shell__brand-badge"`,
		`href="https://github.com/araihu/goshtoso/releases/tag/v1.2.3"`,
		`aria-label="Goshtoso release v1.2.3"`,
		`>v1.2.3</a>`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("linked brand badge missing %q", want)
		}
	}
}

func TestLayoutRendersUnlinkedBrandBadge(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Brand.Badge = &BrandBadge{Label: "dev", AriaLabel: "Development build"}
	var buffer bytes.Buffer
	if err := Layout(cfg, validPage()).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	body := buffer.String()
	if !strings.Contains(body, `<span class="component-doc-shell__brand-badge" aria-label="Development build">dev</span>`) {
		t.Fatalf("unlinked brand badge markup missing: %s", body)
	}
	if strings.Contains(body, `href=""`) {
		t.Fatal("unlinked brand badge rendered an empty link")
	}
}

func TestLayoutCanPreserveApplicationThemeSelectorID(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Appearance.ThemeSelectorID = "site-theme"
	var buffer bytes.Buffer
	if err := Layout(cfg, validPage()).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	body := buffer.String()
	for _, want := range []string{`id="site-theme-trigger"`, `id="site-theme-listbox"`, `name="theme"`} {
		if !strings.Contains(body, want) {
			t.Errorf("layout application theme-selector ID missing %q", want)
		}
	}
}

func TestLayoutCanPreserveApplicationTOCIDs(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.TOC = TOCConfig{RailID: "toc-rail", ListID: "toc-list"}
	page := validPage()
	page.EnableTOC = true
	var buffer bytes.Buffer
	if err := Layout(cfg, page).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	body := buffer.String()
	for _, want := range []string{
		`id="toc-rail" class="component-doc-shell__toc" data-componentdocshell-toc`,
		`id="toc-list" class="component-doc-shell__toc-list" data-componentdocshell-toc-list`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("layout application TOC ID missing %q", want)
		}
	}
}

func TestLayoutCanExposeLocalHTMXBeforeBodyContent(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Interactions.LocalRuntime = true
	var buffer bytes.Buffer
	if err := Layout(cfg, validPage()).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	body := buffer.String()
	if strings.Contains(body, `/assets/js/dependency-loader.js`) {
		t.Fatal("local runtime layout contains dependency loader")
	}
	htmlHTMX := `<script src="/assets/js/runtime/htmx.org/`
	if !strings.Contains(body, htmlHTMX) {
		t.Fatalf("local runtime layout missing eager HTMX script %q", htmlHTMX)
	}
}

func TestLayoutRendersApplicationSearchOverlayAndRuntimeExtensions(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Navigation.SearchSlot = templ.Raw(`<button id="docs-search">Search docs</button>`)
	cfg.BodyEnd = templ.Raw(`<div id="storage-consent">Consent</div>`)
	cfg.Interactions.LocalRuntime = true
	cfg.Interactions.RuntimeScripts = []string{"/assets/js/runtime/htmx-ext-ws.js", "/assets/js/runtime/htmx-ext-sse.js"}
	var buffer bytes.Buffer
	if err := Layout(cfg, validPage()).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	body := buffer.String()
	for _, want := range []string{`id="docs-search"`, `id="storage-consent"`, `src="/assets/js/runtime/htmx-ext-ws.js"`, `src="/assets/js/runtime/htmx-ext-sse.js"`} {
		if !strings.Contains(body, want) {
			t.Errorf("layout missing application extension %q", want)
		}
	}
	if core := strings.Index(body, `/assets/js/runtime/htmx.org/`); core < 0 || core > strings.Index(body, `/assets/js/runtime/htmx-ext-ws.js`) {
		t.Fatal("HTMX core must render before runtime extensions")
	}
}

func TestLayoutDoesNotMutateNavigation(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	if cfg.Navigation.Sections[0].Items[0].Active {
		t.Fatal("test fixture unexpectedly active")
	}
	var buffer bytes.Buffer
	if err := Layout(cfg, validPage()).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	if cfg.Navigation.Sections[0].Items[0].Active {
		t.Fatal("Layout mutated caller-owned navigation")
	}
}

func TestLayoutAllowsExactDocumentTitle(t *testing.T) {
	t.Parallel()
	page := validPage()
	page.DocumentTitle = "Exact established SEO title"
	var buffer bytes.Buffer
	if err := Layout(validConfig(), page).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	if !strings.Contains(buffer.String(), "<title>Exact established SEO title</title>") {
		t.Fatalf("layout did not preserve the exact document title")
	}
}

func TestGoshtosoBrandUsesCanonicalMarkAndFavicon(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Brand = GoshtosoBrand("Goshtoso Charts", "/", "")
	var buffer bytes.Buffer
	if err := Layout(cfg, validPage()).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	for _, want := range []string{"goshtoso-mark.svg", "goshtoso-mark-reverse.svg", "goshtoso-favicon.svg"} {
		if !strings.Contains(buffer.String(), want) {
			t.Errorf("layout missing %q", want)
		}
	}
}

func TestFragmentRendersMainAndOutOfBandSidebar(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Interactions.EnableHTMX = true
	var buffer bytes.Buffer
	if err := Fragment(cfg, validPage()).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Fragment().Render() error = %v", err)
	}
	body := buffer.String()
	if strings.Contains(body, "<html") {
		t.Fatal("fragment contains complete document")
	}
	for _, want := range []string{`<title>Line · Reference</title>`, `id="main-content"`, `hx-swap-oob="outerHTML:#componentdocshell-sidebar-content"`, `aria-current="page"`} {
		if !strings.Contains(body, want) {
			t.Errorf("fragment missing %q", want)
		}
	}
}

func TestLayoutReportsValidationErrorAtRender(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Brand.Name = ""
	var buffer bytes.Buffer
	err := Layout(cfg, validPage()).Render(context.Background(), &buffer)
	if err == nil || !strings.Contains(err.Error(), "brand name is required") {
		t.Fatalf("Layout().Render() error = %v", err)
	}
}
