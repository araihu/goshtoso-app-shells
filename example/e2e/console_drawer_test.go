package e2e

import (
	"fmt"
	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-app-shells/consoleshell"
	shellassets "github.com/araihu/goshtoso-app-shells/consoleshell/assets"
	"github.com/araihu/goshtoso/assets"
	"github.com/araihu/goshtoso/components/sidebar"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConsoleDrawerMenuAtEveryWidth(t *testing.T) {
	requireE2E(t)
	mux := http.NewServeMux()
	mux.Handle("GET /assets/", assets.Handler())
	mux.Handle("GET /consoleshell/assets/", shellassets.Handler())
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		cfg := consoleshell.Config{
			Brand:        consoleshell.Brand{Name: "Console", HomeURL: "/"},
			Navigation:   consoleshell.Navigation{Drawer: true, IconOnlyMenu: true, DisableSearch: true, Items: []sidebar.Item{{ID: "home", Label: "Home", Href: "/"}, {ID: "staging", Label: "Staging", Href: "/staging"}}},
			Appearance:   consoleshell.AppearanceConfig{DefaultTheme: r.URL.Query().Get("theme"), InitialColorScheme: consoleshell.ColorScheme(r.URL.Query().Get("scheme"))},
			Interactions: consoleshell.InteractionConfig{LocalRuntime: true},
		}
		if err := consoleshell.Layout(cfg, consoleshell.Page{Title: "Services", Content: templ.Raw("<h1>Services</h1>")}).Render(r.Context(), w); err != nil {
			t.Error(err)
		}
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	harness := newBrowserHarness(t)
	for _, width := range []int{390, 1440} {
		for _, theme := range []string{"goshtoso", "minimal"} {
			for _, scheme := range []string{"light", "dark"} {
				t.Run(fmt.Sprintf("%d/%s/%s", width, theme, scheme), func(t *testing.T) {
					page := harness.newPage(t, true)
					if err := page.SetViewportSize(width, 900); err != nil {
						t.Fatal(err)
					}
					gotoFamilyPage(t, page, srv.URL, "/?theme="+theme+"&scheme="+scheme)
					if _, err := page.WaitForFunction(`() => document.querySelector('#consoleshell-menu').getAttribute('aria-expanded') === 'false'`, nil); err != nil {
						t.Fatal(err)
					}
					if value, err := page.Locator("#main-content").Evaluate("el => el.getBoundingClientRect().left === 0", nil); err != nil || value != true {
						t.Fatalf("main left = %v, %v", value, err)
					}
					menu := page.Locator("#consoleshell-menu")
					if err := menu.Click(); err != nil {
						t.Fatal(err)
					}
					if _, err := page.WaitForFunction(`() => document.querySelector('#consoleshell-menu').getAttribute('aria-expanded') === 'true'`, nil); err != nil {
						t.Fatal(err)
					}
					if err := page.Locator("#consoleshell-sidebar a").First().Press("Escape"); err != nil {
						t.Fatal(err)
					}
					if _, err := page.WaitForFunction(`() => document.activeElement.id === 'consoleshell-menu' && document.activeElement.getAttribute('aria-expanded') === 'false'`, nil); err != nil {
						t.Fatal(err)
					}
				})
			}
		}
	}
}
