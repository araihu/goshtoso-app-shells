package consoleshell

import (
	"bytes"
	"context"
	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/sidebar"
	"strings"
	"testing"
)

func validConfig() Config {
	return Config{Brand: Brand{Name: "Console", HomeURL: "/"}, Navigation: Navigation{Items: []sidebar.Item{{ID: "home", Label: "Home", Href: "/"}, {ID: "jobs", Label: "Jobs", Href: "/jobs"}}}}
}
func validPage() Page {
	return Page{Title: "Jobs", Active: "jobs", Content: templ.Raw("<h1>Jobs</h1>")}
}
func render(t *testing.T, component templ.Component) string {
	t.Helper()
	var b bytes.Buffer
	if err := component.Render(context.Background(), &b); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

func TestLayoutRendersConsoleContract(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Interactions.EnableHTMX = true
	cfg.Interactions.NavigationOOB = true
	cfg.Brand.Logo = templ.Raw(`<svg id="brand-logo"></svg>`)
	cfg.Header = templ.Raw(`<div id="header-slot">Header</div>`)
	cfg.BodyEnd = templ.Raw(`<div id="body-end">Body</div>`)
	cfg.ModalSlot = templ.Raw(`<div id="modal-slot">Modal</div>`)
	body := render(t, Layout(cfg, validPage()))
	for _, want := range []string{"<!doctype html>", `id="main-content"`, `id="console-content"`, `id="consoleshell-sidebar"`, `id="brand-logo"`, `id="header-slot"`, `id="body-end"`, `id="modal-slot"`, `hx-get="/jobs"`, `hx-target="#main-content"`, `hx-swap="outerHTML"`, `hx-push-url="true"`, `data-consoleshell-nav-id="jobs"`, `/consoleshell/assets/shell.js`, `x-on:consoleshell:navigated.window="closeDrawer(false)"`, `href="/jobs"`, `classList.add('js')`} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q", want)
		}
	}
	if strings.Count(body, "<main ") != 1 {
		t.Fatalf("layout main landmarks = %d, want 1", strings.Count(body, "<main "))
	}
	if strings.Contains(body, "hx-swap-oob") {
		t.Fatal("layout must not emit hx-swap-oob")
	}
}
func TestFragmentUsesStableMainAndOptionalOOBNavigation(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Interactions.EnableHTMX = true
	cfg.Interactions.NavigationOOB = true
	body := render(t, Fragment(cfg, validPage()))
	for _, want := range []string{`<title>Jobs · Console</title>`, `id="main-content"`, `hx-swap-oob="outerHTML:#consoleshell-sidebar-content"`, `aria-current="page"`} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q", want)
		}
	}
	if strings.Contains(body, "<html") {
		t.Fatal("fragment rendered document")
	}
	if strings.Count(body, "<main ") != 1 {
		t.Fatalf("fragment main landmarks = %d, want 1", strings.Count(body, "<main "))
	}
	if !strings.Contains(body, `data-active-navigation="jobs"`) {
		t.Fatal("fragment missing active-navigation lifecycle state")
	}
}
func TestLayoutSupportsCustomContentContractAndLocalRuntime(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.MainID = "workspace"
	cfg.ContentID = "workspace-content"
	cfg.Interactions.EnableHTMX = true
	cfg.Interactions.FragmentTarget = "#workspace"
	cfg.Interactions.LocalRuntime = true
	cfg.Interactions.RuntimeScripts = []string{"/assets/extension.js"}
	body := render(t, Layout(cfg, validPage()))
	for _, want := range []string{`id="workspace"`, `id="workspace-content"`, `hx-target="#workspace"`, `src="/assets/extension.js"`, `/assets/js/runtime/htmx.org/`} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q", want)
		}
	}
}
func TestValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, want string
		edit       func(*Config, *Page)
		fragment   bool
	}{{"brand", "brand name is required", func(c *Config, _ *Page) { c.Brand.Name = "" }, false}, {"fragment", "fragment rendering requires HTMX", func(*Config, *Page) {}, true}, {"duplicate", "duplicate navigation ID", func(c *Config, _ *Page) {
		c.Navigation.Items = append(c.Navigation.Items, sidebar.Item{ID: "jobs", Label: "again"})
	}, false}, {"runtime", "runtime scripts require local runtime", func(c *Config, _ *Page) { c.Interactions.RuntimeScripts = []string{"/x.js"} }, false}}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			c, p := validConfig(), validPage()
			tt.edit(&c, &p)
			if err := validate(c, p, tt.fragment); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error=%v", err)
			}
		})
	}
}
