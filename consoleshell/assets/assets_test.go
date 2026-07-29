package assets

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerServesVersionedRuntimeAssets(t *testing.T) {
	t.Parallel()
	for _, path := range []string{"/consoleshell/assets/shell.css", "/consoleshell/assets/shell.js"} {
		r := httptest.NewRecorder()
		Handler().ServeHTTP(r, httptest.NewRequest(http.MethodGet, path, nil))
		if r.Code != http.StatusOK {
			t.Fatalf("%s status=%d", path, r.Code)
		}
		if r.Header().Get("Cache-Control") == "" {
			t.Fatal("missing cache control")
		}
	}
}
func TestRuntimeLifecycleContract(t *testing.T) {
	t.Parallel()
	source := string(script)
	for _, want := range []string{"htmx:beforeSwap", "htmx:afterSettle", "htmx:historyRestore", "sidebarScrollTop", "main.scrollTo({top:0})", "focusMain(main)", "reconcileNavigation(main)", "closeDrawer", "restoreFocus === false", "listenersInstalled", "window.__consoleShellAlpineRegistered", "htmx.process", "document.title", "popstate"} {
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
