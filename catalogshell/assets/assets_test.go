package assets

import (
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
		{"/catalogshell/assets/shell.css", "text/css", ".catalog-shell"},
		{"/catalogshell/assets/shell.js", "text/javascript", "catalogShell"},
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

func TestHandlerRejectsUnknownAndTraversalPaths(t *testing.T) {
	t.Parallel()
	for _, path := range []string{
		"/catalogshell/assets/missing.css",
		"/catalogshell/assets/../go.mod",
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
}

func TestHandlerSupportsCustomPrefix(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	Handler("/custom/").ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/custom/shell.css", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET custom shell.css status = %d", recorder.Code)
	}
}
