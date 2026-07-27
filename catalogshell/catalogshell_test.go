package catalogshell

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func TestLayoutRendersCatalogShellContract(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.RepositoryURL = "https://github.com/araihu/reference"
	cfg.EnableHTMX = true
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
		`class="catalog-shell__header"`,
		`aria-label="Open navigation"`,
		`aria-label="Theme"`,
		`aria-label="Toggle dark mode"`,
		`aria-current="page"`,
		`id="catalogshell-sidebar-content"`,
		`id="main-content"`,
		`id="catalogshell-toc"`,
		`/catalogshell/assets/shell.css`,
		`/catalogshell/assets/shell.js`,
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

func TestFragmentRendersMainAndOutOfBandSidebar(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.EnableHTMX = true
	var buffer bytes.Buffer
	if err := Fragment(cfg, validPage()).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Fragment().Render() error = %v", err)
	}
	body := buffer.String()
	if strings.Contains(body, "<html") {
		t.Fatal("fragment contains complete document")
	}
	for _, want := range []string{`id="main-content"`, `hx-swap-oob="outerHTML:#catalogshell-sidebar-content"`, `aria-current="page"`} {
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
