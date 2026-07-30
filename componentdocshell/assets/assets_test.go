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
		{"/componentdocshell/assets/goshtoso-logo.svg", "image/svg+xml", "<svg"},
		{"/componentdocshell/assets/goshtoso-mark.svg", "image/svg+xml", "<svg"},
		{"/componentdocshell/assets/goshtoso-mark-reverse.svg", "image/svg+xml", "<svg"},
		{"/componentdocshell/assets/goshtoso-favicon.svg", "image/svg+xml", "<svg"},
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

func TestAraiHuThemeIncludesAdaptiveLogoContract(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/componentdocshell/assets/araihu.css", nil))
	for _, want := range []string{"--araihu-logo-surface", "--araihu-logo-ink", "--araihu-logo-signal", `.dark [data-theme="araihu"]`} {
		if !strings.Contains(recorder.Body.String(), want) {
			t.Errorf("Arai Hû theme missing V11 contract %q", want)
		}
	}
}

func TestReleasedGoshtosoFallbackHashes(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name     string
		contents []byte
		want     string
	}{
		{"logo", goshtosoLogo, "5801b31fc6b1f54cde98b1a3f3f5e57553f6e67aa3fa0318879e5e2603cd540e"},
		{"mark", goshtosoMark, "150741f362c418b541a1d05e684b26dcd46ebfe50a11d530a81c244de7231c17"},
		{"reverse mark", goshtosoMarkReverse, "1877530c7ea23f9c597caf064e2596de5d83b78ff8795bc73a80e51d2770471e"},
		{"favicon", goshtosoFavicon, "56e8b185f2572ad4c7ea6fa8e715aa7dacd7381422b431ce07c35965ab05b3b7"},
	} {
		sum := sha256.Sum256(test.contents)
		if got := hex.EncodeToString(sum[:]); got != test.want {
			t.Errorf("%s SHA-256 = %s, want released %s", test.name, got, test.want)
		}
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

func TestShellStylesUseBoundedAnchorScrollPadding(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/componentdocshell/assets/shell.css", nil))
	body := recorder.Body.String()
	for _, want := range []string{
		`.component-doc-shell__main-scroll {`,
		`scroll-padding-block: 2rem`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("shell stylesheet missing bounded anchor scroll padding %q", want)
		}
	}
	if strings.Contains(body, `.component-doc-shell__main::after {`) {
		t.Error("shell stylesheet must not add a viewport-sized pseudo-element after page content")
	}
}

func TestShellRuntimeExposesThemeSetter(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/componentdocshell/assets/shell.js", nil))
	body := recorder.Body.String()
	for _, want := range []string{`setTheme: function (value)`, `root.dataset.themeSource === "preference"`, `document.documentElement.dataset.themeSource = "preference"`, `this.theme = value`, `if (!self.persistTheme) return;`} {
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
