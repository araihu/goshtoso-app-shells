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
		{"/", "Component docs shell example"},
		{"/components/button", "Button"},
		{"/assets/styles.css", "--color-primary"},
		{"/componentdocshell/assets/shell.css", ".component-doc-shell"},
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
	for _, option := range pages.ShellConfig().Appearance.Themes {
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
		`<img class="component-doc-shell__managed-logo" data-asset-brand="logo" src="/fixtures/brand/logo.svg" alt="Component docs shell" width="120" height="32">`,
		`<button class="component-doc-shell__icon-button component-doc-shell__campaign-toggle" type="button" hidden data-campaign-toggle`,
		`<script defer src="/fixtures/campaign/v1.js" data-channel="/fixtures/releases/current" integrity="sha384-`,
		`crossorigin="anonymous"></script>`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("example presentation contract missing %q", want)
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
	request.Header.Set("HX-Request", "true")
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
