package landingshell

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/a-h/templ"
)

// Layout renders a complete server-side public landing document.
func Layout(cfg Config, page Page) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, writer io.Writer) error {
		if err := validate(cfg, page); err != nil {
			return err
		}
		return layoutTemplate(cfg, page).Render(ctx, writer)
	})
}

// Head renders shell runtime and stylesheet dependencies for custom layouts.
func Head(cfg Config) templ.Component { return headTemplate(cfg) }

// MobileNavigation renders responsive navigation chrome around consumer-owned
// navigation content. The slot is rendered once for the enhanced Drawer and
// once for the native no-runtime fallback.
func MobileNavigation(cfg MobileNavigationConfig, content templ.Component) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, writer io.Writer) error {
		if err := validateMobileNavigation(cfg); err != nil {
			return err
		}
		if content == nil {
			return fmt.Errorf("landing shell mobile navigation content is required")
		}
		return mobileNavigationTemplate(cfg, content).Render(ctx, writer)
	})
}

type shellOptions struct {
	Persist     bool        `json:"persist"`
	Theme       string      `json:"theme"`
	ColorScheme ColorScheme `json:"colorScheme"`
}

func shellData(cfg Config) string {
	options, _ := json.Marshal(shellOptions{
		Persist:     cfg.Appearance.PersistPreferences,
		Theme:       cfg.defaultTheme(),
		ColorScheme: cfg.initialColorScheme(),
	})
	return "landingShell(" + string(options) + ")"
}

func appearanceBootstrapScript(cfg Config) string {
	options, _ := json.Marshal(shellOptions{
		Persist:     cfg.Appearance.PersistPreferences,
		Theme:       cfg.defaultTheme(),
		ColorScheme: cfg.initialColorScheme(),
	})
	return `(function(o){document.documentElement.classList.add('js');document.documentElement.dataset.theme=o.theme;var d=o.colorScheme==='dark'||(o.colorScheme==='system'&&matchMedia('(prefers-color-scheme: dark)').matches);try{if(o.persist){var s=localStorage.getItem('darkMode');if(s!==null)d=s==='true'}}catch(_){ }document.documentElement.classList.toggle('dark',d)})(` + string(options) + `);`
}

func currentPageTitle(cfg Config, page Page) string {
	if page.DocumentTitle != "" {
		return page.DocumentTitle
	}
	return page.Title + " · " + cfg.Brand.Name
}

func navLinkClass(item Link) string {
	classes := []string{"landing-shell__nav-link", "is-secondary"}
	if item.Primary {
		classes[1] = "is-primary"
	}
	return strings.Join(classes, " ")
}

func landingHeaderClass(cfg Config) string {
	if cfg.MobileNavigation != nil {
		return "landing-shell__header landing-shell__header--mobile-navigation"
	}
	return "landing-shell__header"
}
