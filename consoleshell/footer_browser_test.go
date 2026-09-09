package consoleshell

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/a-h/templ"
	shellassets "github.com/araihu/goshtoso-app-shells/consoleshell/assets"
	"github.com/araihu/goshtoso/assets"
	"github.com/mxschmitt/playwright-go"
)

func TestFooterBottomBrowser(t *testing.T) {
	if os.Getenv("COMPONENTDOCSHELL_E2E") != "1" {
		t.Skip("set COMPONENTDOCSHELL_E2E=1")
	}
	mux := http.NewServeMux()
	mux.Handle("/assets/", assets.Handler())
	mux.Handle("/consoleshell/assets/", shellassets.Handler())
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		cfg := validConfig()
		cfg.Interactions.LocalRuntime = true
		cfg.Footer = templ.Raw("<span>Footer</span>")
		page := validPage()
		height := 80
		if r.URL.Path == "/long" {
			height = 1500
		}
		page.Content = templ.Raw(fmt.Sprintf(`<div style="height:%dpx">Content</div>`, height))
		if err := Layout(cfg, page).Render(r.Context(), w); err != nil {
			t.Error(err)
		}
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	pw, err := playwright.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer pw.Stop()
	browser, err := pw.Chromium.Launch()
	if err != nil {
		t.Fatal(err)
	}
	defer browser.Close()
	page, err := browser.NewPage()
	if err != nil {
		t.Fatal(err)
	}
	for _, width := range []int{390, 831, 1440} {
		if err := page.SetViewportSize(width, 994); err != nil {
			t.Fatal(err)
		}
		for _, route := range []string{"/", "/long"} {
			if _, err := page.Goto(server.URL + route); err != nil {
				t.Fatal(err)
			}
			_, err := page.WaitForFunction(`()=>document.documentElement.classList.contains('js')`, nil)
			if err != nil {
				t.Fatal(err)
			}
			fits, err := page.Evaluate(`long => {const m=document.querySelector('main');const c=m.querySelector('.console-shell__content');const f=m.querySelector('footer');m.scrollTop=m.scrollHeight;const r=f.getBoundingClientRect();return Math.abs(r.bottom-(innerHeight-24))<1 && r.top>=c.getBoundingClientRect().bottom && (long ? m.scrollHeight>m.clientHeight : m.scrollHeight===m.clientHeight);}`, route == "/long")
			if err != nil || fits != true {
				t.Fatalf("footer layout %d %s: %v %v", width, route, fits, err)
			}
		}
	}
}
