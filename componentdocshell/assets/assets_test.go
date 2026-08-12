package assets

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func servedAsset(t *testing.T, path string) string {
	t.Helper()
	recorder := httptest.NewRecorder()
	Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET %s status = %d, want 200", path, recorder.Code)
	}
	return recorder.Body.String()
}

func TestHandlerServesEmbeddedAssetsAtStablePaths(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		path        string
		contentType string
		contains    string
	}{
		{"/componentdocshell/assets/shell.css", "text/css", ".component-doc-shell"},
		{"/componentdocshell/assets/shell.js", "text/javascript", "componentDocShell"},
		{"/componentdocshell/assets/araihu.css", "text/css", `[data-theme="araihu"]`},
		{"/componentdocshell/assets/goshtoso-logo.svg", "image/svg+xml", "<svg"},
		{"/componentdocshell/assets/goshtoso-mark.svg", "image/svg+xml", "<svg"},
		{"/componentdocshell/assets/goshtoso-mark-reverse.svg", "image/svg+xml", "<svg"},
		{"/componentdocshell/assets/goshtoso-favicon.svg", "image/svg+xml", "<svg"},
	} {
		recorder := httptest.NewRecorder()
		Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))
		if recorder.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want 200", test.path, recorder.Code)
		}
		if got := recorder.Header().Get("Content-Type"); !strings.HasPrefix(got, test.contentType) {
			t.Errorf("GET %s content-type = %q, want prefix %q", test.path, got, test.contentType)
		}
		if !strings.Contains(recorder.Body.String(), test.contains) {
			t.Errorf("GET %s body missing %q", test.path, test.contains)
		}
		if got := recorder.Header().Get("X-Content-Type-Options"); got != "nosniff" {
			t.Errorf("GET %s X-Content-Type-Options = %q", test.path, got)
		}
	}
}

func TestAraiHuThemeIncludesAdaptiveLogoContract(t *testing.T) {
	t.Parallel()
	body := servedAsset(t, "/componentdocshell/assets/araihu.css")
	for _, want := range []string{"--araihu-logo-surface", "--araihu-logo-ink", "--araihu-logo-signal", `.dark [data-theme="araihu"]`} {
		if !strings.Contains(body, want) {
			t.Errorf("Arai Hû theme missing V11 contract %q", want)
		}
	}
}

func TestReleasedGoshtosoFallbackHashes(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name     string
		contents []byte
		want     string
	}{
		{"logo", goshtosoLogo, "5801b31fc6b1f54cde98b1a3f3f5e57553f6e67aa3fa0318879e5e2603cd540e"},
		{"mark", goshtosoMark, "150741f362c418b541a1d05e684b26dcd46ebfe50a11d530a81c244de7231c17"},
		{"reverse mark", goshtosoMarkReverse, "1877530c7ea23f9c597caf064e2596de5d83b78ff8795bc73a80e51d2770471e"},
		{"favicon", goshtosoFavicon, "56e8b185f2572ad4c7ea6fa8e715aa7dacd7381422b431ce07c35965ab05b3b7"},
	} {
		sum := sha256.Sum256(test.contents)
		if got := hex.EncodeToString(sum[:]); got != test.want {
			t.Errorf("%s SHA-256 = %s, want released %s", test.name, got, test.want)
		}
	}
}

func TestShellStylesDoNotOwnComponentPageComposition(t *testing.T) {
	t.Parallel()
	body := servedAsset(t, "/componentdocshell/assets/shell.css")
	for _, forbidden := range []string{
		`.component-page__example-body`,
		`.component-page__preview`,
	} {
		if strings.Contains(body, forbidden) {
			t.Errorf("shell stylesheet must not own component-page composition selector %q", forbidden)
		}
	}
}

func TestShellStylesKeepCodeCopyTargetReachable(t *testing.T) {
	t.Parallel()
	body := servedAsset(t, "/componentdocshell/assets/shell.css")
	for _, want := range []string{
		`.component-doc-shell [data-code-block-header] button {`,
		`min-width: 2.75rem;`,
		`min-height: 2.75rem;`,
		`.component-doc-shell [data-code-block-header] button:focus-visible {`,
		`outline: 2px solid var(--color-primary);`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("shell stylesheet missing code-copy target contract %q", want)
		}
	}
}

