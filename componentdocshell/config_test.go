package componentdocshell

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/sidebar"
)

func validConfig() Config {
	return Config{
		Brand: Brand{Name: "Reference", HomeURL: "/"},
		Navigation: Navigation{
			Items: []sidebar.Item{{ID: "overview", Label: "Overview", Href: "/"}},
			Sections: []sidebar.Section{{Title: "Components", Items: []sidebar.Item{
				{ID: "line", Label: "Line", Href: "/components/line"},
			}}},
		},
	}
}

func validPage() Page {
	return Page{Title: "Line", Active: "line", Content: templ.NopComponent}
}

func renderValid(t *testing.T, cfg Config) string {
	t.Helper()
	var buffer bytes.Buffer
	if err := Layout(cfg, validPage()).Render(context.Background(), &buffer); err != nil {
		t.Fatal(err)
	}
	return buffer.String()
}

func TestPresentationChannelIsDisabledByDefault(t *testing.T) {
	t.Parallel()
	html := renderValid(t, validConfig())
	if strings.Contains(html, "data-campaign-toggle") || strings.Contains(html, "/assets/campaign/") {
		t.Fatal("zero-value config enrolled a runtime")
	}
}

func TestValidatePresentationChannel(t *testing.T) {
	t.Parallel()
	valid := func() Config {
		cfg := validConfig()
		cfg.Brand.ManageFavicon = true
		cfg.Interactions.PresentationChannel = &PresentationChannelConfig{
			RuntimeURL:       "/assets/campaign/v1.js",
			ChannelURL:       "/assets/campaign/channel.json",
			Integrity:        "sha384-campaign",
			UseCampaignLabel: "Use campaign",
			UseBaselineLabel: "Use baseline",
		}
		return cfg
	}

	if err := validate(valid(), validPage(), false); err != nil {
		t.Fatalf("validate() valid root-relative presentation channel error = %v", err)
	}

	tests := []struct {
		name string
		edit func(*Config)
	}{
		{"missing channel", func(cfg *Config) { cfg.Interactions.PresentationChannel.ChannelURL = "" }},
		{"bad integrity", func(cfg *Config) { cfg.Interactions.PresentationChannel.Integrity = "sha256-campaign" }},
		{"empty integrity digest", func(cfg *Config) { cfg.Interactions.PresentationChannel.Integrity = "sha384-" }},
		{"missing campaign label", func(cfg *Config) { cfg.Interactions.PresentationChannel.UseCampaignLabel = "" }},
		{"missing baseline label", func(cfg *Config) { cfg.Interactions.PresentationChannel.UseBaselineLabel = "" }},
		{"no managed asset", func(cfg *Config) { cfg.Brand.ManageFavicon = false }},
		{"conflicting logos", func(cfg *Config) {
			cfg.Brand.Logo = templ.NopComponent
			cfg.Brand.ManagedLogo = &ManagedBrandAsset{URL: "/brand.svg", Width: 1, Height: 1}
		}},
		{"zero logo width", func(cfg *Config) {
			cfg.Brand.ManageFavicon = false
			cfg.Brand.ManagedLogo = &ManagedBrandAsset{URL: "/brand.svg", Width: 0, Height: 1}
		}},
		{"mixed origins", func(cfg *Config) {
			cfg.Interactions.PresentationChannel.ChannelURL = "https://assets.example/campaign.json"
		}},
		{"mismatched origins", func(cfg *Config) {
			cfg.Interactions.PresentationChannel.RuntimeURL = "https://assets.example/campaign.js"
			cfg.Interactions.PresentationChannel.ChannelURL = "https://cdn.example/campaign.json"
		}},
		{"http origin", func(cfg *Config) {
			cfg.Interactions.PresentationChannel.RuntimeURL = "http://assets.example/campaign.js"
			cfg.Interactions.PresentationChannel.ChannelURL = "http://assets.example/campaign.json"
		}},
		{"credentials", func(cfg *Config) {
			cfg.Interactions.PresentationChannel.RuntimeURL = "https://user@assets.example/campaign.js"
			cfg.Interactions.PresentationChannel.ChannelURL = "https://user@assets.example/campaign.json"
		}},
		{"fragment", func(cfg *Config) {
			cfg.Interactions.PresentationChannel.ChannelURL = "/assets/campaign/channel.json#v1"
		}},
		{"scheme-relative path", func(cfg *Config) {
			cfg.Interactions.PresentationChannel.RuntimeURL = "///assets.example/campaign.js"
		}},
		{"backslash", func(cfg *Config) {
			cfg.Interactions.PresentationChannel.RuntimeURL = `/\\assets.example/campaign.js`
		}},
		{"control character", func(cfg *Config) { cfg.Interactions.PresentationChannel.UseCampaignLabel = "Use\ncampaign" }},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			cfg := valid()
			test.edit(&cfg)
			if err := validate(cfg, validPage(), false); err == nil {
				t.Fatal("validate accepted invalid presentation channel")
			}
		})
	}
}

