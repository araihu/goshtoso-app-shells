package e2e

import "testing"

// A nested Alpine overlay pauses the drawer trap, then returns ownership to it.
func TestDocsDrawerNestedNativeTrap(t *testing.T) {
	requireE2E(t)
	harness := newBrowserHarness(t)
	page := harness.newPage(t, true)
	if err := page.SetViewportSize(390, 700); err != nil {
		t.Fatal(err)
	}
	gotoFamilyPage(t, page, harness.baseURL, "/components")
	if err := page.Locator(".component-doc-shell__menu-button").Click(); err != nil {
		t.Fatal(err)
	}
	waitForDrawer(t, page, true)
	if _, err := page.Evaluate(`() => {const overlay=document.createElement('div');overlay.id='nested-trap';overlay.setAttribute('x-data','{open: true}');overlay.setAttribute('x-trap.noreturn','open');overlay.style.cssText='position:fixed;inset:80px;z-index:10000;background:white';overlay.innerHTML='<button id="nested-first">First</button><button id="nested-last">Last</button>';document.body.appendChild(overlay);}`); err != nil {
		t.Fatal(err)
	}
	if _, err := page.WaitForFunction(`() => document.activeElement?.id==='nested-first'`, nil); err != nil {
		t.Fatal(err)
	}
	if err := page.Locator("#nested-last").Focus(); err != nil {
		t.Fatal(err)
	}
	if err := page.Keyboard().Press("Tab"); err != nil {
		t.Fatal(err)
	}
	if !activeMatches(t, page, "#nested-first") {
		t.Fatal("nested trap did not wrap")
	}
	if _, err := page.Evaluate(`async () => {Alpine.$data(document.querySelector('#nested-trap')).open=false;await Alpine.nextTick();document.querySelector('#nested-trap').remove();document.querySelector('#main-content').focus();}`); err != nil {
		t.Fatal(err)
	}
	if _, err := page.WaitForFunction(`() => document.querySelector('#componentdocshell-sidebar').contains(document.activeElement)`, nil); err != nil {
		t.Fatal(err)
	}
}
