package landingshell

import "strings"

func (cfg MobileNavigationConfig) id() string {
	if cfg.ID == "" {
		return "landing-navigation"
	}
	return cfg.ID
}

func (cfg MobileNavigationConfig) title() string {
	if cfg.Title == "" {
		return "Navigation"
	}
	return cfg.Title
}

func (cfg MobileNavigationConfig) triggerLabel() string {
	if cfg.TriggerLabel == "" {
		return "Menu"
	}
	return cfg.TriggerLabel
}

func (cfg MobileNavigationConfig) navigationLabel() string {
	if cfg.NavigationLabel == "" {
		return "Primary navigation"
	}
	return cfg.NavigationLabel
}

func (cfg MobileNavigationConfig) position() FloatingPosition {
	if cfg.Position == "" {
		return FloatingBottomLeft
	}
	return cfg.Position
}

func (cfg MobileNavigationConfig) rootClass() string {
	return joinClasses("landing-shell__mobile-navigation", cfg.RootClass)
}

func (cfg MobileNavigationConfig) triggerClass() string {
	return joinClasses("landing-shell__mobile-trigger", "is-"+string(cfg.position()), cfg.TriggerClass)
}

func (cfg MobileNavigationConfig) fallbackTriggerClass() string {
	return joinClasses("landing-shell__mobile-fallback-trigger", "is-"+string(cfg.position()), cfg.TriggerClass)
}

func joinClasses(values ...string) string {
	var classes []string
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			classes = append(classes, value)
		}
	}
	return strings.Join(classes, " ")
}

func (cfg MobileNavigationConfig) openExpression() string {
	return "$dispatch('drawer:open', { id: '" + cfg.id() + "' })"
}

func (cfg MobileNavigationConfig) openStateExpression() string {
	return "if ($event.detail && $event.detail.id === '" + cfg.id() + "') navigationOpen = true"
}

func (cfg MobileNavigationConfig) closeStateExpression() string {
	return "if ($event.detail && $event.detail.id === '" + cfg.id() + "') navigationOpen = false"
}

func (cfg MobileNavigationConfig) closeRequestStateExpression() string {
	return "if ($event.detail && $event.detail.id === '" + cfg.id() + "') $nextTick(() => { if (!$event.defaultPrevented) navigationOpen = false })"
}

func (cfg MobileNavigationConfig) closeExpression() string {
	return "if ($event.target.closest('a')) $dispatch('drawer:close', { id: '" + cfg.id() + "' })"
}