func TestShellStylesContainDocumentScrolling(t *testing.T) {
	t.Parallel()
	body := servedAsset(t, "/componentdocshell/assets/shell.css")
	for _, want := range []string{
		`.component-doc-shell-root {`,
		`overflow: hidden`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("shell stylesheet missing root scroll containment %q", want)
		}
	}
}

func TestShellStylesUseBoundedAnchorScrollPadding(t *testing.T) {
	t.Parallel()
	body := servedAsset(t, "/componentdocshell/assets/shell.css")
	for _, want := range []string{
		`.component-doc-shell__main-scroll {`,
		`scroll-padding-block: 2rem`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("shell stylesheet missing bounded anchor scroll padding %q", want)
		}
	}
	if strings.Contains(body, `.component-doc-shell__main::after {`) {
		t.Error("shell stylesheet must not add a viewport-sized pseudo-element after page content")
	}
}

func TestShellStylesDefineFamilyNavigationBreakpoints(t *testing.T) {
	t.Parallel()
	body := servedAsset(t, "/componentdocshell/assets/shell.css")
	for _, want := range []string{
		`--component-doc-shell-header-height: 4rem`,
		`height: calc(100vh - var(--component-doc-shell-header-height))`,
		`inset: var(--component-doc-shell-header-height) auto 0 0`,
		`inset: var(--component-doc-shell-header-height) 0 0`,
		`top: var(--component-doc-shell-header-height)`,
		`@media (min-width: 720px) and (max-width: 1439px)`,
		`--component-doc-shell-header-height: 6.75rem`,
		`row-gap: 0;`,
		`background: transparent;`,
		`@media (min-width: 1440px)`,
		`.component-doc-shell__family-menu`,
		`.component-doc-shell__family-links`,
		`.component-doc-shell__mobile-utilities`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("shell stylesheet missing family layout contract %q", want)
		}
	}
}

func TestShellStylesKeepFamilyNavigationControlsReachable(t *testing.T) {
	t.Parallel()
	body := servedAsset(t, "/componentdocshell/assets/shell.css")
	for _, want := range []string{
		`.component-doc-shell__sidebar-content {
  display: flex;
  height: 100%;
  min-height: 0;
  flex-direction: column;
  overflow: hidden;
}`,
		`.component-doc-shell__sidebar-content > nav {
  height: auto;
  min-height: 0;
  flex: 1 1 auto;
  overflow: hidden;
}`,
		`.component-doc-shell__sidebar-nav .sidebar-scroll {`,
		`.component-doc-shell__default-search,`,
		`.component-doc-shell[data-family-navigation="true"] .component-doc-shell__mobile-utilities button,
.component-doc-shell[data-family-navigation="true"] .component-doc-shell__brand {
  min-width: 2.75rem;
  min-height: 2.75rem;
}`,
		`.component-doc-shell__family-links .component-doc-shell__family-link:focus-visible {
    outline-offset: -2px;
  }`,
		`.component-doc-shell[data-family-navigation="true"] .component-doc-shell__sidebar-content {
    overflow-x: hidden;
    overflow-y: auto;
    overscroll-behavior: contain;
  }`,
		`.component-doc-shell[data-family-navigation="true"] .component-doc-shell__sidebar {
    width: min(320px, calc(100vw - 3rem));
    max-width: none;
    border-right: 1px solid var(--color-outline);
    background: var(--color-surface);
    box-shadow: 1rem 0 2rem rgb(0 0 0 / 0.2);
  }`,
		`.dark .component-doc-shell[data-family-navigation="true"] .component-doc-shell__sidebar {
    border-right-color: var(--color-outline-dark);
    background: var(--color-surface-dark);
  }`,
		`.component-doc-shell[data-family-navigation="true"] .component-doc-shell__scope {
    padding-inline: 1rem;
    border-right: 0;
    background: var(--color-surface-alt);
  }`,
		`.component-doc-shell[data-family-navigation="true"] .component-doc-shell__sidebar-nav {
    border-right: 0;
    background: transparent;
  }`,
		`.component-doc-shell__mobile-utilities div:has(> [id$="-mobile-listbox"]) {
    top: auto;
    bottom: calc(100% + 0.25rem);
    margin-top: 0;
    max-width: 100%;
  }`,
		`.component-doc-shell__mobile-utilities [id$="-mobile-listbox"] {
    max-height: min(13rem, calc(100vh - var(--component-doc-shell-header-height) - 2rem));
    overflow-x: hidden;
    overflow-y: auto;
  }`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("shell stylesheet missing family reachability contract %q", want)
		}
	}
}

