package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRoutesAndAssets(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		path string
		want string
	}{
		{"/", "Catalog shell example"},
		{"/components/button", "Button"},
		{"/assets/styles.css", "--color-primary"},
		{"/catalogshell/assets/shell.css", ".catalog-shell"},
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
	for _, want := range []string{`id="main-content"`, `hx-swap-oob="outerHTML:#catalogshell-sidebar-content"`} {
		if !strings.Contains(body, want) {
			t.Errorf("HTMX response missing %q", want)
		}
	}
}
