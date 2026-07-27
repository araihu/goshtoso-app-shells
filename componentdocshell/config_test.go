package componentdocshell

import (
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
