package componentdocshell

import (
	"context"
	"encoding/json"
	"io"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/sidebar"
)

// Layout returns a complete server-rendered component documentation document.
func Layout(cfg Config, page Page) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, writer io.Writer) error {
		if err := validate(cfg, page, false); err != nil {
			return err
		}
		return layoutTemplate(cfg, page, navigationConfig(cfg, page.Active)).Render(ctx, writer)
	})
}

// Fragment returns HTMX-compatible main content and out-of-band navigation.
func Fragment(cfg Config, page Page) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, writer io.Writer) error {
		if err := validate(cfg, page, true); err != nil {
			return err
		}
		return fragmentTemplate(cfg, page, navigationConfig(cfg, page.Active)).Render(ctx, writer)
	})
}

// Head renders Goshtoso runtime dependencies and component docs shell assets.
func Head(cfg Config) templ.Component {
	return headTemplate(cfg)
}

func navigationConfig(cfg Config, active string) sidebar.Config {
	return sidebar.Config{
		Items:             cloneItems(cfg.Navigation.Items, active, cfg.Interactions.EnableHTMX),
		SectionsTitle:     cfg.Navigation.SectionsTitle,
		Sections:          cloneSections(cfg.Navigation.Sections, active, cfg.Interactions.EnableHTMX),
		ShowSearch:        !cfg.Navigation.DisableSearch,
		SearchPlaceholder: cfg.searchPlaceholder(),
		DisableSkipLink:   true,
	}
}

func cloneSections(sections []sidebar.Section, active string, htmx bool) []sidebar.Section {
	result := make([]sidebar.Section, len(sections))
	for index, section := range sections {
		result[index] = section
		result[index].Items = cloneItems(section.Items, active, htmx)
	}
	return result
}

func cloneItems(items []sidebar.Item, active string, htmx bool) []sidebar.Item {
	result := make([]sidebar.Item, len(items))
	for index, item := range items {
		result[index] = item
		result[index].Active = item.ID == active
		result[index].Items = cloneItems(item.Items, active, htmx)
		if htmx && !item.Disabled && item.Href != "" {
			attrs := templ.Attributes{}
			for key, value := range item.LinkAttrs {
				attrs[key] = value
			}
			attrs["hx-get"] = item.Href
			attrs["hx-target"] = "#main-content"
			attrs["hx-push-url"] = "true"
			result[index].LinkAttrs = attrs
		}
	}
	return result
}

func shellData(cfg Config) string {
	options, _ := json.Marshal(shellOptions{
		Persist:     cfg.Appearance.PersistPreferences,
		Theme:       cfg.defaultTheme(),
		ColorScheme: cfg.initialColorScheme(),
	})
	return "componentDocShell(" + string(options) + ")"
}

type shellOptions struct {
	Persist     bool        `json:"persist"`
	Theme       string      `json:"theme"`
	ColorScheme ColorScheme `json:"colorScheme"`
}

func appearanceBootstrapScript(cfg Config) string {
	options, _ := json.Marshal(shellOptions{
		Persist:     cfg.Appearance.PersistPreferences,
		Theme:       cfg.defaultTheme(),
		ColorScheme: cfg.initialColorScheme(),
	})
	return `(function(o){var theme=o.theme;var dark=o.colorScheme==="dark"||(o.colorScheme==="system"&&window.matchMedia("(prefers-color-scheme: dark)").matches);try{if(o.persist){theme=localStorage.getItem("theme")||theme;var saved=localStorage.getItem("darkMode");if(saved!==null)dark=saved==="true";}}catch(_){}document.documentElement.setAttribute("data-theme",theme);document.documentElement.classList.toggle("dark",dark);})(` + string(options) + `);`
}

func currentPageTitle(cfg Config, page Page) string {
	if page.DocumentTitle != "" {
		return page.DocumentTitle
	}
	return page.Title + " · " + cfg.Brand.Name
}

func activeCurrent(active bool) string {
	if active {
		return "page"
	}
	return ""
}

func sidebarOOBAttributes(enabled bool) templ.Attributes {
	if !enabled {
		return nil
	}
	return templ.Attributes{"hx-swap-oob": "outerHTML:#componentdocshell-sidebar-content"}
}

func mainOOBAttributes(enabled bool) templ.Attributes {
	if !enabled {
		return nil
	}
	return templ.Attributes{"hx-swap-oob": "outerHTML:#main-content"}
}

func boolText(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
