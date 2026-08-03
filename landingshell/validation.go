package landingshell

import (
	"fmt"
	"strings"
)

func validate(cfg Config, page Page) error {
	if strings.TrimSpace(cfg.Brand.Name) == "" {
		return fmt.Errorf("landing shell brand name is required")
	}
	if strings.TrimSpace(cfg.Brand.HomeURL) == "" {
		return fmt.Errorf("landing shell brand home URL is required")
	}
	if strings.TrimSpace(page.Title) == "" {
		return fmt.Errorf("landing shell page title is required")
	}
	if page.Hero == nil {
		return fmt.Errorf("landing shell hero is required")
	}
	if page.Content == nil {
		return fmt.Errorf("landing shell content is required")
	}
	if prefix := cfg.assetPrefix(); !strings.HasPrefix(prefix, "/") || !strings.HasSuffix(prefix, "/") {
		return fmt.Errorf("landing shell asset prefix must start and end with /")
	}
	switch cfg.initialColorScheme() {
	case ColorSchemeSystem, ColorSchemeLight, ColorSchemeDark:
	default:
		return fmt.Errorf("landing shell initial color scheme %q is invalid", cfg.initialColorScheme())
	}
	for _, item := range cfg.Navigation {
		if err := validateLink("navigation", item); err != nil {
			return err
		}
	}
	if cfg.MobileNavigation != nil {
		if err := validateMobileNavigation(*cfg.MobileNavigation); err != nil {
			return err
		}
	}
	for _, item := range cfg.Footer.Links {
		if err := validateLink("footer", item); err != nil {
			return err
		}
	}
	if organization := cfg.Footer.Organization; organization != nil {
		if strings.TrimSpace(organization.Name) == "" || strings.TrimSpace(organization.URL) == "" {
			return fmt.Errorf("landing shell footer organization name and URL are required together")
		}
	}
	return nil
}

func validateMobileNavigation(cfg MobileNavigationConfig) error {
	switch cfg.position() {
	case FloatingBottomLeft, FloatingBottomRight:
	default:
		return fmt.Errorf("landing shell mobile navigation position %q is invalid", cfg.Position)
	}
	if !validMobileNavigationID(cfg.id()) {
		return fmt.Errorf("landing shell mobile navigation ID %q is invalid", cfg.id())
	}
	return nil
}

func validMobileNavigationID(value string) bool {
	for index, char := range value {
		valid := char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || char == '-' || char == '_'
		if !valid || index == 0 && char >= '0' && char <= '9' {
			return false
		}
	}
	return value != ""
}

func validateLink(scope string, item Link) error {
	if strings.TrimSpace(item.Label) == "" {
		return fmt.Errorf("landing shell %s link label is required", scope)
	}
	if strings.TrimSpace(item.Href) == "" {
		return fmt.Errorf("landing shell %s link href is required for %q", scope, item.Label)
	}
	return nil
}
