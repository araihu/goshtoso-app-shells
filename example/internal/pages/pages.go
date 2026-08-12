package pages

import (
	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-app-shells/componentdocshell"
	selectfield "github.com/araihu/goshtoso/components/select"
	"github.com/araihu/goshtoso/components/sidebar"
)

const (
	fixtureRuntimeIntegrity = "sha384-wCfS7WktEJUyXXUJVX+Hd2vWoWqU3XX2oZeezp7QLgRz8hoI9iBLdBKhPGBfddvH"
	productionOrigin        = "https://goshtoso.araihu.com"
	socialImageURL          = productionOrigin + "/assets/images/goshtoso-social-card.png"
)

var families = []componentdocshell.FamilyLink{
	{ID: "components", Label: "Components", Href: "/components"},
	{ID: "charts", Label: "Charts", Href: "/charts"},
	{ID: "app-shells", Label: "App Shells", Href: "/app-shells"},
	{ID: "icons", Label: "Icons", Href: "/icons"},
	{ID: "llms", Label: "LLMs", Href: "/llms"},
	{ID: "examples", Label: "Examples", Href: "/examples"},
}

type familyFixture struct {
	ID          string
	Label       string
	ModulePath  string
	ModuleLabel string
	ModuleURL   string
	Version     string
	VersionURL  string
}

var familyFixtures = map[string]familyFixture{
	"components": {ID: "components", Label: "Components", ModulePath: "github.com/araihu/goshtoso", ModuleLabel: "araihu/goshtoso", ModuleURL: "https://github.com/araihu/goshtoso", Version: "v0.1.6", VersionURL: "https://github.com/araihu/goshtoso/releases/tag/v0.1.6"},
	"charts":     {ID: "charts", Label: "Charts", ModulePath: "github.com/araihu/goshtoso-charts", ModuleLabel: "araihu/goshtoso-charts", ModuleURL: "https://github.com/araihu/goshtoso-charts", Version: "v0.0.1", VersionURL: "https://github.com/araihu/goshtoso-charts/releases/tag/v0.0.1"},
	"app-shells": {ID: "app-shells", Label: "App Shells", ModulePath: "github.com/araihu/goshtoso-app-shells", ModuleLabel: "araihu/goshtoso-app-shells", ModuleURL: "https://github.com/araihu/goshtoso-app-shells"},
	"icons":      {ID: "icons", Label: "Icons"},
	"llms":       {ID: "llms", Label: "LLMs"},
	"examples":   {ID: "examples", Label: "Examples"},
}

// ShellConfig returns the shared example-site frame configuration.
func ShellConfig(activeFamily string) componentdocshell.Config {
	fixture := familyFixtures[activeFamily]
	brand := componentdocshell.GoshtosoBrand("Component docs shell example", "/", "")
	brand.CompactLogo = brand.Logo
	brand.ManageFavicon = true
	items := []sidebar.Item{{ID: "overview", Label: "Overview", Href: "/" + fixture.ID}}
	sections := []sidebar.Section{}
	if activeFamily == "components" {
		sections = []sidebar.Section{{
			Title: "Component catalog",
			Items: []sidebar.Item{{ID: "button", Label: "Button", Href: "/components/button"}},
		}}
	}

	return componentdocshell.Config{
		Brand: brand,
		Navigation: componentdocshell.Navigation{
			Families: append([]componentdocshell.FamilyLink(nil), families...),
			Scope:    &componentdocshell.ScopeMetadata{ModulePath: fixture.ModulePath, ModuleLabel: fixture.ModuleLabel, ModuleURL: fixture.ModuleURL, Version: fixture.Version, VersionURL: fixture.VersionURL},
			Items:    items,
			Sections: sections,
		},
		Appearance: componentdocshell.AppearanceConfig{
			Themes: []selectfield.Option{
				{Value: "araihu", Label: "Arai Hû"},
				{Value: "goshtoso", Label: "Goshtoso"},
				{Value: "minimal", Label: "Minimal"},
			},
			DefaultTheme:         "araihu",
			DisableThemeSelector: true,
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
		Title:         fixture.Label,
		DocumentTitle: fixture.Label + " Documentation - Goshtoso",
		Description:   fixture.Label + " documentation family overview.",
		CanonicalURL:  productionOrigin + "/" + fixture.ID,
		SiteName:      "Goshtoso",
		Locale:        "en_US",
		SocialImage:   socialImage(fixture.Label),
		ActiveFamily:  fixture.ID,
		Active:        "overview",
		Content:       familyOverviewContent(fixture.Label),
	}, true
}

// Button returns a representative component reference page.
func Button() componentdocshell.Page {
	return componentdocshell.Page{
		Title:         "Button",
		DocumentTitle: "Button Component - Goshtoso UI Library for Go",
		Description:   "Build Go button components with variants, accessible states, and HTMX-ready server rendering.",
		CanonicalURL:  productionOrigin + "/components/button",
		SiteName:      "Goshtoso",
		Locale:        "en_US",
		SocialImage:   socialImage("Button"),
		ActiveFamily:  "components", Active: "button", Content: buttonContent(), EnableTOC: true,
	}
}

func socialImage(title string) componentdocshell.SocialImage {
	return componentdocshell.SocialImage{
		URL:      socialImageURL,
		MIMEType: "image/png",
		Width:    1200,
		Height:   630,
		Alt:      title + " — Goshtoso Go UI component library preview",
	}
}