func TestValidateAllowsSameHTTPSOriginWithDefaultPort(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Brand.ManageFavicon = true
	cfg.Interactions.PresentationChannel = &PresentationChannelConfig{
		RuntimeURL:       "https://assets.example:443/campaign.js",
		ChannelURL:       "https://assets.example/channel.json",
		Integrity:        "sha384-campaign",
		UseCampaignLabel: "Use campaign",
		UseBaselineLabel: "Use baseline",
	}
	if err := validate(cfg, validPage(), false); err != nil {
		t.Fatalf("validate() equivalent HTTPS origins error = %v", err)
	}
}

func TestValidateAllowsHTTPSPresentationChannelWithManagedLogo(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Brand.ManagedLogo = &ManagedBrandAsset{URL: "https://assets.example/brand.svg", Alt: "Reference", Width: 120, Height: 32}
	cfg.Interactions.PresentationChannel = &PresentationChannelConfig{
		RuntimeURL:       "https://assets.example/campaign.js",
		ChannelURL:       "https://assets.example/channel.json",
		Integrity:        "sha384-campaign",
		UseCampaignLabel: "Use campaign",
		UseBaselineLabel: "Use baseline",
	}
	if err := validate(cfg, validPage(), false); err != nil {
		t.Fatalf("validate() HTTPS presentation channel error = %v", err)
	}
}

func TestValidateAcceptsMinimalComponentDocsSite(t *testing.T) {
	t.Parallel()
	if err := validate(validConfig(), validPage(), false); err != nil {
		t.Fatalf("validate() error = %v", err)
	}
}

func TestValidateRejectsMissingRequiredFields(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		edit func(*Config, *Page)
		want string
	}{
		{"brand name", func(cfg *Config, _ *Page) { cfg.Brand.Name = "" }, "brand name is required"},
		{"home URL", func(cfg *Config, _ *Page) { cfg.Brand.HomeURL = "" }, "brand home URL is required"},
		{"page title", func(_ *Config, page *Page) { page.Title = "" }, "page title is required"},
		{"content", func(_ *Config, page *Page) { page.Content = nil }, "page content is required"},
		{"asset prefix", func(cfg *Config, _ *Page) { cfg.AssetPrefix = "relative" }, "asset prefix must start and end with /"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			cfg, page := validConfig(), validPage()
			test.edit(&cfg, &page)
			err := validate(cfg, page, false)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("validate() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestValidateRejectsDuplicateNavigationIDs(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Navigation.Sections[0].Items = append(cfg.Navigation.Sections[0].Items,
		sidebar.Item{ID: "line", Label: "Again", Href: "/again"})
	err := validate(cfg, validPage(), false)
	if err == nil || !strings.Contains(err.Error(), `duplicate navigation ID "line"`) {
		t.Fatalf("validate() error = %v", err)
	}
}

func TestValidateRejectsUnknownActiveNavigation(t *testing.T) {
	t.Parallel()
	page := validPage()
	page.Active = "missing"
	err := validate(validConfig(), page, false)
	if err == nil || !strings.Contains(err.Error(), `active navigation ID "missing" is not configured`) {
		t.Fatalf("validate() error = %v", err)
	}
}

func TestValidateRejectsFragmentWhenHTMXDisabled(t *testing.T) {
	t.Parallel()
	err := validate(validConfig(), validPage(), true)
	if err == nil || !strings.Contains(err.Error(), "fragment rendering requires HTMX") {
		t.Fatalf("validate() error = %v", err)
	}
}

func TestValidateRejectsRuntimeScriptsWithoutLocalRuntime(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Interactions.RuntimeScripts = []string{"/extension.js"}
	err := validate(cfg, validPage(), false)
	if err == nil || !strings.Contains(err.Error(), "runtime scripts require local runtime") {
		t.Fatalf("validate() error = %v", err)
	}
}

func TestDefaultAppearanceIncludesEveryGoshtosoThemeAndSelectsAraiHu(t *testing.T) {
	t.Parallel()
	options := validConfig().themes()
	if len(options) != 16 {
		t.Fatalf("default themes = %d, want Arai Hû plus 15 Goshtoso themes", len(options))
	}
	if options[0].Value != "araihu" || !options[0].Selected {
		t.Fatalf("default theme = %#v, want selected Arai Hû", options[0])
	}
}

func TestValidateAllowsLockedCustomTheme(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Appearance.DefaultTheme = "customer-theme"
	cfg.Appearance.DisableThemeSelector = true
	cfg.Appearance.DisableDarkModeToggle = true
	cfg.Appearance.DisableDefaultThemeStylesheet = true
	cfg.Appearance.ThemeStylesheets = []string{"/customer-theme.css"}
	if err := validate(cfg, validPage(), false); err != nil {
		t.Fatalf("validate() error = %v", err)
	}
}

func TestValidateRejectsUnavailableVisibleDefaultTheme(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Appearance.DefaultTheme = "missing"
	err := validate(cfg, validPage(), false)
	if err == nil || !strings.Contains(err.Error(), `default theme "missing" is not available`) {
		t.Fatalf("validate() error = %v", err)
	}
}
