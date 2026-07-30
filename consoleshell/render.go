package consoleshell

import (
	"context"
	"encoding/json"
	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/sidebar"
	"io"
)

// Layout renders a complete server-side console document.
func Layout(cfg Config, page Page) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if err := validate(cfg, page, false); err != nil {
			return err
		}
		return layoutTemplate(cfg, page, navigationConfig(cfg, page.Active)).Render(ctx, w)
	})
}

// Fragment renders the stable main region and optional OOB navigation update.
func Fragment(cfg Config, page Page) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if err := validate(cfg, page, true); err != nil {
			return err
		}
		return fragmentTemplate(cfg, page, navigationConfig(cfg, page.Active)).Render(ctx, w)
	})
}

// Head renders shell runtime and stylesheet dependencies for custom layouts.
func Head(cfg Config) templ.Component { return headTemplate(cfg) }

func navigationConfig(cfg Config, active string) sidebar.Config {
	return sidebar.Config{Items: cloneItems(cfg.Navigation.Items, active, cfg), SectionsTitle: cfg.Navigation.SectionsTitle, Sections: cloneSections(cfg.Navigation.Sections, active, cfg), ShowSearch: !cfg.Navigation.DisableSearch, SearchPlaceholder: cfg.searchPlaceholder(), SearchSlot: cfg.Navigation.SearchSlot, DisableSkipLink: true}
}
func cloneSections(in []sidebar.Section, active string, cfg Config) []sidebar.Section {
	out := make([]sidebar.Section, len(in))
	for i, section := range in {
		out[i] = section
		out[i].Items = cloneItems(section.Items, active, cfg)
	}
	return out
}
func cloneItems(in []sidebar.Item, active string, cfg Config) []sidebar.Item {
	out := make([]sidebar.Item, len(in))
	for i, item := range in {
		out[i] = item
		out[i].Active = item.ID == active
		out[i].Items = cloneItems(item.Items, active, cfg)
		attrs := templ.Attributes{}
		for k, v := range item.LinkAttrs {
			attrs[k] = v
		}
		attrs["data-consoleshell-nav-id"] = item.ID
		out[i].LinkAttrs = attrs
		if cfg.Interactions.EnableHTMX && !item.Disabled && item.Href != "" {
			attrs["hx-get"] = item.Href
			attrs["hx-target"] = cfg.fragmentTarget()
			attrs["hx-swap"] = "outerHTML"
			attrs["hx-push-url"] = "true"
		}
	}
	return out
}

type shellOptions struct {
	Persist     bool        `json:"persist"`
	Theme       string      `json:"theme"`
	ColorScheme ColorScheme `json:"colorScheme"`
	MainID      string      `json:"mainID"`
	ContentID   string      `json:"contentID"`
}

func shellData(cfg Config) string {
	b, _ := json.Marshal(shellOptions{cfg.Appearance.PersistPreferences, cfg.defaultTheme(), cfg.initialColorScheme(), cfg.mainID(), cfg.contentID()})
	return "consoleShell(" + string(b) + ")"
}
func appearanceBootstrapScript(cfg Config) string {
	b, _ := json.Marshal(shellOptions{Persist: cfg.Appearance.PersistPreferences, Theme: cfg.defaultTheme(), ColorScheme: cfg.initialColorScheme()})
	return `(function(o){document.documentElement.classList.add('js');var t=o.theme;var source="default";var d=o.colorScheme==='dark'||(o.colorScheme==='system'&&matchMedia('(prefers-color-scheme: dark)').matches);try{if(o.persist){var savedTheme=localStorage.getItem('goshtoso-theme');if(savedTheme){t=savedTheme;source="preference"}var s=localStorage.getItem('goshtoso-dark');if(s!==null)d=s==='true'}}catch(_){ }document.documentElement.dataset.theme=t;document.documentElement.dataset.themeSource=source;document.documentElement.classList.toggle('dark',d)})(` + string(b) + `);`
}
func currentPageTitle(cfg Config, page Page) string {
	if page.DocumentTitle != "" {
		return page.DocumentTitle
	}
	return page.Title + " · " + cfg.Brand.Name
}
func sidebarOOBAttributes(cfg Config, oob bool) templ.Attributes {
	if !oob || !cfg.Interactions.NavigationOOB {
		return nil
	}
	return templ.Attributes{"hx-swap-oob": "outerHTML:#consoleshell-sidebar-content"}
}
