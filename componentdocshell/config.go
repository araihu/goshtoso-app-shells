// Package componentdocshell renders reusable server-side component documentation
// frames built from Goshtoso primitives.
package componentdocshell

import (
	"github.com/a-h/templ"
	selectfield "github.com/araihu/goshtoso/components/select"
	"github.com/araihu/goshtoso/components/sidebar"
)

const defaultAssetPrefix = "/componentdocshell/assets/"

// ColorScheme controls the initial light/dark state before an allowed stored
// preference is applied.
type ColorScheme string

const (
	ColorSchemeSystem ColorScheme = "system"
	ColorSchemeLight  ColorScheme = "light"
	ColorSchemeDark   ColorScheme = "dark"
)

// Brand describes the application identity shown in the shell header.
type Brand struct {
	Name    string
	HomeURL string
	Logo    templ.Component
	// CompactLogo is an optional purpose-built mark for the small responsive
	// header. When omitted, the shell falls back to the first rune of Name.
	CompactLogo   templ.Component
	HideName      bool
	ManagedLogo   *ManagedBrandAsset
	ManageFavicon bool
	FaviconURL    string
	Badge         *BrandBadge
}

// ManagedBrandAsset describes a brand image whose lifecycle is owned by the
// presentation channel rather than an application-provided templ component.
type ManagedBrandAsset struct {
	URL    string
	Alt    string
	Width  uint
	Height uint
}

// BrandBadge adds compact build or release metadata beside the brand.
// Href is optional; an empty value renders non-interactive text.
type BrandBadge struct {
	Label     string
	AriaLabel string
	Href      string
}

// DarkModeBinding lets an application preserve an established Alpine dark-mode
// store and DOM identifier while the shell keeps its built-in control markup.
// Empty fields retain the shell defaults.
type DarkModeBinding struct {
	ButtonID         string
	StateExpression  string
	ToggleExpression string
}

// FamilyLink describes one destination in the shell's global family
// navigation. When Navigation.Families is non-empty, each link needs a unique
// non-blank ID without leading or trailing whitespace, a non-blank label, and
// a root-relative or absolute HTTPS Href;
// the page's ActiveFamily must match one of those IDs. The shell renders each
// link on desktop and mobile surfaces, so LinkAttrs must not contain an id key
// (case-insensitively). The shell copies LinkAttrs without mutating the
// caller's map, retains unrelated attributes, and always owns aria-current on
// both responsive surfaces. When HTMX enhancement is enabled, it also owns
// hx-get, hx-target, and hx-push-url. The ordinary Href anchor remains in every
// mode, including no-JavaScript or HTMX-disabled full-page navigation.
type FamilyLink struct {
	ID        string
	Label     string
	Href      string
	LinkAttrs templ.Attributes
}

// ScopeMetadata is optional module and release context shown in the scoped
// sidebar, not in global family navigation. Consumers provide route-specific
// local navigation and scope metadata; Page.ActiveFamily only selects the
// active family state and does not map or select Items, Sections, SearchSlot,
// or Scope. ModulePath, Version, and VersionURL are independently optional.
// ModuleLabel and ModuleURL require ModulePath; ModuleLabel defaults to
// ModulePath. VersionURL requires Version. Both URLs must be root-relative or
// absolute HTTPS URLs.
type ScopeMetadata struct {
	ModulePath  string
	ModuleLabel string
	ModuleURL   string
	Version     string
	VersionURL  string
}

// Navigation describes the shell's global family navigation plus the active
// family's local navigation and grouped sidebar entries. Families is optional:
// an empty slice preserves legacy rendering and does not require
// Page.ActiveFamily. With families configured, validation requires unique
// non-blank IDs without leading or trailing whitespace, non-blank labels,
// root-relative or absolute HTTPS overview URLs, and a matching
// Page.ActiveFamily. Items, Sections, and SearchSlot stay consumer-owned and
// should be supplied for each route's active family.
type Navigation struct {
	Families          []FamilyLink
	Scope             *ScopeMetadata
	Items             []sidebar.Item
	SectionsTitle     string
	Sections          []sidebar.Section
	SearchPlaceholder string
	DisableSearch     bool
	SearchSlot        templ.Component
}

// AppearanceConfig controls theming and the related header controls.
type AppearanceConfig struct {
	Themes                        []selectfield.Option
	DefaultTheme                  string
	ThemeSelectorID               string
	InitialColorScheme            ColorScheme
	DisableThemeSelector          bool
	DisableDarkModeToggle         bool
	PersistPreferences            bool
	DisableDefaultThemeStylesheet bool
	ThemeStylesheets              []string
	DarkModeBinding               *DarkModeBinding
}

// InteractionConfig controls optional progressive enhancement.
type InteractionConfig struct {
	EnableHTMX          bool
	LocalRuntime        bool
	RuntimeScripts      []string
	PresentationChannel *PresentationChannelConfig
}

// PresentationChannelConfig identifies an integrity-pinned campaign runtime
// and channel manifest. A nil value keeps presentation channels disabled.
type PresentationChannelConfig struct {
	RuntimeURL       string
	ChannelURL       string
	Integrity        string
	UseCampaignLabel string
	UseBaselineLabel string
}

// TOCConfig preserves application-owned identifiers while the shell uses
// semantic data hooks for behavior. Empty fields retain shell defaults.
type TOCConfig struct {
	RailID string
	ListID string
}

// Config describes presentation shared by every page in one component documentation site.
type Config struct {
	Brand        Brand
	Navigation   Navigation
	Appearance   AppearanceConfig
	Interactions InteractionConfig
	TOC          TOCConfig
	// HeaderActions is one consumer-owned component rendered once in the
	// header. The shell does not clone or move it for small layouts because it
	// may contain IDs or state; consumers own responsive reachability.
	HeaderActions templ.Component
	Footer        templ.Component
	BodyEnd       templ.Component
	RepositoryURL string
	AssetPrefix   string
}

