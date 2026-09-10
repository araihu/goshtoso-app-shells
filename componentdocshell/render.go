package componentdocshell

import (
	"context"
	"encoding/json"
	"io"
	"strconv"

	"github.com/a-h/templ"
	selectfield "github.com/araihu/goshtoso/components/select"
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
	searchSlot := cfg.Navigation.SearchSlot
	if searchSlot == nil && !cfg.Navigation.DisableSearch {
		searchSlot = defaultSidebarSearch(cfg.searchPlaceholder())
	}
	return sidebar.Config{
		Items:             cloneItems(cfg.Navigation.Items, active, cfg.Interactions.EnableHTMX),
		SectionsTitle:     cfg.Navigation.SectionsTitle,
		Sections:          cloneSections(cfg.Navigation.Sections, active, cfg.Interactions.EnableHTMX),
		ShowSearch:        false,
		SearchPlaceholder: cfg.searchPlaceholder(),
		SearchSlot:        searchSlot,
		RootClass:         "component-doc-shell__sidebar-nav",
		DisableSkipLink:   true,
	}
}

func brandInitial(name string) string {
	runes := []rune(name)
	if len(runes) == 0 {
		return ""
	}
	return string(runes[0])
}

func brandClass(brand Brand) string {
	classes := "component-doc-shell__brand"
	if brand.CompactLogo != nil && brand.ManagedLogo == nil && brand.Logo == nil {
		classes += " component-doc-shell__brand--compact-only"
	}
	return classes
}

func familyLinks(cfg Config, active string) []FamilyLink {
	result := make([]FamilyLink, len(cfg.Navigation.Families))
	for index, family := range cfg.Navigation.Families {
		result[index] = family
		if family.LinkAttrs != nil {
			result[index].LinkAttrs = make(templ.Attributes, len(family.LinkAttrs)+4)
			for key, value := range family.LinkAttrs {
				result[index].LinkAttrs[key] = value
			}
		}
		if family.ID == active {
			if result[index].LinkAttrs == nil {
				result[index].LinkAttrs = templ.Attributes{}
			}
			result[index].LinkAttrs["aria-current"] = "location"
		} else if result[index].LinkAttrs != nil {
			delete(result[index].LinkAttrs, "aria-current")
		}
		if cfg.Interactions.EnableHTMX {
			if result[index].LinkAttrs == nil {
				result[index].LinkAttrs = templ.Attributes{}
			}
			result[index].LinkAttrs["hx-get"] = family.Href
			result[index].LinkAttrs["hx-target"] = "#main-content"
			result[index].LinkAttrs["hx-push-url"] = "true"
		}
	}
	return result
}

func activeFamilyLabel(cfg Config, page Page) string {
	for _, family := range cfg.Navigation.Families {
		if family.ID == page.ActiveFamily {
			return family.Label
		}
	}
	return ""
}

func activeFamilyHref(cfg Config, page Page) string {
	for _, family := range cfg.Navigation.Families {
		if family.ID == page.ActiveFamily {
			return family.Href
		}
	}
	return ""
}

func familySelectOptions(cfg Config, page Page) []selectfield.Option {
	options := make([]selectfield.Option, len(cfg.Navigation.Families))
	for index, family := range cfg.Navigation.Families {
		options[index] = selectfield.Option{
			Value:    family.Href,
			Label:    family.Label,
			Selected: family.ID == page.ActiveFamily,
		}
	}
	return options
}

func familySelectData(cfg Config, page Page) string {
	href := activeFamilyHref(cfg, page)
	data, _ := json.Marshal(struct {
		FamilyHref        string `json:"familyHref"`
		InitialFamilyHref string `json:"initialFamilyHref"`
	}{
		FamilyHref:        href,
		InitialFamilyHref: href,
	})
	return string(data)
}

func scopeModuleLabel(scope *ScopeMetadata) string {
	if scope == nil {
		return ""
	}
	if scope.ModuleLabel != "" {
		return scope.ModuleLabel
	}
	return scope.ModulePath
}

func familyNavigationOOBAttributes(enabled bool) templ.Attributes {
	if !enabled {
		return nil
	}
	return templ.Attributes{"hx-swap-oob": "outerHTML:#componentdocshell-family-navigation"}
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
		Persist:      cfg.Appearance.PersistPreferences,
		PersistTheme: cfg.Appearance.PersistPreferences && !cfg.Appearance.DisableThemeSelector,
		Theme:        cfg.defaultTheme(),
		ColorScheme:  cfg.initialColorScheme(),
	})
	return "componentDocShell(" + string(options) + ")"
}

type shellOptions struct {
	Persist      bool        `json:"persist"`
	PersistTheme bool        `json:"persistTheme"`
	Theme        string      `json:"theme"`
	ColorScheme  ColorScheme `json:"colorScheme"`
}

func appearanceBootstrapScript(cfg Config) string {
	options, _ := json.Marshal(shellOptions{
		Persist:      cfg.Appearance.PersistPreferences,
		PersistTheme: cfg.Appearance.PersistPreferences && !cfg.Appearance.DisableThemeSelector,
		Theme:        cfg.defaultTheme(),
		ColorScheme:  cfg.initialColorScheme(),
	})
	return `(function(o){var theme=o.theme;var source="default";var dark=o.colorScheme==="dark"||(o.colorScheme==="system"&&window.matchMedia("(prefers-color-scheme: dark)").matches);try{if(o.persistTheme){var savedTheme=localStorage.getItem("theme");if(savedTheme){theme=savedTheme;source="preference"}}if(o.persist){var saved=localStorage.getItem("darkMode");if(saved!==null)dark=saved==="true"}}catch(_){}document.documentElement.setAttribute("data-theme",theme);document.documentElement.dataset.themeSource=source;document.documentElement.classList.toggle("dark",dark);})(` + string(options) + `);`
}

func currentPageTitle(cfg Config, page Page) string {
	if page.DocumentTitle != "" {
		return page.DocumentTitle
	}
	return page.Title + " · " + cfg.Brand.Name
}

func socialSiteName(cfg Config, page Page) string {
	if page.SiteName != "" {
		return page.SiteName
	}
	return cfg.Brand.Name
}

func socialDimension(value int) string {
	return strconv.Itoa(value)
}

func sidebarOOBAttributes(enabled bool) templ.Attributes {
	if !enabled {
		return nil
	}
	// Keep the scroll container mounted so navigation cannot flash at scrollTop 0
	// before the after-swap handler restores its position.
	return templ.Attributes{"hx-swap-oob": "outerMorph:#componentdocshell-sidebar-content"}
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