func TestShellStylesPreserveResponsiveNavigationAffordances(t *testing.T) {
	t.Parallel()
	body := servedAsset(t, "/componentdocshell/assets/shell.css")
	for _, want := range []string{
		`.component-doc-shell__scope {
  padding: 1rem 2rem;
  border-right: 1px solid var(--color-outline);
  background: var(--color-surface);
}`,
		`.dark .component-doc-shell__scope {
  border-right-color: var(--color-outline-dark);
  background: var(--color-surface-dark);
}`,
		`.component-doc-shell__scope-details {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;`,
		`.component-doc-shell__scope-version-badge {`,
		`a.component-doc-shell__scope-module {
  color: var(--color-primary);
}`,
		`.component-doc-shell__scope-external-icon {
  display: inline-flex;
  flex: 0 0 auto;
  opacity: 0;`,
		`a.component-doc-shell__scope-module:hover .component-doc-shell__scope-external-icon,
a.component-doc-shell__scope-module:focus-visible .component-doc-shell__scope-external-icon {
  opacity: 1;`,
		`.component-doc-shell__scope-version {
  display: inline-flex;
  min-width: 2.75rem;
  min-height: 2.75rem;`,
		`.component-doc-shell__default-search input[type="search"],`,
		`.component-doc-shell__sidebar-nav .docs-sidebar-search input[type="search"] {`,
		`.component-doc-shell__sidebar-nav .sidebar-scroll a {
  min-height: 2.75rem;
}`,
		`.component-doc-shell [id$="-listbox"] [role="option"] {
  min-height: 3rem;
  align-items: center;
}`,
		`grid-template-columns: max-content minmax(0, 1fr) max-content`,
		`width: 48px;`,
		`width: 44px;`,
		`flex: 0 0 44px;`,
		`min-width: 44px;`,
		`max-height: 2rem;`,
		`object-fit: contain;`,
		`gap: 0;`,
		`padding: 0;`,
		`padding-inline: 0;`,
		`overflow-x: auto;`,
		`scrollbar-gutter: stable;`,
		`flex: 0 0 auto;`,
		`place-items: center;`,
		`width: 1.25rem;`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("shell stylesheet missing responsive navigation affordance contract %q", want)
		}
	}
}

func TestShellStylesDoNotRestyleArbitrarySidebarSlotRoots(t *testing.T) {
	t.Parallel()
	body := servedAsset(t, "/componentdocshell/assets/shell.css")
	if strings.Contains(body, `.component-doc-shell__sidebar-nav > div:not(.sidebar-scroll)`) {
		t.Fatal("shell stylesheet still applies structural padding to arbitrary sidebar slot roots")
	}
}

func TestShellStylesKeepTextZoomAndManagedLogoBounds(t *testing.T) {
	t.Parallel()
	body := servedAsset(t, "/componentdocshell/assets/shell.css")
	for _, want := range []string{
		`.component-doc-shell[data-family-navigation="true"] .component-doc-shell__managed-logo {`,
		`width: 48px;`,
		`height: auto;`,
		`max-height: 2rem;`,
		`object-fit: contain;`,
		`.component-doc-shell[data-family-navigation="true"] .component-doc-shell__family-select > div > button {`,
		`padding-inline: 0;`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("shell stylesheet missing text-zoom or managed-logo bound contract %q", want)
		}
	}
	if strings.Contains(body, `font-size: min(1rem, 4vw)`) {
		t.Error("small family summary must inherit root font scaling rather than cap text zoom")
	}
}

func TestShellStylesKeepCompactBrandArtworkInsideFixedMark(t *testing.T) {
	t.Parallel()
	body := servedAsset(t, "/componentdocshell/assets/shell.css")
	for _, want := range []string{
		`.component-doc-shell__brand-compact-mark .component-doc-shell__brand-logo`,
		`.component-doc-shell__brand-compact-mark .component-doc-shell__brand-logo img`,
		`width: 100%;`,
		`height: 100%;`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("shell stylesheet missing compact mark containment rule %q", want)
		}
	}
}

