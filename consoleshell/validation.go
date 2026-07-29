package consoleshell

import (
	"fmt"
	"github.com/araihu/goshtoso/components/sidebar"
	"strings"
)

func validate(cfg Config, page Page, fragment bool) error {
	if strings.TrimSpace(cfg.Brand.Name) == "" {
		return fmt.Errorf("console shell brand name is required")
	}
	if strings.TrimSpace(cfg.Brand.HomeURL) == "" {
		return fmt.Errorf("console shell brand home URL is required")
	}
	if strings.TrimSpace(page.Title) == "" {
		return fmt.Errorf("console shell page title is required")
	}
	if page.Content == nil {
		return fmt.Errorf("console shell page content is required")
	}
	if !strings.HasPrefix(cfg.assetPrefix(), "/") || !strings.HasSuffix(cfg.assetPrefix(), "/") {
		return fmt.Errorf("console shell asset prefix must start and end with /")
	}
	if strings.TrimSpace(cfg.mainID()) == "" || strings.TrimSpace(cfg.contentID()) == "" {
		return fmt.Errorf("console shell content IDs are required")
	}
	if fragment && !cfg.Interactions.EnableHTMX {
		return fmt.Errorf("console shell fragment rendering requires HTMX")
	}
	if len(cfg.Interactions.RuntimeScripts) > 0 && !cfg.Interactions.LocalRuntime {
		return fmt.Errorf("console shell runtime scripts require local runtime")
	}
	switch cfg.initialColorScheme() {
	case ColorSchemeSystem, ColorSchemeLight, ColorSchemeDark:
	default:
		return fmt.Errorf("console shell initial color scheme %q is invalid", cfg.initialColorScheme())
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
			return fmt.Errorf("console shell active navigation ID %q is not configured", page.Active)
		}
	}
	return nil
}
func validateItem(item sidebar.Item, ids map[string]struct{}) error {
	if strings.TrimSpace(item.ID) == "" {
		return fmt.Errorf("console shell navigation ID is required for %q", item.Label)
	}
	if _, ok := ids[item.ID]; ok {
		return fmt.Errorf("console shell duplicate navigation ID %q", item.ID)
	}
	ids[item.ID] = struct{}{}
	for _, child := range item.Items {
		if err := validateItem(child, ids); err != nil {
			return err
		}
	}
	return nil
}
