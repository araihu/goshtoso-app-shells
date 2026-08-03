package landingshell

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func TestMobileNavigationComposesTopDrawerAndNativeFallback(t *testing.T) {
	t.Parallel()
	cfg := MobileNavigationConfig{
		ID:              "storm-navigation",
		Title:           "Navigation",
		TriggerLabel:    "Menu",
		NavigationLabel: "Primary navigation",
		Position:        FloatingBottomRight,
		RootClass:       "storm-menu",
		TriggerClass:    "storm-trigger",
		PanelClass:      "storm-panel",
	}
	links := templ.Raw(`<a href="/en/" aria-current="page">EN</a>`)

	var buffer bytes.Buffer
	if err := MobileNavigation(cfg, links).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("MobileNavigation().Render() error = %v", err)
	}
	body := buffer.String()
	for _, want := range []string{
		`class="landing-shell__mobile-navigation storm-menu"`,
		`class="landing-shell__mobile-trigger is-bottom-right storm-trigger"`,
		`aria-haspopup="dialog"`,
		`aria-controls="storm-navigation-body"`,
		`x-bind:aria-expanded="navigationOpen"`,
		`x-on:drawer:open.window`,
		`x-on:drawer:close.window`,
		`x-on:drawer:close-request.window`,
		`drawer:open`,
		`left-0 right-0 top-0 border-b`,
		`storm-panel`,
		`class="landing-shell__mobile-fallback"`,
		`<summary class="landing-shell__mobile-fallback-trigger is-bottom-right storm-trigger">Menu</summary>`,
		`aria-label="Primary navigation"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("mobile navigation missing %q:\n%s", want, body)
		}
	}
	if got := strings.Count(body, `href="/en/" aria-current="page"`); got != 2 {
		t.Fatalf("navigation slot render count = %d; want enhanced and native fallback", got)
	}
}

func TestMobileNavigationRejectsUnsafeIDAndMissingContent(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name    string
		cfg     MobileNavigationConfig
		content templ.Component
		want    string
	}{
		{name: "unsafe ID", cfg: MobileNavigationConfig{ID: "menu'"}, content: templ.Raw("links"), want: "ID"},
		{name: "missing content", cfg: MobileNavigationConfig{}, want: "content"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var buffer bytes.Buffer
			err := MobileNavigation(test.cfg, test.content).Render(context.Background(), &buffer)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("MobileNavigation().Render() error = %v; want %q", err, test.want)
			}
		})
	}
}

func TestLayoutOptInComposesMobileNavigationFromPrimaryLinks(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.MobileNavigation = &MobileNavigationConfig{TriggerLabel: "Menu", Title: "Navigation"}

	var buffer bytes.Buffer
	if err := Layout(cfg, validPage()).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	body := buffer.String()
	if !strings.Contains(body, `landing-shell__header--mobile-navigation`) {
		t.Fatal("opt-in layout does not mark the responsive header")
	}
	for _, item := range cfg.Navigation {
		want := 3
		for _, footerItem := range cfg.Footer.Links {
			if footerItem.Href == item.Href {
				want++
			}
		}
		if got := strings.Count(body, `href="`+item.Href+`"`); got != want {
			t.Errorf("navigation link %q render count = %d; want %d including desktop, enhanced, fallback, and footer", item.Href, got, want)
		}
	}
}

func TestLayoutRejectsInvalidMobileNavigationPosition(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.MobileNavigation = &MobileNavigationConfig{Position: "middle"}
	var buffer bytes.Buffer
	err := Layout(cfg, validPage()).Render(context.Background(), &buffer)
	if err == nil || !strings.Contains(err.Error(), "mobile navigation position") {
		t.Fatalf("Layout().Render() error = %v; want mobile navigation position", err)
	}
}
