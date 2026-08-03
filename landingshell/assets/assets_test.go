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

func TestHandlerKeepsUnversionedAssetsRevalidatable(t *testing.T) {
	t.Parallel()
	for _, assetURL := range []string{"/landingshell/assets/shell.css", "/landingshell/assets/shell.js"} {
		request := httptest.NewRequest(http.MethodGet, assetURL, nil)
		recorder := httptest.NewRecorder()
		Handler().ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d", assetURL, recorder.Code)
		}
		cacheControl := recorder.Header().Get("Cache-Control")
		if !strings.Contains(cacheControl, "must-revalidate") || strings.Contains(cacheControl, "immutable") {
			t.Errorf("GET %s cache control = %q", assetURL, cacheControl)
		}
	}
}

func TestHandlerRejectsStaleAssetVersions(t *testing.T) {
	t.Parallel()
	for _, assetURL := range []string{"/landingshell/assets/shell.css?v=stale", "/landingshell/assets/shell.js?v=stale"} {
		request := httptest.NewRequest(http.MethodGet, assetURL, nil)
		recorder := httptest.NewRecorder()
		Handler().ServeHTTP(recorder, request)
		if recorder.Code != http.StatusNotFound {
			t.Errorf("GET %s status = %d, want %d", assetURL, recorder.Code, http.StatusNotFound)
		}
		cacheControl := recorder.Header().Get("Cache-Control")
		if cacheControl != "no-store" || strings.Contains(cacheControl, "immutable") {
			t.Errorf("GET %s cache control = %q", assetURL, cacheControl)
		}
	}
}

func TestHandlerHonorsCustomPrefix(t *testing.T) {
	t.Parallel()
	request := httptest.NewRequest(http.MethodGet, StylesheetURL("/brand-shell"), nil)
	recorder := httptest.NewRecorder()
	Handler("/brand-shell/").ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("custom prefix status = %d", recorder.Code)
	}
}
