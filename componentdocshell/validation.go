package componentdocshell

import (
	"fmt"
	"strings"

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