func TestShellStylesClampSmallFamilyMenuAtTextZoom(t *testing.T) {
	t.Parallel()
	body := servedAsset(t, "/componentdocshell/assets/shell.css")
	for _, want := range []string{
		`.component-doc-shell__family-select > div > button {
  position: relative;
  width: 100%;
  min-width: 0;
  min-height: 2.75rem;
  justify-content: center;
  overflow: hidden;
  padding-inline: 2.75rem;
  font-size: 1rem;
}`,
		`.component-doc-shell__family-select > div > button > svg {
  position: absolute;
  right: 1rem;
}`,
		`.component-doc-shell__family-select > div > button[aria-expanded="true"] {
  border-color: var(--color-primary);
  background: var(--color-surface-alt);
}`,
		`.dark .component-doc-shell__family-select > div > button[aria-expanded="true"] {
  border-color: var(--color-primary-dark);
  background: var(--color-surface-dark-alt);
}`,
		`.component-doc-shell__family-select div:has(> #componentdocshell-family-listbox) {
  right: 0;
  left: 0;
  z-index: 60;
  width: auto;
  min-width: 0;
  max-width: none;
  box-sizing: border-box;
}`,
		`.component-doc-shell[data-family-navigation="true"] .component-doc-shell__family-menu-links {
    right: 0;
    left: auto;
    width: max-content;
    min-width: min(12rem, calc(100vw - 2rem));
    max-width: calc(100vw - 2rem);
    box-sizing: border-box;
  }`,
		`#componentdocshell-family-listbox [role="option"] {
  position: relative;
  justify-content: center;
  padding-inline: 2.5rem;
  text-align: center;
}`,
		`#componentdocshell-family-listbox [role="option"] > svg {
  position: absolute;
  right: 1rem;
}`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("shell stylesheet missing small family menu text-zoom clamp %q", want)
		}
	}
	if strings.Contains(body, `.component-doc-shell[data-family-navigation="true"] .component-doc-shell__family-select > div > button {
    gap: 0;
    min-height: 44px;
    border-color: transparent;
    background: transparent;`) {
		t.Error("small family trigger must preserve Goshtoso Select chrome")
	}
}

func TestShellStylesKeepFamilyLinksReachableAtMediumAndWideWidths(t *testing.T) {
	t.Parallel()
	body := servedAsset(t, "/componentdocshell/assets/shell.css")
	for _, want := range []string{
		`@media (min-width: 720px) and (max-width: 1439px)`,
		`background: transparent;`,
		`@media (min-width: 1440px)`,
		`justify-content: safe center;`,
		`overflow-x: auto;`,
		`scrollbar-width: thin;`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("shell stylesheet missing family overflow contract %q", want)
		}
	}
}

func TestShellStylesProvideCompactBrandFallback(t *testing.T) {
	t.Parallel()
	body := servedAsset(t, "/componentdocshell/assets/shell.css")
	for _, want := range []string{
		`.component-doc-shell__brand-logo-source {
  display: contents;
}`,
		`.component-doc-shell__brand-compact-mark {
  display: none;`,
		`.component-doc-shell[data-family-navigation="true"] .component-doc-shell__managed-logo {`,
		`.component-doc-shell[data-family-navigation="true"] .component-doc-shell__brand-mark {`,
		`.component-doc-shell[data-family-navigation="true"] .component-doc-shell__brand-logo-source {`,
		`.component-doc-shell[data-family-navigation="true"] .component-doc-shell__brand-compact-mark {`,
		`.component-doc-shell__brand-compact-mark > * {`,
		`--component-doc-shell-header-height: 64px`,
		`width: 44px;`,
		`opacity: 0;`,
		`letter-spacing: -0.02em;`,
		`width: 16px;`,
		`flex: 0 0 16px;`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("shell stylesheet missing compact-brand contract %q", want)
		}
	}
	for _, unscoped := range []string{
		"\n  .component-doc-shell__managed-logo {",
		"\n  .component-doc-shell__brand-mark {",
		"\n  .component-doc-shell__brand-logo-source {",
		"\n  .component-doc-shell__brand-compact-mark {",
	} {
		if strings.Contains(body, unscoped) {
			t.Errorf("small compact-brand rule escaped family scope %q", unscoped)
		}
	}
}

func TestShellRuntimeExposesThemeSetter(t *testing.T) {
	t.Parallel()
	body := servedAsset(t, "/componentdocshell/assets/shell.js")
	for _, want := range []string{`setTheme: function (value)`, `root.dataset.themeSource === "preference"`, `document.documentElement.dataset.themeSource = "preference"`, `this.theme = value`, `if (!self.persistTheme) return;`} {
		if !strings.Contains(body, want) {
			t.Errorf("shell runtime missing theme setter contract %q", want)
		}
	}
}

