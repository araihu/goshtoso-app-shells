package server

import (
	"crypto/sha512"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/araihu/goshtoso-app-shells/example/internal/pages"
)

func TestRoutesAndAssets(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		path string
		want string
	}{
		{path: "/", want: "Components"},
		{path: "/components/button", want: "Button"},
		{path: "/assets/styles.css", want: "--color-primary"},
		{path: "/componentdocshell/assets/shell.css", want: ".component-doc-shell"},
	} {
		recorder := httptest.NewRecorder()
		New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))
		if recorder.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want 200", test.path, recorder.Code)
		}
		if !strings.Contains(recorder.Body.String(), test.want) {
			t.Errorf("GET %s missing %q", test.path, test.want)
		}
	}
}

func TestRepresentativeRoutesRenderCompleteSocialMetadata(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		path        string
		title       string
		description string
	}{
		{path: "/components", title: "Components Documentation - Goshtoso", description: "Components documentation family overview."},
		{path: "/components/button", title: "Button Component - Goshtoso UI Library for Go", description: "Build Go button components with variants, accessible states, and HTMX-ready server rendering."},
	} {
		recorder := httptest.NewRecorder()
		New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))
		if recorder.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d, want 200", test.path, recorder.Code)
		}
		body := recorder.Body.String()
		for _, want := range []string{
			`<title>` + test.title + `</title>`,
			`<meta name="description" content="` + test.description + `">`,
			`<link rel="canonical" href="https://goshtoso.araihu.com` + test.path + `">`,
			`<meta property="og:title" content="` + test.title + `">`,
			`<meta property="og:description" content="` + test.description + `">`,
			`<meta property="og:image" content="https://goshtoso.araihu.com/assets/images/goshtoso-social-card.png">`,
			`<meta name="twitter:card" content="summary_large_image">`,
			`<meta name="twitter:title" content="` + test.title + `">`,
			`<meta name="twitter:image" content="https://goshtoso.araihu.com/assets/images/goshtoso-social-card.png">`,
		} {
			if got := strings.Count(body, want); got != 1 {
				t.Errorf("GET %s metadata %q count = %d, want 1", test.path, want, got)
			}
		}
	}
}

func TestFamilyRoutesRenderScopedShells(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		path   string
		family string
	}{
		{path: "/components", family: "Components"},
		{path: "/charts", family: "Charts"},
		{path: "/app-shells", family: "App Shells"},
		{path: "/icons", family: "Icons"},
		{path: "/llms", family: "LLMs"},
		{path: "/examples", family: "Examples"},
	} {
		t.Run(test.family, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))
			if recorder.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want 200", test.path, recorder.Code)
			}
			body := recorder.Body.String()
			if !strings.Contains(body, `href="`+test.path+`" aria-current="location"`) {
				t.Errorf("GET %s missing active family %q", test.path, test.family)
			}
			if !strings.Contains(body, `<h1`) || !strings.Contains(body, test.family) {
				t.Errorf("GET %s missing family overview content", test.path)
			}
		})
	}
}

func TestFamilyRoutesAreExact(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/charts/extra", nil))
	if recorder.Code != http.StatusNotFound {
		t.Errorf("GET /charts/extra status = %d, want 404", recorder.Code)
	}
}

func TestFamilyShellConfigTruthFixtures(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		family      string
		module      string
		moduleLabel string
		moduleURL   string
		version     string
		versionURL  string
	}{
		{family: "components", module: "github.com/araihu/goshtoso", moduleLabel: "araihu/goshtoso", moduleURL: "https://github.com/araihu/goshtoso", version: "v0.1.6", versionURL: "https://github.com/araihu/goshtoso/releases/tag/v0.1.6"},
		{family: "charts", module: "github.com/araihu/goshtoso-charts", moduleLabel: "araihu/goshtoso-charts", moduleURL: "https://github.com/araihu/goshtoso-charts", version: "v0.0.1", versionURL: "https://github.com/araihu/goshtoso-charts/releases/tag/v0.0.1"},
		{family: "app-shells", module: "github.com/araihu/goshtoso-app-shells", moduleLabel: "araihu/goshtoso-app-shells", moduleURL: "https://github.com/araihu/goshtoso-app-shells"},
		{family: "icons"},
		{family: "llms"},
		{family: "examples"},
	} {
		config := pages.ShellConfig(test.family)
		if config.Navigation.Scope == nil {
			t.Fatalf("ShellConfig(%q) scope is nil", test.family)
		}
		if got := config.Navigation.Scope.ModulePath; got != test.module {
			t.Errorf("ShellConfig(%q) module = %q, want %q", test.family, got, test.module)
		}
		if got := config.Navigation.Scope.ModuleLabel; got != test.moduleLabel {
			t.Errorf("ShellConfig(%q) module label = %q, want %q", test.family, got, test.moduleLabel)
		}
		if got := config.Navigation.Scope.ModuleURL; got != test.moduleURL {
			t.Errorf("ShellConfig(%q) module URL = %q, want %q", test.family, got, test.moduleURL)
		}
		if got := config.Navigation.Scope.Version; got != test.version {
			t.Errorf("ShellConfig(%q) version = %q, want %q", test.family, got, test.version)
		}
		if got := config.Navigation.Scope.VersionURL; got != test.versionURL {
			t.Errorf("ShellConfig(%q) version URL = %q, want %q", test.family, got, test.versionURL)
		}
		if !config.Interactions.LocalRuntime {
			t.Errorf("ShellConfig(%q) local runtime is disabled", test.family)
		}
		if !config.Appearance.DisableThemeSelector {
			t.Errorf("ShellConfig(%q) exposes the theme selector", test.family)
		}
		if config.Brand.Logo == nil || config.Brand.CompactLogo == nil {
			t.Errorf("ShellConfig(%q) does not use the canonical mark on both responsive surfaces", test.family)
		}
		if config.Brand.ManagedLogo != nil {
			t.Errorf("ShellConfig(%q) still uses the managed wordmark", test.family)
		}
	}
}

