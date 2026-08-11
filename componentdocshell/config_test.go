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

func withHTMX(cfg Config) Config {
	cfg.Interactions.EnableHTMX = true
	return cfg
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
		cfg.Brand.FaviconURL = "/assets/brand/favicon.svg"
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
		{"missing managed favicon URL", func(cfg *Config) { cfg.Brand.FaviconURL = "" }},
		{"invalid managed favicon URL", func(cfg *Config) { cfg.Brand.FaviconURL = "http://assets.example/favicon.svg" }},
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
	cfg.Brand.FaviconURL = "https://assets.example/favicon.svg"
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
		{"padded ID", func(cfg *Config, _ *Page) { cfg.Navigation.Families[0].ID = " components " }, "family ID must not have leading or trailing whitespace"},
		{"duplicate ID", func(cfg *Config, _ *Page) { cfg.Navigation.Families[1].ID = "components" }, `duplicate family ID "components"`},
		{"empty label", func(cfg *Config, _ *Page) { cfg.Navigation.Families[0].Label = " " }, `family "components" label is required`},
		{"empty URL", func(cfg *Config, _ *Page) { cfg.Navigation.Families[0].Href = "" }, `family "components" URL is required`},
		{"HTTP URL", func(cfg *Config, _ *Page) { cfg.Navigation.Families[0].Href = "http://docs.example/components" }, "must be root-relative or an absolute HTTPS URL"},
		{"scheme-relative URL", func(cfg *Config, _ *Page) { cfg.Navigation.Families[0].Href = "//docs.example/components" }, "must be root-relative or an absolute HTTPS URL"},
		{"link attribute ID", func(cfg *Config, _ *Page) {
			cfg.Navigation.Families[0].LinkAttrs = templ.Attributes{"ID": "consumer-family"}
		}, `family "components" link attributes must not contain id`},
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

func TestFamilyValidationFailsBeforeWritingBytes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		edit func(*Config)
	}{
		{
			name: "invalid URL",
			edit: func(cfg *Config) {
				cfg.Navigation.Families[0].Href = "javascript:alert(1)"
			},
		},
		{
			name: "padded family ID",
			edit: func(cfg *Config) {
				cfg.Navigation.Families[0].ID = " components "
			},
		},
		{
			name: "consumer link ID",
			edit: func(cfg *Config) {
				cfg.Navigation.Families[0].LinkAttrs = templ.Attributes{"id": "consumer-family"}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg, page := validFamilyConfig(), validFamilyPage()
			test.edit(&cfg)
			for _, component := range []templ.Component{Layout(cfg, page), Fragment(withHTMX(cfg), page)} {
				var buffer bytes.Buffer
				if err := component.Render(context.Background(), &buffer); err == nil {
					t.Fatal("Render() accepted invalid family configuration")
				}
				if buffer.Len() != 0 {
					t.Fatalf("Render() wrote %d bytes before validation failure", buffer.Len())
				}
			}
		})
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
