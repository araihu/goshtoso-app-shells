// Package catalogshell renders reusable server-side catalog and documentation
// frames built from Goshtoso primitives.
package catalogshell

import (
	"github.com/a-h/templ"
	selectfield "github.com/araihu/goshtoso/components/select"
	"github.com/araihu/goshtoso/components/sidebar"
)

const defaultAssetPrefix = "/catalogshell/assets/"

// Brand describes the application identity shown in the shell header.
type Brand struct {
	Name    string
	HomeURL string
	Logo    templ.Component
}

// Navigation describes top-level and grouped sidebar entries.
type Navigation struct {
	Items             []sidebar.Item
	SectionsTitle     string
	Sections          []sidebar.Section
	SearchPlaceholder string
}

// Config describes presentation shared by every page in one catalog site.
type Config struct {
	Brand              Brand
	Navigation         Navigation
	Themes             []selectfield.Option
	HeaderActions      templ.Component
	Footer             templ.Component
	RepositoryURL      string
	AssetPrefix        string
	EnableHTMX         bool
	PersistPreferences bool
}

// Page describes one server-rendered catalog response.
type Page struct {
	Title        string
	Description  string
	CanonicalURL string
	Active       string
	Content      templ.Component
	Head         templ.Component
	EnableTOC    bool
}

func (cfg Config) assetPrefix() string {
	if cfg.AssetPrefix == "" {
		return defaultAssetPrefix
	}
	return cfg.AssetPrefix
}

func (cfg Config) themes() []selectfield.Option {
	if len(cfg.Themes) > 0 {
		return cfg.Themes
	}
	return []selectfield.Option{
		{Value: "goshtoso", Label: "Goshtoso", Selected: true},
		{Value: "minimal", Label: "Minimal"},
	}
}

func (cfg Config) searchPlaceholder() string {
	if cfg.Navigation.SearchPlaceholder != "" {
		return cfg.Navigation.SearchPlaceholder
	}
	return "Search catalog"
}
