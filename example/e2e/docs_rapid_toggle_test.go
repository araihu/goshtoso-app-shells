package e2e

import "testing"

// Closing before Alpine Focus's deferred activation must not leave a hidden
// drawer trapping Tab after its visible state has already closed.
func TestDocsDrawerRapidCloseReleasesTab(t *testing.T) {
	requireE2E(t)
	harness := newBrowserHarness(t)
	page := harness.newPage(t, true)
	if err := page.SetViewportSize(390, 700); err != nil {
		t.Fatal(err)
	}
	gotoFamilyPage(t, page, harness.baseURL, "/components")
	if _, err := page.Evaluate(`async () => {
  const state=Alpine.$data(document.documentElement);
  state.sidebarOpen=true;
  await Promise.resolve();await Promise.resolve();
  state.sidebarOpen=false;
  await Alpine.nextTick();
  // Exercise the exact delayed-activation race, after its 15ms timer fires.
  await new Promise(resolve=>setTimeout(resolve,50));
  document.querySelector('#main-content').focus();
  window.drawerTabBlocked=null;
  window.addEventListener('keydown',event=>{if(event.key==='Tab')window.drawerTabBlocked=event.defaultPrevented;},{once:true});
 }`); err != nil {
		t.Fatal(err)
	}
	if err := page.Keyboard().Press("Tab"); err != nil {
		t.Fatal(err)
	}
	blocked, err := page.Evaluate(`() => window.drawerTabBlocked`)
	if err != nil {
		t.Fatal(err)
	}
	if blocked != false {
		t.Fatalf("closed drawer trapped Tab: %v", blocked)
	}
}