func TestShellRuntimeTracksResponsiveSidebarPersistence(t *testing.T) {
	t.Parallel()
	body := servedAsset(t, "/componentdocshell/assets/shell.js")
	for _, want := range []string{
		`var sidebarMedia = window.matchMedia("(min-width: 720px)")`,
		`sidebarPersistent: sidebarMedia.matches`,
		`self.sidebarPersistent = event.matches`,
		`if (event.matches) self.sidebarOpen = false;`,
		`sidebarMedia.addEventListener("change", syncSidebarPersistence)`,
		`sidebarMedia.addListener(syncSidebarPersistence)`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("shell runtime missing responsive sidebar contract %q", want)
		}
	}
	if strings.Contains(body, `sidebarPersistent: window.innerWidth`) {
		t.Error("shell runtime must not derive sidebar persistence from a one-shot viewport width")
	}
	persistent := strings.Index(body, `self.sidebarPersistent = event.matches`)
	closeDrawer := strings.Index(body, `if (event.matches) self.sidebarOpen = false;`)
	if persistent == -1 || closeDrawer == -1 || persistent >= closeDrawer {
		t.Errorf("shell runtime persistent transition order = state:%d close:%d, want state < close", persistent, closeDrawer)
	}
}

func TestShellRuntimeContainsMobileDrawerFocus(t *testing.T) {
	t.Parallel()
	body := servedAsset(t, "/componentdocshell/assets/shell.js")
	for _, want := range []string{
		`function drawerFocusables(sidebar)`,
		`self.sidebarOpen && !self.sidebarPersistent`,
		`if (event.key !== "Tab") return;`,
		`document.addEventListener("keydown", containDrawerTab, true)`,
		`document.addEventListener("focusin", containDrawerFocus, true)`,
		`document.removeEventListener("keydown", containDrawerTab, true)`,
		`document.removeEventListener("focusin", containDrawerFocus, true)`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("shell runtime missing mobile drawer containment contract %q", want)
		}
	}
}

func TestShellRuntimePreparesMainHeadingForProgrammaticFocus(t *testing.T) {
	t.Parallel()
	body := servedAsset(t, "/componentdocshell/assets/shell.js")
	for _, want := range []string{
		`function mainFocusTarget()`,
		`if (!target.hasAttribute("tabindex")) target.setAttribute("tabindex", "-1")`,
		`document.addEventListener("DOMContentLoaded", function ()`,
		`mainFocusTarget();`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("shell runtime missing main focus target contract %q", want)
		}
	}
}

func TestShellRuntimeUsesTOCRolesAndLegacyLinkHook(t *testing.T) {
	t.Parallel()
	body := servedAsset(t, "/componentdocshell/assets/shell.js")
	for _, want := range []string{
		`[data-componentdocshell-toc]`,
		`[data-componentdocshell-toc-list]`,
		`link.setAttribute("data-toc-link", heading.id)`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("shell runtime missing TOC contract %q", want)
		}
	}
}

func TestShellRuntimeAlignsHashInsideMainScroller(t *testing.T) {
	t.Parallel()
	body := servedAsset(t, "/componentdocshell/assets/shell.js")
	for _, want := range []string{
		`function scrollTarget(target, behavior)`,
		`document.documentElement.scrollTop = 0`,
		`document.body.scrollTop = 0`,
		`scroller.scrollTo({ top: nextTop, behavior: behavior || "auto" })`,
		`function tocScrollBehavior()`,
		`window.matchMedia("(prefers-reduced-motion: reduce)")`,
		`scrollTarget(heading, tocScrollBehavior())`,
		`requestAnimationFrame(function () { scrollTarget(active, "auto"); })`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("shell runtime missing hash alignment %q", want)
		}
	}
}

