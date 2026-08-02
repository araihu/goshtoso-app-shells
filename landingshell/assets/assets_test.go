package assets

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerServesVersionedAssets(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		url         string
		contentType string
		contains    string
	}{
		{url: StylesheetURL(""), contentType: "text/css; charset=utf-8", contains: ".landing-shell__header"},
		{url: ScriptURL(""), contentType: "text/javascript; charset=utf-8", contains: `Alpine.data("landingShell"`},
	} {
		request := httptest.NewRequest(http.MethodGet, test.url, nil)
		recorder := httptest.NewRecorder()
		Handler().ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d", test.url, recorder.Code)
		}
		if got := recorder.Header().Get("Content-Type"); got != test.contentType {
			t.Errorf("GET %s content type = %q", test.url, got)
		}
		if !strings.Contains(recorder.Body.String(), test.contains) {
			t.Errorf("GET %s body missing %q", test.url, test.contains)
		}
		if !strings.Contains(recorder.Header().Get("Cache-Control"), "immutable") {
			t.Errorf("GET %s is not immutable", test.url)
		}
	}
}

func TestHandlerHonorsCustomPrefix(t *testing.T) {
	t.Parallel()
	request := httptest.NewRequest(http.MethodGet, "/brand-shell/shell.css", nil)
	recorder := httptest.NewRecorder()
	Handler("/brand-shell/").ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("custom prefix status = %d", recorder.Code)
	}
}
