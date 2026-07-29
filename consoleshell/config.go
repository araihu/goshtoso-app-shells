// Package consoleshell renders reusable server-side console and application
// frames. It deliberately has no documentation or table-of-contents semantics.
package consoleshell

import (
	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/sidebar"
)

const defaultAssetPrefix = "/consoleshell/assets/"

type ColorScheme string

const (
	ColorSchemeSystem ColorScheme = "system"
	ColorSchemeLight  ColorScheme = "light"
	ColorSchemeDark   ColorScheme = "dark"
)

type Brand struct {
	Name       string
	HomeURL    string
	Logo       templ.Component
	FaviconURL string
}

// Navigation is server-rendered. Links always retain href fallback; HTMX
// attributes are added only when Interactions.EnableHTMX is enabled.
type Navigation struct {
	Items             []sidebar.Item
	SectionsTitle     string
	Sections          []sidebar.Section
	DisableSearch     bool
	SearchPlaceholder string
	SearchSlot        templ.Component
}

type AppearanceConfig struct {
	DefaultTheme       string
	InitialColorScheme ColorScheme
	PersistPreferences bool
	ThemeStylesheets   []string
}

// InteractionConfig selects progressive enhancement and the stable fragment
// contract. LocalRuntime makes Goshtoso dependencies fully local.
type InteractionConfig struct {
	EnableHTMX     bool
	LocalRuntime   bool
	RuntimeScripts []string
	FragmentTarget string
	NavigationOOB  bool
}

type Config struct {
	Brand         Brand
	Navigation    Navigation
	Appearance    AppearanceConfig
	Interactions  InteractionConfig
	Header        templ.Component
	HeaderActions templ.Component
	SidebarHeader templ.Component
	Footer        templ.Component
	BodyEnd       templ.Component
	ModalSlot     templ.Component
	AssetPrefix   string
	MainID        string
	ContentID     string
}

type Page struct {
	Title         string
	DocumentTitle string
	Description   string
	CanonicalURL  string
	Active        string
	Content       templ.Component
	Head          templ.Component
}

func (cfg Config) assetPrefix() string {
	if cfg.AssetPrefix == "" {
		return defaultAssetPrefix
	}
	return cfg.AssetPrefix
}
func (cfg Config) mainID() string {
	if cfg.MainID == "" {
		return "main-content"
	}
	return cfg.MainID
}
func (cfg Config) contentID() string {
	if cfg.ContentID == "" {
		return "console-content"
	}
	return cfg.ContentID
}
func (cfg Config) fragmentTarget() string {
	if cfg.Interactions.FragmentTarget == "" {
		return "#" + cfg.mainID()
	}
	return cfg.Interactions.FragmentTarget
}
func (cfg Config) defaultTheme() string {
	if cfg.Appearance.DefaultTheme == "" {
		return "goshtoso"
	}
	return cfg.Appearance.DefaultTheme
}
func (cfg Config) initialColorScheme() ColorScheme {
	if cfg.Appearance.InitialColorScheme == "" {
		return ColorSchemeSystem
	}
	return cfg.Appearance.InitialColorScheme
}
func (cfg Config) searchPlaceholder() string {
	if cfg.Navigation.SearchPlaceholder == "" {
		return "Search"
	}
	return cfg.Navigation.SearchPlaceholder
}
