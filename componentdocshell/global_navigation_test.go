package componentdocshell

import (
	"bytes"
	"context"
	"github.com/a-h/templ"
	"strings"
	"testing"
)

func TestGlobalNavigationComposition(t *testing.T) {
	for _, local := range []bool{false, true} {
		for _, global := range []bool{false, true} {
			cfg := validConfig()
			cfg.Navigation.DisableLocalNavigation = !local
			if global {
				cfg.Navigation.GlobalLinks = []GlobalLink{{Label: "Blog", Href: "/blog/"}, {Label: "External", Href: "https://example.com/#contact"}}
			}
			for _, fragment := range []bool{false, true} {
				var out bytes.Buffer
				cfg.Interactions.EnableHTMX = true
				component := Layout(cfg, validPage())
				if fragment {
					component = Fragment(cfg, validPage())
				}
				if err := component.Render(context.Background(), &out); err != nil {
					t.Fatal(err)
				}
				html := out.String()
				wantLocal := `data-local-navigation="false"`
				if local {
					wantLocal = `data-local-navigation="true"`
				}
				if !strings.Contains(html, wantLocal) {
					t.Fatalf("local=%v global=%v fragment=%v: missing state", local, global, fragment)
				}
				if strings.Contains(html, `component-doc-shell__sidebar-nav`) != (local) {
					t.Fatalf("local=%v: wrong contextual navigation", local)
				}
				want := 0
				if global {
					want = 1
					if !fragment {
						want = 2
					}
				}
				if got := strings.Count(html, `href="/blog/"`); got != want {
					t.Fatalf("global=%v fragment=%v: links=%d want=%d", global, fragment, got, want)
				}
				if strings.Contains(html, `hx-get="https://example.com/#contact"`) || strings.Contains(html, `hx-get="/blog/"`) {
					t.Fatal("global links acquired implicit HTMX")
				}
				if fragment && strings.Count(html, `hx-swap-oob="outerHTML:#componentdocshell-sidebar-content"`) != 1 {
					t.Fatal("fragment lost composed navigation")
				}
			}
		}
	}
}

func TestInvalidGlobalNavigationWritesNothing(t *testing.T) {
	for _, test := range []struct {
		name   string
		modify func(*Config)
	}{
		{"unsafe URL", func(c *Config) { c.Navigation.GlobalLinks[0].Href = "javascript:alert(1)" }},
		{"credentials", func(c *Config) { c.Navigation.GlobalLinks[0].Href = "https://user:password@example.com/" }},
		{"protocol relative", func(c *Config) { c.Navigation.GlobalLinks[0].Href = "//example.com/" }},
		{"blank label", func(c *Config) { c.Navigation.GlobalLinks[0].Label = " " }},
		{"duplicate DOM ID", func(c *Config) { c.Navigation.GlobalLinks[0].LinkAttrs = templ.Attributes{"ID": "duplicate"} }},
		{"legacy slot conflict", func(c *Config) { c.HeaderNavigation = templ.NopComponent }},
	} {
		t.Run(test.name, func(t *testing.T) {
			cfg := validConfig()
			cfg.Navigation.GlobalLinks = []GlobalLink{{Label: "Blog", Href: "/blog/"}}
			test.modify(&cfg)
			var out bytes.Buffer
			if err := Layout(cfg, validPage()).Render(context.Background(), &out); err == nil || out.Len() != 0 {
				t.Fatalf("error=%v bytes=%d", err, out.Len())
			}
		})
	}
}