func TestShellRuntimeKeepsNavigationLifecycleIndependentFromFamilySelect(t *testing.T) {
	t.Parallel()
	body := servedAsset(t, "/componentdocshell/assets/shell.js")
	for _, want := range []string{
		`function syncFamilySelect()`,
		`function closeFamilySelect()`,
		`componentdocshell-family-select-control`,
		`selectState.syncFromInput(href)`,
		`window.addEventListener("componentdocshell:close-family-select", closeFamilySelect)`,
		`window.dispatchEvent(new CustomEvent("componentdocshell:navigated"))`,
		`focusMain();`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("shell runtime missing navigation lifecycle contract %q", want)
		}
	}
	for _, obsolete := range []string{`function closeFamilyMenu()`, `[data-componentdocshell-family-menu]`, `menu.open = false`} {
		if strings.Contains(body, obsolete) {
			t.Errorf("shell runtime still owns obsolete custom family disclosure behavior %q", obsolete)
		}
	}

	afterSwap := strings.Index(body, `document.addEventListener("htmx:afterSwap"`)
	if afterSwap == -1 {
		t.Fatal("shell runtime missing htmx:afterSwap handler")
	}
	mainBranch := body[afterSwap:]
	guard := strings.Index(mainBranch, `event.detail.target.id !== "main-content"`)
	syncSelect := strings.Index(mainBranch, `syncFamilySelect();`)
	dispatch := strings.Index(mainBranch, `window.dispatchEvent(new CustomEvent("componentdocshell:navigated"))`)
	focus := strings.Index(mainBranch, `focusMain();`)
	if guard == -1 || syncSelect == -1 || dispatch == -1 || focus == -1 {
		t.Fatal("shell runtime main-target branch missing navigation lifecycle ordering markers")
	}
	if !(guard < syncSelect && syncSelect < dispatch && dispatch < focus) {
		t.Errorf("shell runtime navigation lifecycle order = guard:%d sync:%d dispatch:%d focus:%d, want guard < sync < dispatch < focus", guard, syncSelect, dispatch, focus)
	}
}

func TestShellRuntimeRestoresFamilyLifecycleFromHistory(t *testing.T) {
	t.Parallel()
	body := servedAsset(t, "/componentdocshell/assets/shell.js")
	historyRestore := strings.Index(body, `document.addEventListener("htmx:historyRestore"`)
	if historyRestore == -1 {
		t.Fatal("shell runtime missing htmx:historyRestore handler")
	}
	historyBranch := body[historyRestore:]
	handlerEnd := strings.Index(historyBranch, "\n  });")
	if handlerEnd == -1 {
		t.Fatal("shell runtime history restore handler has no bounded end")
	}
	historyBranch = historyBranch[:handlerEnd]

	mainGuard := strings.Index(historyBranch, `if (!mainContent()) return;`)
	syncSelect := strings.Index(historyBranch, `syncFamilySelect();`)
	dispatch := strings.Index(historyBranch, `window.dispatchEvent(new CustomEvent("componentdocshell:navigated"))`)
	build := strings.Index(historyBranch, `buildTOC();`)
	focus := strings.Index(historyBranch, `focusMain();`)
	if mainGuard == -1 || syncSelect == -1 || dispatch == -1 || build == -1 || focus == -1 {
		t.Fatalf("shell runtime history restore lifecycle incomplete:\n%s", historyBranch)
	}
	if !(mainGuard < syncSelect && syncSelect < dispatch && dispatch < build && build < focus) {
		t.Errorf("shell runtime history restore order = guard:%d sync:%d dispatch:%d build:%d focus:%d, want guard < sync < dispatch < build < focus", mainGuard, syncSelect, dispatch, build, focus)
	}
	if strings.Contains(historyBranch, `scrollTo(`) {
		t.Error("shell runtime history restore must preserve restored scroll state")
	}
}

func TestHandlerRejectsUnknownAndTraversalPaths(t *testing.T) {
	t.Parallel()
	for _, path := range []string{
		"/componentdocshell/assets/missing.css",
		"/componentdocshell/assets/../go.mod",
		"/shell.css",
	} {
		recorder := httptest.NewRecorder()
		Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		if recorder.Code != http.StatusNotFound {
			t.Errorf("GET %s status = %d, want 404", path, recorder.Code)
		}
	}
}

func TestAssetURLsRespectPrefix(t *testing.T) {
	t.Parallel()
	if got := StylesheetURL("/custom/"); !strings.HasPrefix(got, "/custom/shell.css?v=") {
		t.Errorf("StylesheetURL() = %q", got)
	}
	if got := ScriptURL("/custom/"); !strings.HasPrefix(got, "/custom/shell.js?v=") {
		t.Errorf("ScriptURL() = %q", got)
	}
	if got := AraiHuThemeURL("/custom/"); !strings.HasPrefix(got, "/custom/araihu.css?v=") {
		t.Errorf("AraiHuThemeURL() = %q", got)
	}
}

func TestHandlerSupportsCustomPrefix(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	Handler("/custom/").ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/custom/shell.css", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET custom shell.css status = %d", recorder.Code)
	}
}
