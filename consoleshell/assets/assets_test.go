package assets

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerServesVersionedRuntimeAssets(t *testing.T) {
	t.Parallel()
	for _, path := range []string{StylesheetURL(""), ScriptURL("")} {
		r := httptest.NewRecorder()
		Handler().ServeHTTP(r, httptest.NewRequest(http.MethodGet, path, nil))
		if r.Code != http.StatusOK {
			t.Fatalf("%s status=%d", path, r.Code)
		}
		if cache := r.Header().Get("Cache-Control"); !strings.Contains(cache, "immutable") {
			t.Fatalf("%s cache control=%q, want immutable", path, cache)
		}
	}
}

func TestHandlerDoesNotCacheUnversionedOrStaleAssetsAsImmutable(t *testing.T) {
	t.Parallel()

	unversioned := httptest.NewRecorder()
	Handler().ServeHTTP(unversioned, httptest.NewRequest(http.MethodGet, "/consoleshell/assets/shell.css", nil))
	if unversioned.Code != http.StatusOK {
		t.Fatalf("unversioned status=%d", unversioned.Code)
	}
	if cache := unversioned.Header().Get("Cache-Control"); strings.Contains(cache, "immutable") {
		t.Fatalf("unversioned cache control=%q", cache)
	}

	stale := httptest.NewRecorder()
	Handler().ServeHTTP(stale, httptest.NewRequest(http.MethodGet, "/consoleshell/assets/shell.js?v=stale", nil))
	if stale.Code != http.StatusNotFound {
		t.Fatalf("stale status=%d, want %d", stale.Code, http.StatusNotFound)
	}
	if cache := stale.Header().Get("Cache-Control"); strings.Contains(cache, "immutable") {
		t.Fatalf("stale cache control=%q", cache)
	}
}
func TestRuntimeLifecycleContract(t *testing.T) {
	t.Parallel()
	source := string(script)
	for _, want := range []string{"htmx:beforeSwap", "htmx:afterSettle", "htmx:historyRestore", "sidebarScrollTop", "main.scrollTo({top:0})", "focusMain(main)", "reconcileNavigation(main)", "closeDrawer", "restoreFocus === false", "window.__consoleShellLifecycleInstalled", "event.target.matches(\"main.console-shell__main\")", "window.__consoleShellAlpineRegistered", "htmx.process", "document.title", "popstate"} {
		if want == "htmx.process" {
			if strings.Contains(source, want) {
				t.Fatal("runtime must not manually reinitialize htmx-swapped content")
			}
			continue
		}
		if want == "document.title" || want == "popstate" {
			if strings.Contains(source, want) {
				t.Fatalf("runtime must leave %s lifecycle ownership to HTMX", want)
			}
			continue
		}
		if !strings.Contains(source, want) {
			t.Errorf("runtime missing %q", want)
		}
	}
}

func TestNoJavaScriptNavigationContract(t *testing.T) {
	t.Parallel()
	css := string(stylesheet)
	for _, want := range []string{".console-shell-root.js .console-shell__sidebar", ".console-shell-root:not(.js) .console-shell__frame", ".console-shell-root:not(.js) .console-shell__menu"} {
		if !strings.Contains(css, want) {
			t.Errorf("stylesheet missing %q", want)
		}
	}
}
