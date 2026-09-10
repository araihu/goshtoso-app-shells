package e2e

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-app-shells/consoleshell"
	shellassets "github.com/araihu/goshtoso-app-shells/consoleshell/assets"
	"github.com/araihu/goshtoso/assets"
	"github.com/araihu/goshtoso/components/sidebar"
)

func TestConsoleHTMX4HistoryAndFocus(t *testing.T) {
	requireE2E(t)
	mux := http.NewServeMux()
	mux.Handle("GET /assets/", assets.Handler())
	mux.Handle("GET /consoleshell/assets/", shellassets.Handler())
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		cfg := consoleshell.Config{
			Brand: consoleshell.Brand{Name: "Console", HomeURL: "/"},
			Navigation: consoleshell.Navigation{Items: []sidebar.Item{
				{ID: "home", Label: "Home", Href: "/"},
				{ID: "staging", Label: "Staging", Href: "/staging"},
			}},
			Interactions: consoleshell.InteractionConfig{LocalRuntime: true, EnableHTMX: true},
		}
		active, title := "home", "Home"
		if r.URL.Path == "/staging" {
			active, title = "staging", "Staging"
		}
		page := consoleshell.Page{Title: title, Active: active, Content: templ.Raw("<h1>" + title + "</h1>")}
		component := consoleshell.Layout(cfg, page)
		if r.Header.Get("HX-Request-Type") == "partial" {
			component = consoleshell.Fragment(cfg, page)
		}
		if err := component.Render(r.Context(), w); err != nil {
			t.Error(err)
		}
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	harness := newBrowserHarness(t)
	page := harness.newPage(t, true)
	if err := page.SetViewportSize(1440, 900); err != nil {
		t.Fatal(err)
	}
	failures := watchBrowserFailures(page)
	gotoFamilyPage(t, page, srv.URL, "/")
	if err := page.Locator("#consoleshell-sidebar a[href='/staging']").Click(); err != nil {
		t.Fatal(err)
	}
	assertConsoleState := func(active string) {
		t.Helper()
		if _, err := page.WaitForFunction(`active => {
const main = document.querySelector('main.console-shell__main');
const link = document.querySelector('[data-consoleshell-nav-id="' + active + '"]');
return main?.dataset.activeNavigation === active && link?.getAttribute('aria-current') === 'page' && document.activeElement === main.querySelector('h1');
}`, active); err != nil {
			t.Fatal(err)
		}
	}
	assertConsoleState("staging")
	if _, err := page.GoBack(); err != nil {
		t.Fatal(err)
	}
	assertConsoleState("home")
	if _, err := page.GoForward(); err != nil {
		t.Fatal(err)
	}
	assertConsoleState("staging")
	if problems := failures.snapshot(); len(problems) > 0 {
		t.Fatalf("browser errors: %v", problems)
	}
}
