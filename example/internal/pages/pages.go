package pages

import (
	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-app-shells/componentdocshell"
	selectfield "github.com/araihu/goshtoso/components/select"
	"github.com/araihu/goshtoso/components/sidebar"
)

// ShellConfig returns the shared example-site frame configuration.
func ShellConfig() componentdocshell.Config {
	return componentdocshell.Config{
		Brand: componentdocshell.Brand{Name: "Component docs shell example", HomeURL: "/"},
		Navigation: componentdocshell.Navigation{
			Items:         []sidebar.Item{{ID: "overview", Label: "Overview", Href: "/"}},
			SectionsTitle: "Components",
			Sections: []sidebar.Section{{Title: "Components", Items: []sidebar.Item{
				{ID: "button", Label: "Button", Href: "/components/button"},
			}}},
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
		Interactions:  componentdocshell.InteractionConfig{EnableHTMX: true},
		Footer:        templ.Raw(`<p class="text-sm text-on-surface-muted dark:text-on-surface-dark-muted">Built with Goshtoso Component Docs Shell.</p>`),
	}
}

// Overview returns the example landing page.
func Overview() componentdocshell.Page {
	return componentdocshell.Page{
		Title: "Overview", Description: "Reusable Goshtoso application shell example.",
		Active: "overview", Content: overviewContent(),
	}
}

// Button returns a representative component reference page.
func Button() componentdocshell.Page {
	return componentdocshell.Page{
		Title: "Button", Description: "Button component reference example.",
		Active: "button", Content: buttonContent(), EnableTOC: true,
	}
}
