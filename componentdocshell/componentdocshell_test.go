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
	for _, want := range []string{`"persist":true`, `"theme":"minimal"`, `"colorScheme":"dark"`, `document.documentElement.setAttribute("data-theme",theme)`} {
		if !strings.Contains(body, want) {
			t.Errorf("layout appearance bootstrap missing %q", want)
		}
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
