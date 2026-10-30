package e2e

import (
	"fmt"
	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-app-shells/componentdocshell"
	"github.com/araihu/goshtoso-app-shells/example/internal/pages"
	"github.com/mxschmitt/playwright-go"
	"net/http"
	"testing"
)

func serveComposedNavigation(w http.ResponseWriter, r *http.Request) {
	cfg := pages.ShellConfig("components")
	cfg.Navigation.Families = nil
	cfg.Navigation.DisableLocalNavigation = r.URL.Query().Get("local") != "true"
	cfg.Appearance.DisableThemeSelector = true
	cfg.Navigation.GlobalLinks = []componentdocshell.GlobalLink{
		{Label: "Site", Href: "/__e2e/composed", LinkAttrs: templ.Attributes{"hx-get": "/__e2e/composed", "hx-target": "#main-content", "hx-push-url": "true"}},
		{Label: "Docs", Href: "/__e2e/composed?local=true", LinkAttrs: templ.Attributes{"hx-get": "/__e2e/composed?local=true", "hx-target": "#main-content", "hx-push-url": "true"}},
		{Label: "External", Href: "https://example.com/"},
	}
	page := pages.Overview()
	title := "Site"
	if !cfg.Navigation.DisableLocalNavigation {
		title = "Docs"
	}
	page.Content = templ.Raw("<h1>" + title + "</h1>")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	view := componentdocshell.Layout(cfg, page)
	if r.Header.Get("HX-Request") == "true" {
		view = componentdocshell.Fragment(cfg, page)
	}
	if err := view.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func TestComposedNavigationResponsiveAndHTMX(t *testing.T) {
	requireE2E(t)
	h := newBrowserHarness(t)
	for _, width := range []int{390, 1023, 1024, 1440} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			p := h.newPage(t, true)
			if err := p.SetViewportSize(width, 900); err != nil {
				t.Fatal(err)
			}
			gotoFamilyPage(t, p, h.baseURL, "/__e2e/composed")
			mobile := width < 1024
			visible, err := p.Locator(".component-doc-shell__menu-button").IsVisible()
			if err != nil || visible != mobile {
				t.Fatalf("trigger=%v %v", visible, err)
			}
			headerVisible, _ := p.Locator(".component-doc-shell__global-navigation--header").IsVisible()
			if headerVisible == mobile {
				t.Fatal("wrong header projection")
			}
			if _, err := p.Evaluate(`window.composedSentinel=true;window.composedTrigger=document.querySelector('.component-doc-shell__menu-button');window.composedSidebar=document.querySelector('#componentdocshell-sidebar')`); err != nil {
				t.Fatal(err)
			}
			nav := ".component-doc-shell__global-navigation--header"
			if mobile {
				if err := p.Locator(".component-doc-shell__menu-button").Click(); err != nil {
					t.Fatal(err)
				}
				waitForDrawer(t, p, true)
				if _, err := p.Evaluate(`(() => {const color=getComputedStyle(document.querySelector('#componentdocshell-sidebar')).backgroundColor;if(color==='rgba(0, 0, 0, 0)' || color==='transparent')throw Error('transparent drawer');return color})()`); err != nil {
					t.Fatal(err)
				}
				nav = ".component-doc-shell__global-navigation--drawer"
				if _, err := p.WaitForFunction(`document.activeElement===document.querySelector('.component-doc-shell__global-navigation--drawer a')`, nil); err != nil {
					t.Fatal(err)
				}
				if err := p.Keyboard().Press("Shift+Tab"); err != nil {
					t.Fatal(err)
				}
				if _, err := p.WaitForFunction(`document.activeElement===document.querySelector('.component-doc-shell__global-navigation--drawer a:last-child')`, nil); err != nil {
					t.Fatal(err)
				}
				if err := p.Keyboard().Press("Escape"); err != nil {
					t.Fatal(err)
				}
				waitForDrawer(t, p, false)
				if _, err := p.WaitForFunction(`document.activeElement===window.composedTrigger`, nil); err != nil {
					t.Fatal(err)
				}
				if err := p.Locator(".component-doc-shell__menu-button").Click(); err != nil {
					t.Fatal(err)
				}
			}
			if err := p.Locator(nav + " a").Filter(playwright.LocatorFilterOptions{HasText: "Docs"}).Click(); err != nil {
				t.Fatal(err)
			}
			if _, err := p.WaitForFunction(`document.querySelector('#main-content h1')?.textContent==='Docs' && document.querySelector('#componentdocshell-sidebar-content')?.dataset.localNavigation==='true' && window.composedSentinel && window.composedTrigger===document.querySelector('.component-doc-shell__menu-button') && window.composedSidebar===document.querySelector('#componentdocshell-sidebar') && (innerWidth<1024 || getComputedStyle(window.composedSidebar).display!=='none')`, nil); err != nil {
				t.Fatal(err)
			}
			if mobile {
				waitForDrawer(t, p, false)
				if err := p.Locator(".component-doc-shell__menu-button").Click(); err != nil {
					t.Fatal(err)
				}
			}
			if err := p.Locator(nav + " a").Filter(playwright.LocatorFilterOptions{HasText: "Site"}).Click(); err != nil {
				t.Fatal(err)
			}
			if _, err := p.WaitForFunction(`document.querySelector('#main-content h1')?.textContent==='Site' && document.querySelector('#componentdocshell-sidebar-content')?.dataset.localNavigation==='false' && window.composedSentinel && (innerWidth<1024 || (getComputedStyle(window.composedSidebar).display==='none' && window.composedSidebar.inert))`, nil); err != nil {
				t.Fatal(err)
			}
			if _, err := p.GoBack(); err != nil {
				t.Fatal(err)
			}
			if _, err := p.WaitForFunction(`document.querySelector('#main-content h1')?.textContent==='Docs'`, nil); err != nil {
				t.Fatal(err)
			}
			if _, err := p.Evaluate(`(() => {const ids=Array.from(document.querySelectorAll('[id]')).map(n=>n.id);if(new Set(ids).size!==ids.length)throw Error('duplicate IDs');if(document.documentElement.scrollWidth>innerWidth)throw Error('horizontal overflow');return true})()`); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestComposedNavigationWithoutJavaScript(t *testing.T) {
	requireE2E(t)
	h := newBrowserHarness(t)
	p := h.newPage(t, false)
	if err := p.SetViewportSize(390, 900); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Goto(h.baseURL + "/__e2e/composed"); err != nil {
		t.Fatal(err)
	}
	if err := p.Locator(".component-doc-shell__global-navigation--header a").Filter(playwright.LocatorFilterOptions{HasText: "Docs"}).Click(); err != nil {
		t.Fatal(err)
	}
	if text, err := p.Locator("h1").InnerText(); err != nil || text != "Docs" {
		t.Fatalf("native navigation: %q %v", text, err)
	}
}
