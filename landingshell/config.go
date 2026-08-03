// Package landingshell renders reusable server-side public landing-page frames
// built from Goshtoso assets and semantic theme tokens.
package landingshell

import "github.com/a-h/templ"

const defaultAssetPrefix = "/landingshell/assets/"

// ColorScheme selects the initial light, dark, or system appearance.
type ColorScheme string

const (
	ColorSchemeSystem ColorScheme = "system"
	ColorSchemeLight  ColorScheme = "light"
	ColorSchemeDark   ColorScheme = "dark"
)

// Brand configures the identity displayed in the shell header.
type Brand struct {
	Name       string
	HomeURL    string
	Logo       templ.Component
	Tagline    string
	FaviconURL string
	Badge      *BrandBadge
}

// BrandBadge adds compact version or release metadata beside the brand.
type BrandBadge struct {
	Label     string
	AriaLabel string
	Href      string
}

// Link configures one header or footer destination.
type Link struct {
	Label    string
	Href     string
	External bool
	Primary  bool
}

// AppearanceConfig configures the fixed product theme and color mode.
type AppearanceConfig struct {
	DefaultTheme          string
	InitialColorScheme    ColorScheme
	PersistPreferences    bool
	DisableDarkModeToggle bool
	ThemeStylesheets      []string
}

// InteractionConfig controls the Goshtoso dependency source.
type InteractionConfig struct {
	LocalRuntime bool
}

// Organization identifies the organization linked in the default footer copy.
type Organization struct {
	Name string
	URL  string
}

// Footer configures the structured Goshtoso-style landing footer.
type Footer struct {
	Logo         templ.Component
	Name         string
	Meta         []string
	Organization *Organization
	Links        []Link
}

// Config defines shell-wide identity, navigation, appearance, and slots.
type Config struct {
	Brand         Brand
	Navigation    []Link
	Appearance    AppearanceConfig
	Interactions  InteractionConfig
	Footer        Footer
	RepositoryURL string
	HeaderActions templ.Component
	BodyEnd       templ.Component
	AssetPrefix   string
}

// Page defines one complete public landing document.
type Page struct {
	Title          string
	DocumentTitle  string
	Description    string
	CanonicalURL   string
	SocialImageURL string
	Head           templ.Component
	Hero           templ.Component
	Content        templ.Component
}

func (cfg Config) assetPrefix() string {
	if cfg.AssetPrefix == "" {
		return defaultAssetPrefix
	}
	return cfg.AssetPrefix
}

func (cfg Config) defaultTheme() string {
	if cfg.Appearance.DefaultTheme == "" {
		return "araihu"
	}
	return cfg.Appearance.DefaultTheme
}

func (cfg Config) initialColorScheme() ColorScheme {
	if cfg.Appearance.InitialColorScheme == "" {
		return ColorSchemeSystem
	}
	return cfg.Appearance.InitialColorScheme
}

func (cfg Config) footerName() string {
	if cfg.Footer.Name != "" {
		return cfg.Footer.Name
	}
	return cfg.Brand.Name
}