func TestFamilyShellConfigReturnsFreshNavigationSlices(t *testing.T) {
	t.Parallel()
	first := pages.ShellConfig("components")
	first.Navigation.Families[0].Label = "Changed family"
	first.Navigation.Items[0].Label = "Changed item"
	first.Navigation.Sections[0].Items[0].Label = "Changed section item"

	second := pages.ShellConfig("components")
	if got := second.Navigation.Families[0].Label; got != "Components" {
		t.Errorf("fresh family label = %q, want Components", got)
	}
	if got := second.Navigation.Items[0].Label; got != "Overview" {
		t.Errorf("fresh overview label = %q, want Overview", got)
	}
	if got := second.Navigation.Sections[0].Items[0].Label; got != "Button" {
		t.Errorf("fresh section item label = %q, want Button", got)
	}
}

func TestFamilyHTMXResponseIsAtomic(t *testing.T) {
	t.Parallel()
	request := httptest.NewRequest(http.MethodGet, "/charts", nil)
	request.Header.Set("HX-Request-Type", "partial")
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, request)
	body := recorder.Body.String()
	for _, target := range []string{"#main-content", "#componentdocshell-sidebar-content", "#componentdocshell-family-navigation"} {
		if got := strings.Count(body, "outerHTML:"+target); got != 1 {
			t.Errorf("HTMX target %s count = %d, want 1", target, got)
		}
	}
}

func TestExampleUsesAraiHuThemeByDefault(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	body := recorder.Body.String()
	for _, want := range []string{`"theme":"araihu"`, `/componentdocshell/assets/araihu.css`} {
		if !strings.Contains(body, want) {
			t.Errorf("example default theme contract missing %q", want)
		}
	}
	foundAraiHu := false
	for _, option := range pages.ShellConfig("components").Appearance.Themes {
		if option.Value == "araihu" && option.Label == "Arai Hû" {
			foundAraiHu = true
		}
	}
	if !foundAraiHu {
		t.Error("example configuration misses the Arai Hû theme label")
	}
}

func TestExamplePresentationChannelContract(t *testing.T) {
	t.Parallel()
	handler := New()
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200", recorder.Code)
	}
	body := recorder.Body.String()
	for _, want := range []string{
		`var source="default"`,
		`document.documentElement.dataset.themeSource=source`,
		`goshtoso-mark.svg`,
		`goshtoso-mark-reverse.svg`,
		`<link rel="icon" data-asset-brand="icon"`,
		`<button class="component-doc-shell__icon-button component-doc-shell__campaign-toggle" type="button" hidden data-campaign-toggle`,
		`<script defer src="/fixtures/campaign/v1.js" data-channel="/fixtures/releases/current" integrity="sha384-`,
		`crossorigin="anonymous"></script>`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("example presentation contract missing %q", want)
		}
	}
	for _, unwanted := range []string{`id="componentdocshell-theme-trigger"`, `id="componentdocshell-theme-mobile-trigger"`, `component-doc-shell__managed-logo`} {
		if strings.Contains(body, unwanted) {
			t.Errorf("example shell still renders %q", unwanted)
		}
	}

	hash := sha512.Sum384([]byte(fixtureRuntime))
	wantIntegrity := "sha384-" + base64.StdEncoding.EncodeToString(hash[:])
	if wantIntegrity != "sha384-wCfS7WktEJUyXXUJVX+Hd2vWoWqU3XX2oZeezp7QLgRz8hoI9iBLdBKhPGBfddvH" {
		t.Fatalf("fixture runtime integrity = %q, want committed SRI", wantIntegrity)
	}
	if !strings.Contains(body, `integrity="`+wantIntegrity+`"`) {
		t.Errorf("example runtime missing fixture integrity %q", wantIntegrity)
	}

	for _, test := range []struct {
		path string
		want string
	}{
		{"/fixtures/campaign/v1.js", fixtureRuntime},
		{"/fixtures/releases/current", fixtureChannel},
		{"/fixtures/brand/logo.svg", fixtureLogo},
	} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))
		if recorder.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want 200", test.path, recorder.Code)
		}
		if recorder.Body.String() != test.want {
			t.Errorf("GET %s body = %q, want fixture bytes", test.path, recorder.Body.String())
		}
	}
}

func TestHTMXRequestReturnsFragment(t *testing.T) {
	t.Parallel()
	request := httptest.NewRequest(http.MethodGet, "/components/button", nil)
	request.Header.Set("HX-Request-Type", "partial")
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, request)
	body := recorder.Body.String()
	if strings.Contains(body, "<html") {
		t.Fatal("HTMX response contains complete document")
	}
	for _, want := range []string{`id="main-content"`, `hx-swap-oob="outerHTML:#componentdocshell-sidebar-content"`} {
		if !strings.Contains(body, want) {
			t.Errorf("HTMX response missing %q", want)
		}
	}
}
