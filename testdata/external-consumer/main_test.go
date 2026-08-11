package consumer_test

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-app-shells/componentdocshell"
	shellassets "github.com/araihu/goshtoso-app-shells/componentdocshell/assets"
	"github.com/araihu/goshtoso/components/sidebar"
)

var docsConfig = componentdocshell.Config{
	Brand: componentdocshell.Brand{Name: "Example docs", HomeURL: "/"},
	Navigation: componentdocshell.Navigation{
		Families: []componentdocshell.FamilyLink{
			{ID: "components", Label: "Components", Href: "/components"},
			{ID: "charts", Label: "Charts", Href: "/charts"},
			{ID: "app-shells", Label: "App Shells", Href: "/app-shells"},
			{ID: "icons", Label: "Icons", Href: "/icons"},
			{ID: "llms", Label: "LLMs", Href: "/llms"},
			{ID: "examples", Label: "Examples", Href: "/examples"},
		},
		Scope: &componentdocshell.ScopeMetadata{
			ModulePath: "example.com/componentdocshell",
			Version:    "v0.0.0-example",
			VersionURL: "https://example.com/componentdocshell/releases/v0.0.0-example",
		},
		Items: []sidebar.Item{{ID: "overview", Label: "Overview", Href: "/components"}},
	},
	Appearance: componentdocshell.AppearanceConfig{
		DefaultTheme:       "araihu",
		InitialColorScheme: componentdocshell.ColorSchemeSystem,
		PersistPreferences: true,
	},
	Interactions: componentdocshell.InteractionConfig{EnableHTMX: true},
}

func externalConfig() componentdocshell.Config {
	cfg := docsConfig
	cfg.Navigation.Families = append([]componentdocshell.FamilyLink(nil), docsConfig.Navigation.Families...)
	cfg.Navigation.Families[0].LinkAttrs = templ.Attributes{
		"data-consumer": "family-link",
		"aria-current":  "page",
		"hx-target":     "#consumer-target",
	}
	return cfg
}

func externalPage() componentdocshell.Page {
	return componentdocshell.Page{
		Title:        "Components",
		Description:  "Component documentation overview.",
		ActiveFamily: "components",
		Active:       "overview",
		Content:      templ.Raw(`<h1>Components</h1>`),
	}
}

func renderDocs(w http.ResponseWriter, request *http.Request) {
	page := externalPage()
	view := componentdocshell.Layout(docsConfig, page)
	if request.Header.Get("HX-Request") == "true" {
		view = componentdocshell.Fragment(docsConfig, page)
	}
	if err := view.Render(request.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func render(t *testing.T, view templ.Component) string {
	t.Helper()
	var output bytes.Buffer
	if err := view.Render(context.Background(), &output); err != nil {
		t.Fatalf("render error = %v", err)
	}
	return output.String()
}

func TestPublicFamilyNavigationAPI(t *testing.T) {
	cfg := externalConfig()
	page := externalPage()
	body := render(t, componentdocshell.Layout(cfg, page))

	for _, want := range []string{
		`<h1>Components</h1>`,
		`example.com/componentdocshell`,
		`v0.0.0-example`,
		`href="https://example.com/componentdocshell/releases/v0.0.0-example"`,
		`aria-current="location"`,
		`aria-current="page"`,
		`hx-get="/components"`,
		`hx-target="#main-content"`,
		`data-consumer="family-link"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("layout missing %q", want)
		}
	}
	if strings.Contains(body, `hx-target="#consumer-target"`) {
		t.Fatal("shell did not own the HTMX target")
	}
	if got := strings.Count(body, `aria-current="location"`); got != 2 {
		t.Fatalf("active family marker count = %d, want 2 responsive links", got)
	}
	if got := strings.Count(body, `data-consumer="family-link"`); got != 2 {
		t.Fatalf("consumer attribute count = %d, want 2 responsive links", got)
	}
	for _, family := range []struct {
		label string
		href  string
	}{
		{label: "Components", href: "/components"},
		{label: "Charts", href: "/charts"},
		{label: "App Shells", href: "/app-shells"},
		{label: "Icons", href: "/icons"},
		{label: "LLMs", href: "/llms"},
		{label: "Examples", href: "/examples"},
	} {
		if !strings.Contains(body, `>`+family.label+`</a>`) && !strings.Contains(body, `>`+family.label+`</summary>`) {
			t.Errorf("layout missing family label %q", family.label)
		}
		if !strings.Contains(body, `href="`+family.href+`"`) {
			t.Errorf("layout missing family route %q", family.href)
		}
	}
	if got := cfg.Navigation.Families[0].LinkAttrs["aria-current"]; got != "page" {
		t.Fatalf("caller aria-current mutated to %#v", got)
	}
	if got := cfg.Navigation.Families[0].LinkAttrs["hx-target"]; got != "#consumer-target" {
		t.Fatalf("caller hx-target mutated to %#v", got)
	}

	fragment := render(t, componentdocshell.Fragment(cfg, page))
	for _, target := range []string{
		`<title>Components · Example docs</title>`,
		`outerHTML:#main-content`,
		`outerHTML:#componentdocshell-sidebar-content`,
		`outerHTML:#componentdocshell-family-navigation`,
	} {
		if !strings.Contains(fragment, target) {
			t.Errorf("fragment missing %q", target)
		}
	}
	for _, target := range []string{
		`outerHTML:#main-content`,
		`outerHTML:#componentdocshell-sidebar-content`,
		`outerHTML:#componentdocshell-family-navigation`,
	} {
		if got := strings.Count(fragment, target); got != 1 {
			t.Errorf("fragment target %q count = %d, want 1", target, got)
		}
	}

	if got := shellassets.StylesheetURL("/componentdocshell/assets/"); !strings.HasPrefix(got, "/componentdocshell/assets/shell.css?v=") {
		t.Fatalf("stylesheet URL = %q", got)
	}
	if got := shellassets.ScriptURL("/componentdocshell/assets/"); !strings.HasPrefix(got, "/componentdocshell/assets/shell.js?v=") {
		t.Fatalf("script URL = %q", got)
	}
}
