package componentdocshell

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"

	"github.com/araihu/goshtoso/components/sidebar"
)

func validate(cfg Config, page Page, fragment bool) error {
	if len(cfg.Interactions.RuntimeScripts) > 0 && !cfg.Interactions.LocalRuntime {
		return fmt.Errorf("component docs shell runtime scripts require local runtime")
	}
	if strings.TrimSpace(cfg.Brand.Name) == "" {
		return fmt.Errorf("component docs shell brand name is required")
	}
	if strings.TrimSpace(cfg.Brand.HomeURL) == "" {
		return fmt.Errorf("component docs shell brand home URL is required")
	}
	if err := validatePresentationChannel(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(page.Title) == "" {
		return fmt.Errorf("component docs shell page title is required")
	}
	if page.Content == nil {
		return fmt.Errorf("component docs shell page content is required")
	}
	if prefix := cfg.assetPrefix(); !strings.HasPrefix(prefix, "/") || !strings.HasSuffix(prefix, "/") {
		return fmt.Errorf("component docs shell asset prefix must start and end with /")
	}
	if fragment && !cfg.Interactions.EnableHTMX {
		return fmt.Errorf("component docs shell fragment rendering requires HTMX")
	}
	switch cfg.initialColorScheme() {
	case ColorSchemeSystem, ColorSchemeLight, ColorSchemeDark:
	default:
		return fmt.Errorf("component docs shell initial color scheme %q is invalid", cfg.initialColorScheme())
	}
	if !cfg.Appearance.DisableThemeSelector {
		found := false
		for _, theme := range cfg.themes() {
			if theme.Value == cfg.defaultTheme() {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("component docs shell default theme %q is not available", cfg.defaultTheme())
		}
	}

	ids := map[string]struct{}{}
	for _, item := range cfg.Navigation.Items {
		if err := validateItem(item, ids); err != nil {
			return err
		}
	}
	for _, section := range cfg.Navigation.Sections {
		for _, item := range section.Items {
			if err := validateItem(item, ids); err != nil {
				return err
			}
		}
	}
	if page.Active != "" {
		if _, ok := ids[page.Active]; !ok {
			return fmt.Errorf("component docs shell active navigation ID %q is not configured", page.Active)
		}
	}
	return nil
}

type presentationURL struct {
	absolute bool
	origin   string
}

func validatePresentationChannel(cfg Config) error {
	channel := cfg.Interactions.PresentationChannel
	if channel == nil {
		return nil
	}
	if cfg.Brand.Logo != nil && cfg.Brand.ManagedLogo != nil {
		return fmt.Errorf("component docs shell managed logo conflicts with brand logo")
	}
	if cfg.Brand.ManagedLogo == nil && !cfg.Brand.ManageFavicon {
		return fmt.Errorf("component docs shell presentation channel requires a managed logo or managed favicon")
	}
	if cfg.Brand.ManagedLogo != nil {
		if cfg.Brand.ManagedLogo.Width == 0 || cfg.Brand.ManagedLogo.Height == 0 {
			return fmt.Errorf("component docs shell managed logo dimensions must be positive")
		}
		if _, err := validatePresentationURL("managed logo URL", cfg.Brand.ManagedLogo.URL); err != nil {
			return err
		}
		if err := validatePresentationText("managed logo alt text", cfg.Brand.ManagedLogo.Alt, false); err != nil {
			return err
		}
	}
	runtime, err := validatePresentationURL("presentation runtime URL", channel.RuntimeURL)
	if err != nil {
		return err
	}
	manifest, err := validatePresentationURL("presentation channel URL", channel.ChannelURL)
	if err != nil {
		return err
	}
	if runtime.absolute != manifest.absolute || (runtime.absolute && runtime.origin != manifest.origin) {
		return fmt.Errorf("component docs shell presentation runtime and channel URLs must use the same origin")
	}
	if !strings.HasPrefix(channel.Integrity, "sha384-") || strings.TrimSpace(strings.TrimPrefix(channel.Integrity, "sha384-")) == "" {
		return fmt.Errorf("component docs shell presentation channel integrity must use sha384-")
	}
	if err := validatePresentationText("presentation channel integrity", channel.Integrity, true); err != nil {
		return err
	}
	if err := validatePresentationText("presentation campaign label", channel.UseCampaignLabel, true); err != nil {
		return err
	}
	return validatePresentationText("presentation baseline label", channel.UseBaselineLabel, true)
}

func validatePresentationURL(field, value string) (presentationURL, error) {
	if strings.TrimSpace(value) == "" {
		return presentationURL{}, fmt.Errorf("component docs shell %s is required", field)
	}
	if hasControlCharacter(value) {
		return presentationURL{}, fmt.Errorf("component docs shell %s contains a control character", field)
	}
	if strings.Contains(value, "\\") {
		return presentationURL{}, fmt.Errorf("component docs shell %s must not include a backslash", field)
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return presentationURL{}, fmt.Errorf("component docs shell %s is invalid: %w", field, err)
	}
	if parsed.Fragment != "" {
		return presentationURL{}, fmt.Errorf("component docs shell %s must not include a fragment", field)
	}
	if parsed.User != nil {
		return presentationURL{}, fmt.Errorf("component docs shell %s must not include credentials", field)
	}
	if strings.HasPrefix(value, "/") && parsed.Scheme == "" && parsed.Host == "" {
		return presentationURL{}, nil
	}
	if parsed.Scheme != "https" || parsed.Host == "" || parsed.Hostname() == "" {
		return presentationURL{}, fmt.Errorf("component docs shell %s must be root-relative or an absolute HTTPS URL", field)
	}
	origin := strings.ToLower(parsed.Hostname())
	if port := parsed.Port(); port != "" && port != "443" {
		origin += ":" + port
	}
	return presentationURL{absolute: true, origin: origin}, nil
}

func validatePresentationText(field, value string, required bool) error {
	if required && strings.TrimSpace(value) == "" {
		return fmt.Errorf("component docs shell %s is required", field)
	}
	if hasControlCharacter(value) {
		return fmt.Errorf("component docs shell %s contains a control character", field)
	}
	return nil
}

func hasControlCharacter(value string) bool {
	return strings.IndexFunc(value, unicode.IsControl) >= 0
}

func validateItem(item sidebar.Item, ids map[string]struct{}) error {
	id := strings.TrimSpace(item.ID)
	if id == "" {
		return fmt.Errorf("component docs shell navigation ID is required for %q", item.Label)
	}
	if _, exists := ids[id]; exists {
		return fmt.Errorf("component docs shell duplicate navigation ID %q", id)
	}
	ids[id] = struct{}{}
	for _, child := range item.Items {
		if err := validateItem(child, ids); err != nil {
			return err
		}
	}
	return nil
}
