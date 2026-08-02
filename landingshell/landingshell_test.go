package landingshell

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func TestLayoutRendersLandingContracts(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	page := validPage()
	var buffer bytes.Buffer
	if err := Layout(cfg, page).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	body := buffer.String()
	for _, want := range []string{
		`data-theme="araihu"`,
		`x-data="landingShell(`,
		`id="landingshell-dark-mode"`,
		`aria-label="Switch to dark mode"`,
		`aria-label="Source repository"`,
		`aria-label="Product v1 release"`,
		`aria-label="Footer navigation"`,
		`an <a href="https://araihu.com"`,
		`<section class="landing-shell__hero">`,
		`id="product-hero"`,
		`id="product-content"`,
		`property="og:image" content="https://example.com/og.png"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("layout missing %q", want)
		}
	}
	if strings.Index(body, `id="product-hero"`) > strings.Index(body, `id="product-content"`) {
		t.Fatal("hero must render before page content")
	}
}

func TestLayoutSupportsStorageFreeAndLocalRuntimeModes(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Appearance.PersistPreferences = false
	cfg.Interactions.LocalRuntime = true
	var buffer bytes.Buffer
	if err := Layout(cfg, validPage()).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	body := buffer.String()
	if !strings.Contains(shellData(cfg), `"persist":false`) {
		t.Fatal("shell options do not disable persistence")
	}
	if strings.Contains(body, `/assets/js/dependency-loader.js`) {
		t.Fatal("local runtime layout contains dependency loader")
	}
	if !strings.Contains(body, `/assets/js/runtime/alpinejs/`) {
		t.Fatal("local runtime layout does not include local Alpine")
	}
}

func TestLayoutValidation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		edit func(*Config, *Page)
		want string
	}{
		{name: "brand", edit: func(cfg *Config, _ *Page) { cfg.Brand.Name = "" }, want: "brand name"},
		{name: "hero", edit: func(_ *Config, page *Page) { page.Hero = nil }, want: "hero is required"},
		{name: "content", edit: func(_ *Config, page *Page) { page.Content = nil }, want: "content is required"},
		{name: "scheme", edit: func(cfg *Config, _ *Page) { cfg.Appearance.InitialColorScheme = "sepia" }, want: "initial color scheme"},
		{name: "prefix", edit: func(cfg *Config, _ *Page) { cfg.AssetPrefix = "assets" }, want: "asset prefix"},
		{name: "nav", edit: func(cfg *Config, _ *Page) { cfg.Navigation[0].Href = "" }, want: "navigation link href"},
		{name: "organization", edit: func(cfg *Config, _ *Page) { cfg.Footer.Organization.URL = "" }, want: "organization name and URL"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := validConfig()
			page := validPage()
			test.edit(&cfg, &page)
			var buffer bytes.Buffer
			err := Layout(cfg, page).Render(context.Background(), &buffer)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Layout().Render() error = %v, want containing %q", err, test.want)
			}
		})
	}
}

func validConfig() Config {
	return Config{
		Brand: Brand{
			Name: "Product", HomeURL: "/", Tagline: "Useful software",
			Logo:  templ.Raw(`<svg data-logo></svg>`),
			Badge: &BrandBadge{Label: "v1", AriaLabel: "Product v1 release", Href: "/releases/v1"},
		},
		Navigation: []Link{
			{Label: "Docs", Href: "/docs"},
			{Label: "Get started", Href: "/start", Primary: true},
		},
		Appearance: AppearanceConfig{DefaultTheme: "araihu", InitialColorScheme: ColorSchemeSystem, PersistPreferences: true},
		Footer: Footer{
			Meta:         []string{"12 components", "4 themes"},
			Organization: &Organization{Name: "Arai Hû", URL: "https://araihu.com"},
			Links:        []Link{{Label: "Docs", Href: "/docs"}, {Label: "GitHub", Href: "https://github.com/example/product", External: true}},
		},
		RepositoryURL: "https://github.com/example/product",
	}
}

func validPage() Page {
	return Page{
		Title: "Home", Description: "Product description", CanonicalURL: "https://example.com/", SocialImageURL: "https://example.com/og.png",
		Hero:    templ.Raw(`<div id="product-hero"><h1>Product</h1></div>`),
		Content: templ.Raw(`<section id="product-content">Content</section>`),
	}
}
