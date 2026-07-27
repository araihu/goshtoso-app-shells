package pages

import (
	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-app-shells/catalogshell"
	selectfield "github.com/araihu/goshtoso/components/select"
	"github.com/araihu/goshtoso/components/sidebar"
)

// ShellConfig returns the shared example-site frame configuration.
func ShellConfig() catalogshell.Config {
	return catalogshell.Config{
		Brand: catalogshell.Brand{Name: "Catalog shell example", HomeURL: "/"},
		Navigation: catalogshell.Navigation{
			Items: []sidebar.Item{{ID: "overview", Label: "Overview", Href: "/"}},
			SectionsTitle: "Catalog",
			Sections: []sidebar.Section{{Title: "Components", Items: []sidebar.Item{
				{ID: "button", Label: "Button", Href: "/components/button"},
			}}},
		},
		Themes: []selectfield.Option{
			{Value: "goshtoso", Label: "Goshtoso", Selected: true},
			{Value: "minimal", Label: "Minimal"},
		},
		RepositoryURL: "https://github.com/araihu/goshtoso-app-shells",
		EnableHTMX: true,
		Footer: templ.Raw(`<p class="text-sm text-on-surface-muted dark:text-on-surface-dark-muted">Built with Goshtoso Catalog Shell.</p>`),
	}
}

// Overview returns the example landing page.
func Overview() catalogshell.Page {
	return catalogshell.Page{
		Title: "Overview", Description: "Reusable Goshtoso application shell example.",
		Active: "overview", Content: overviewContent(),
	}
}

// Button returns a representative component reference page.
func Button() catalogshell.Page {
	return catalogshell.Page{
		Title: "Button", Description: "Button component reference example.",
		Active: "button", Content: buttonContent(), EnableTOC: true,
	}
}
