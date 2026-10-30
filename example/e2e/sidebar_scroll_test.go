package e2e

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-app-shells/componentdocshell"
	"github.com/araihu/goshtoso-app-shells/example/internal/pages"
	"github.com/araihu/goshtoso/components/sidebar"
)

func sidebarScrollFixture(writer http.ResponseWriter, request *http.Request) {
	cfg := pages.ShellConfig("components")
	cfg.Navigation.Families = nil
	cfg.Navigation.Sections = nil
	cfg.Navigation.Items = nil
	for i := 0; i < 40; i++ {
		id := fmt.Sprint(i)
		cfg.Navigation.Items = append(cfg.Navigation.Items, sidebar.Item{ID: id, Label: "Page " + id, Href: "/__e2e/sidebar/" + id})
	}
	cfg.Interactions.EnableHTMX = true
	cfg.Interactions.LocalRuntime = true
	page, _ := pages.FamilyOverview("components")
	page.Active = request.PathValue("item")
	var content templ.Component = componentdocshell.Layout(cfg, page)
	if request.Header.Get("HX-Request") == "true" {
		content = componentdocshell.Fragment(cfg, page)
	}
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := content.Render(request.Context(), writer); err != nil {
		http.Error(writer, "render sidebar fixture", http.StatusInternalServerError)
	}
}

func TestSidebarScrollSurvivesNavigationAndRevealsOnlyOffscreenItems(t *testing.T) {
	requireE2E(t)
	harness := newBrowserHarness(t)
	page := harness.newPage(t, true)
	if err := page.SetViewportSize(1162, 900); err != nil {
		t.Fatal(err)
	}
	gotoFamilyPage(t, page, harness.baseURL, "/__e2e/sidebar/30")
	visible := `() => {
 const sidebar = document.querySelector('.sidebar-scroll');
 const active = sidebar?.querySelector('a[aria-current="page"]');
 if (!active) return false;
 const view = sidebar.getBoundingClientRect(), item = active.getBoundingClientRect();
 return item.top >= view.top - 1 && item.bottom <= view.bottom + 1;
 }`
	if _, err := page.WaitForFunction(visible, nil); err != nil {
		t.Fatalf("direct load reveals active item: %v", err)
	}
	for _, destination := range []string{"31", "30"} {
		_, err := page.Evaluate(`() => {
   const sidebar = document.querySelector('.sidebar-scroll');
   const first = sidebar.querySelector('a[href="/__e2e/sidebar/30"]');
   sidebar.scrollTop += first.getBoundingClientRect().top-sidebar.getBoundingClientRect().top-32;
   window.sidebarScrollProbe={node:sidebar,start:sidebar.scrollTop,samples:[],done:false};
   document.addEventListener('htmx:after:swap',function after(event){
    if(event.detail.ctx.target.id!=='main-content')return;
    document.removeEventListener('htmx:after:swap',after,true);
    const sample=()=>sidebarScrollProbe.samples.push({same:document.querySelector('.sidebar-scroll')===sidebar,top:sidebar.scrollTop});
    sample();requestAnimationFrame(()=>{sample();requestAnimationFrame(()=>{sample();sidebarScrollProbe.done=true})});
   },true);
  }`)
		if err != nil {
			t.Fatal(err)
		}
		if err := page.Locator(`.sidebar-scroll a[href="/__e2e/sidebar/` + destination + `"]`).Click(); err != nil {
			t.Fatal(err)
		}
		if _, err := page.WaitForFunction(`() => sidebarScrollProbe.done`, nil); err != nil {
			t.Fatal(err)
		}
		result, err := page.Evaluate(`destination=>sidebarScrollProbe.start>0 && sidebarScrollProbe.samples.every(sample=>sample.same&&Math.abs(sample.top-sidebarScrollProbe.start)<=1) && document.querySelector('.sidebar-scroll a[aria-current="page"]').getAttribute('href')==='/__e2e/sidebar/'+destination`, destination)
		if err != nil {
			t.Fatal(err)
		}
		if result != true {
			t.Fatalf("navigation to %s changed the scroller or its position: %v", destination, result)
		}
	}
	_, err := page.Evaluate(`async()=>{document.querySelector('.sidebar-scroll').scrollTop=0;await htmx.ajax('GET','/__e2e/sidebar/30',{target:'#main-content',swap:'innerHTML'})}`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := page.WaitForFunction(visible, nil); err != nil {
		t.Fatalf("navigation reveals offscreen active item: %v", err)
	}
	result, err := page.Evaluate(`async()=>{const sidebar=document.querySelector('.sidebar-scroll'),before=sidebar.scrollTop;await htmx.ajax('GET','/__e2e/sidebar/30',{target:'#main-content',swap:'innerHTML'});return document.querySelector('.sidebar-scroll')===sidebar&&sidebar.scrollTop===before}`)
	if err != nil {
		t.Fatal(err)
	}
	if result != true {
		t.Fatal("visible active item changed sidebar scroll")
	}
}
