package pages

import (
	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-app-shells/componentdocshell"
	selectfield "github.com/araihu/goshtoso/components/select"
	"github.com/araihu/goshtoso/components/sidebar"
)

const fixtureRuntimeIntegrity = "sha384-wCfS7WktEJUyXXUJVX+Hd2vWoWqU3XX2oZeezp7QLgRz8hoI9iBLdBKhPGBfddvH"

var families = []componentdocshell.FamilyLink{
	{ID: "components", Label: "Components", Href: "/components"},
	{ID: "charts", Label: "Charts", Href: "/charts"},
	{ID: "app-shells", Label: "App Shells", Href: "/app-shells"},
	{ID: "icons", Label: "Icons", Href: "/icons"},
	{ID: "llms", Label: "LLMs", Href: "/llms"},
	{ID: "examples", Label: "Examples", Href: "/examples"},
}

type familyFixture struct {
	ID         string
	Label      string
	ModulePath string
	Version    string
	VersionURL string
}

var familyFixtures = map[string]familyFixture{
	"components": {ID: "components", Label: "Components", ModulePath: "github.com/araihu/goshtoso", Version: "v0.1.6", VersionURL: "https://github.com/araihu/goshtoso/releases/tag/v0.1.6"},
	"charts":     {ID: "charts", Label: "Charts", ModulePath: "github.com/araihu/goshtoso-charts", Version: "v0.0.1", VersionURL: "https://github.com/araihu/goshtoso-charts/releases/tag/v0.0.1"},
	"app-shells": {ID: "app-shells", Label: "App Shells", ModulePath: "github.com/araihu/goshtoso-app-shells"},
	"icons":      {ID: "icons", Label: "Icons"},
	"llms":       {ID: "llms", Label: "LLMs"},
	"examples":   {ID: "examples", Label: "Examples"},
}

// ShellConfig returns the shared example-site frame configuration.
func ShellConfig(activeFamily string) componentdocshell.Config {
	fixture := familyFixtures[activeFamily]
	items := []sidebar.Item{{ID: "overview", Label: "Overview", Href: "/" + fixture.ID}}
	sections := []sidebar.Section{}
	if activeFamily == "components" {
		sections = []sidebar.Section{{
			Title: "Components",
			Items: []sidebar.Item{{ID: "button", Label: "Button", Href: "/components/button"}},
		}}
	}

	return componentdocshell.Config{
		Brand: componentdocshell.Brand{
			Name:    "Component docs shell example",
			HomeURL: "/",
			ManagedLogo: &componentdocshell.ManagedBrandAsset{
				URL: "/fixtures/brand/logo.svg", Alt: "Component docs shell", Width: 120, Height: 32,
			},
		},
		Navigation: componentdocshell.Navigation{
			Families:      append([]componentdocshell.FamilyLink(nil), families...),
			Scope:         &componentdocshell.ScopeMetadata{ModulePath: fixture.ModulePath, Version: fixture.Version, VersionURL: fixture.VersionURL},
			Items:         items,
			SectionsTitle: fixture.Label,
			Sections:      sections,
		},
		Appearance: componentdocshell.AppearanceConfig{
			Themes: []selectfield.Option{
				{Value: "araihu", Label: "Arai Hû"},
				{Value: "goshtoso", Label: "Goshtoso"},
				{Value: "minimal", Label: "Minimal"},
			},
			DefaultTheme: "araihu",
		},
		RepositoryURL: "https://github.com/araihu/goshtoso-app-shells",
		Interactions: componentdocshell.InteractionConfig{
			EnableHTMX:   true,
			LocalRuntime: true,
			PresentationChannel: &componentdocshell.PresentationChannelConfig{
				RuntimeURL:       "/fixtures/campaign/v1.js",
				ChannelURL:       "/fixtures/releases/current",
				Integrity:        fixtureRuntimeIntegrity,
				UseCampaignLabel: "Use seasonal appearance",
				UseBaselineLabel: "Use standard appearance",
			},
		},
		Footer: templ.Raw(`<p class="text-sm text-on-surface-muted dark:text-on-surface-dark-muted">Built with Goshtoso Component Docs Shell.</p>`),
	}
}

// Overview returns the example landing page.
func Overview() componentdocshell.Page {
	page, _ := FamilyOverview("components")
	return page
}

// FamilyOverview returns the overview page for a documentation family.
func FamilyOverview(familyID string) (componentdocshell.Page, bool) {
	fixture, ok := familyFixtures[familyID]
	if !ok {
		return componentdocshell.Page{}, false
	}
	return componentdocshell.Page{
		Title:        fixture.Label,
		Description:  fixture.Label + " documentation family overview.",
		ActiveFamily: fixture.ID,
		Active:       "overview",
		Content:      familyOverviewContent(fixture.Label),
	}, true
}

// Button returns a representative component reference page.
func Button() componentdocshell.Page {
	return componentdocshell.Page{
		Title: "Button", Description: "Button component reference example.",
		ActiveFamily: "components", Active: "button", Content: buttonContent(), EnableTOC: true,
	}
}
