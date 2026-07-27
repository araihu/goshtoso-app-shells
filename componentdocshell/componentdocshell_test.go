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
		`github.com/araihu/reference`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("layout missing %q", want)
		}
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
