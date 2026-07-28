package assets

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerServesEmbeddedAssetsAtStablePaths(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		path        string
		contentType string
		contains    string
	}{
		{"/componentdocshell/assets/shell.css", "text/css", ".component-doc-shell"},
		{"/componentdocshell/assets/shell.js", "text/javascript", "componentDocShell"},
		{"/componentdocshell/assets/araihu.css", "text/css", `[data-theme="araihu"]`},
		{"/componentdocshell/assets/goshtoso-logo.svg", "image/svg+xml", "Goshtoso outlined logo"},
		{"/componentdocshell/assets/goshtoso-mark.svg", "image/svg+xml", "Goshtoso mark"},
		{"/componentdocshell/assets/goshtoso-mark-reverse.svg", "image/svg+xml", "Goshtoso reverse mark"},
		{"/componentdocshell/assets/goshtoso-favicon.svg", "image/svg+xml", "Goshtoso favicon"},
	} {
		recorder := httptest.NewRecorder()
		Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))
		if recorder.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want 200", test.path, recorder.Code)
		}
		if got := recorder.Header().Get("Content-Type"); !strings.HasPrefix(got, test.contentType) {
			t.Errorf("GET %s content-type = %q, want prefix %q", test.path, got, test.contentType)
		}
		if !strings.Contains(recorder.Body.String(), test.contains) {
			t.Errorf("GET %s body missing %q", test.path, test.contains)
		}
		if got := recorder.Header().Get("X-Content-Type-Options"); got != "nosniff" {
			t.Errorf("GET %s X-Content-Type-Options = %q", test.path, got)
		}
	}
}

func TestCanonicalGoshtosoFaviconHash(t *testing.T) {
	t.Parallel()
	sum := sha256.Sum256(goshtosoFavicon)
	if got := hex.EncodeToString(sum[:]); got != "e895783a026d60532430a6aba6e2ca70931993f2a1370c46fc14de642590f47a" {
		t.Fatalf("favicon SHA-256 = %s", got)
	}
}

func TestShellStylesOwnComponentPageComposition(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/componentdocshell/assets/shell.css", nil))
	body := recorder.Body.String()
	for _, want := range []string{
		`.component-page__example-body > :not([hidden]) ~ :not([hidden])`,
		`margin-top: 1rem`,
		`.component-page__preview::after {`,
		`inset: 0`,
		`border: 1px solid var(--color-outline)`,
		`border-radius: var(--radius-radius)`,
		`.dark .component-page__preview::after {`,
		`border-color: var(--color-outline-dark)`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("shell stylesheet missing component-page composition contract %q", want)
		}
	}
}

func TestShellStylesContainDocumentScrolling(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/componentdocshell/assets/shell.css", nil))
	body := recorder.Body.String()
	for _, want := range []string{
		`.component-doc-shell-root {`,
		`overflow: hidden`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("shell stylesheet missing root scroll containment %q", want)
		}
	}
}

func TestShellStylesReserveTrailingAnchorSlack(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/componentdocshell/assets/shell.css", nil))
	body := recorder.Body.String()
	for _, want := range []string{
		`.component-doc-shell__main::after {`,
		`height: max(0px, calc(100vh - 10rem))`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("shell stylesheet missing trailing anchor slack %q", want)
		}
	}
}

func TestShellRuntimeExposesThemeSetter(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/componentdocshell/assets/shell.js", nil))
	body := recorder.Body.String()
	for _, want := range []string{`setTheme: function (value)`, `this.theme = value`} {
		if !strings.Contains(body, want) {
			t.Errorf("shell runtime missing theme setter contract %q", want)
		}
	}
}

func TestShellRuntimeUsesTOCRolesAndLegacyLinkHook(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/componentdocshell/assets/shell.js", nil))
	body := recorder.Body.String()
	for _, want := range []string{
		`[data-componentdocshell-toc]`,
		`[data-componentdocshell-toc-list]`,
		`link.setAttribute("data-toc-link", heading.id)`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("shell runtime missing TOC contract %q", want)
		}
	}
}

func TestShellRuntimeAlignsHashInsideMainScroller(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/componentdocshell/assets/shell.js", nil))
	body := recorder.Body.String()
	for _, want := range []string{
		`function scrollTarget(target, behavior)`,
		`document.documentElement.scrollTop = 0`,
		`document.body.scrollTop = 0`,
		`scroller.scrollTo({ top: nextTop, behavior: behavior || "auto" })`,
		`requestAnimationFrame(function () { scrollTarget(active, "auto"); })`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("shell runtime missing hash alignment %q", want)
		}
	}
}

func TestHandlerRejectsUnknownAndTraversalPaths(t *testing.T) {
	t.Parallel()
	for _, path := range []string{
		"/componentdocshell/assets/missing.css",
		"/componentdocshell/assets/../go.mod",
		"/shell.css",
	} {
		recorder := httptest.NewRecorder()
		Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		if recorder.Code != http.StatusNotFound {
			t.Errorf("GET %s status = %d, want 404", path, recorder.Code)
		}
	}
}

func TestAssetURLsRespectPrefix(t *testing.T) {
	t.Parallel()
	if got := StylesheetURL("/custom/"); !strings.HasPrefix(got, "/custom/shell.css?v=") {
		t.Errorf("StylesheetURL() = %q", got)
	}
	if got := ScriptURL("/custom/"); !strings.HasPrefix(got, "/custom/shell.js?v=") {
		t.Errorf("ScriptURL() = %q", got)
	}
	if got := AraiHuThemeURL("/custom/"); !strings.HasPrefix(got, "/custom/araihu.css?v=") {
		t.Errorf("AraiHuThemeURL() = %q", got)
	}
}

func TestHandlerSupportsCustomPrefix(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	Handler("/custom/").ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/custom/shell.css", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET custom shell.css status = %d", recorder.Code)
	}
}