// Page describes one server-rendered component documentation response.
// Active identifies the active local navigation page and renders
// aria-current="page". ActiveFamily identifies the active global family and
// renders aria-current="location" when Navigation.Families is configured;
// it must match a configured family ID. The shell uses ActiveFamily only for
// family state and overview navigation: selecting or switching a family opens
// that link's configured overview Href and never maps to an analogous current
// page. Consumers still provide the route-specific local Items, Sections,
// SearchSlot, and optional Scope. When Navigation.Families is empty,
// ActiveFamily may remain blank for legacy consumers.
type Page struct {
	Title string
	// DocumentTitle overrides the default "Page · Brand" browser title while
	// Title remains the shell's required human-readable page name.
	DocumentTitle string
	Description   string
	CanonicalURL  string
	// SiteName and Locale provide Open Graph identity for share previews.
	// SiteName falls back to Brand.Name when omitted.
	SiteName     string
	Locale       string
	SocialImage  SocialImage
	ActiveFamily string
	Active       string
	Content      templ.Component
	Head         templ.Component
	EnableTOC    bool
}

// SocialImage describes the absolute public image used by Open Graph and X.
// MIMEType, dimensions, and Alt are required whenever URL is configured.
type SocialImage struct {
	URL      string
	MIMEType string
	Width    int
	Height   int
	Alt      string
}

func (cfg Config) assetPrefix() string {
	if cfg.AssetPrefix == "" {
		return defaultAssetPrefix
	}
	return cfg.AssetPrefix
}

func (cfg Config) themes() []selectfield.Option {
	themes := cfg.Appearance.Themes
	if len(themes) == 0 {
		themes = DefaultThemes()
	}
	result := make([]selectfield.Option, len(themes))
	copy(result, themes)
	for index := range result {
		result[index].Selected = result[index].Value == cfg.defaultTheme()
	}
	return result
}

// DefaultThemes returns Arai Hû followed by every theme compiled into
// Goshtoso v0.1.6. Consumers may reorder or replace this list.
func DefaultThemes() []selectfield.Option {
	return []selectfield.Option{
		{Value: "araihu", Label: "Arai Hû"},
		{Value: "goshtoso", Label: "Goshtoso"},
		{Value: "arctic", Label: "Arctic"},
		{Value: "high-contrast", Label: "High Contrast"},
		{Value: "minimal", Label: "Minimal"},
		{Value: "modern", Label: "Modern"},
		{Value: "neo-brutalism", Label: "Neo Brutalism"},
		{Value: "halloween", Label: "Halloween"},
		{Value: "zombie", Label: "Zombie"},
		{Value: "pastel", Label: "Pastel"},
		{Value: "90s", Label: "90s"},
		{Value: "christmas", Label: "Christmas"},
		{Value: "prototype", Label: "Prototype"},
		{Value: "news", Label: "News"},
		{Value: "industrial", Label: "Industrial"},
		{Value: "dracula", Label: "Dracula"},
	}
}

func (cfg Config) defaultTheme() string {
	if cfg.Appearance.DefaultTheme != "" {
		return cfg.Appearance.DefaultTheme
	}
	return "araihu"
}

func (cfg Config) initialColorScheme() ColorScheme {
	if cfg.Appearance.InitialColorScheme != "" {
		return cfg.Appearance.InitialColorScheme
	}
	return ColorSchemeSystem
}

func (cfg Config) themeSelectorID() string {
	if cfg.Appearance.ThemeSelectorID != "" {
		return cfg.Appearance.ThemeSelectorID
	}
	return "componentdocshell-theme"
}

func (cfg Config) desktopThemeSelectorID() string {
	return cfg.themeSelectorID()
}

func (cfg Config) mobileThemeSelectorID() string {
	return cfg.themeSelectorID() + "-mobile"
}

func (cfg Config) darkModeBinding() DarkModeBinding {
	binding := DarkModeBinding{
		ButtonID:         "componentdocshell-dark-mode",
		StateExpression:  "dark",
		ToggleExpression: "toggleDark()",
	}
	if cfg.Appearance.DarkModeBinding == nil {
		return binding
	}
	if cfg.Appearance.DarkModeBinding.ButtonID != "" {
		binding.ButtonID = cfg.Appearance.DarkModeBinding.ButtonID
	}
	if cfg.Appearance.DarkModeBinding.StateExpression != "" {
		binding.StateExpression = cfg.Appearance.DarkModeBinding.StateExpression
	}
	if cfg.Appearance.DarkModeBinding.ToggleExpression != "" {
		binding.ToggleExpression = cfg.Appearance.DarkModeBinding.ToggleExpression
	}
	return binding
}

func (cfg Config) darkModeAriaExpression() string {
	return cfg.darkModeBinding().StateExpression + " ? 'Switch to light mode' : 'Switch to dark mode'"
}

func (cfg Config) darkModeOffExpression() string {
	return "!(" + cfg.darkModeBinding().StateExpression + ")"
}

func (cfg Config) tocRailID() string {
	if cfg.TOC.RailID != "" {
		return cfg.TOC.RailID
	}
	return "componentdocshell-toc"
}

func (cfg Config) tocListID() string {
	if cfg.TOC.ListID != "" {
		return cfg.TOC.ListID
	}
	return "componentdocshell-toc-list"
}

func (cfg Config) searchPlaceholder() string {
	if cfg.Navigation.SearchPlaceholder != "" {
		return cfg.Navigation.SearchPlaceholder
	}
	return "Search docs..."
}
