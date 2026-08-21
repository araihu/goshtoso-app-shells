package e2e

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-app-shells/componentdocshell"
	"github.com/araihu/goshtoso-app-shells/example/internal/pages"
	"github.com/araihu/goshtoso-app-shells/example/internal/server"
	"github.com/mxschmitt/playwright-go"
)

var familyWidths = []int{390, 545, 719, 720, 735, 820, 841, 1023, 1024, 1199, 1200, 1280, 1439, 1440}
var familyThemes = []string{"araihu", "goshtoso", "minimal"}

func requireE2E(t *testing.T) {
	t.Helper()
	if os.Getenv("COMPONENTDOCSHELL_E2E") != "1" {
		t.Skip("set COMPONENTDOCSHELL_E2E=1 to run browser acceptance")
	}
}

type browserHarness struct {
	baseURL string
	browser playwright.Browser
}

func newBrowserHarness(t *testing.T) *browserHarness {
	t.Helper()

	productHandler := server.New()
	testHandler := http.NewServeMux()
	testHandler.HandleFunc("GET /__e2e/persistent", func(writer http.ResponseWriter, request *http.Request) {
		page, _ := pages.FamilyOverview("components")
		config := pages.ShellConfig("components")
		config.Appearance.DisableThemeSelector = false
		config.Appearance.PersistPreferences = true
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := componentdocshell.Layout(config, page).Render(request.Context(), writer); err != nil {
			http.Error(writer, "render persistent test page", http.StatusInternalServerError)
		}
	})
	testHandler.HandleFunc("GET /__e2e/legacy-managed", func(writer http.ResponseWriter, request *http.Request) {
		config := pages.ShellConfig("components")
		config.Navigation.Families = nil
		config.Brand.Logo = nil
		config.Brand.ManagedLogo = &componentdocshell.ManagedBrandAsset{URL: "/fixtures/brand/logo.svg", Alt: "Component docs shell", Width: 120, Height: 32}
		page := pages.Overview()
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := componentdocshell.Layout(config, page).Render(request.Context(), writer); err != nil {
			http.Error(writer, "render legacy managed-logo test page", http.StatusInternalServerError)
		}
	})
	testHandler.HandleFunc("GET /__e2e/legacy-custom", func(writer http.ResponseWriter, request *http.Request) {
		config := pages.ShellConfig("components")
		config.Navigation.Families = nil
		config.Interactions.PresentationChannel = nil
		config.Brand.ManagedLogo = nil
		config.Brand.Logo = templ.Raw(`<svg class="legacy-custom-logo" width="120" height="32" viewBox="0 0 120 32" role="img" aria-label="Legacy custom logo"><rect width="120" height="32" fill="#1e293b"/></svg>`)
		page := pages.Overview()
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := componentdocshell.Layout(config, page).Render(request.Context(), writer); err != nil {
			http.Error(writer, "render legacy custom-logo test page", http.StatusInternalServerError)
		}
	})
	testHandler.Handle("/", productHandler)
	testServer := httptest.NewServer(testHandler)
	t.Cleanup(testServer.Close)

	pw, err := playwright.Run()
	if err != nil {
		t.Fatalf("start Playwright: %v", err)
	}
	t.Cleanup(func() {
		if err := pw.Stop(); err != nil {
			t.Errorf("stop Playwright: %v", err)
		}
	})

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		t.Fatalf("launch Chromium: %v", err)
	}
	t.Cleanup(func() {
		if err := browser.Close(); err != nil {
			t.Errorf("close Chromium: %v", err)
		}
	})

	return &browserHarness{baseURL: testServer.URL, browser: browser}
}

func (h *browserHarness) newPage(t *testing.T, javascript bool) playwright.Page {
	t.Helper()
	return h.newPageWithOptions(t, playwright.BrowserNewContextOptions{
		JavaScriptEnabled: playwright.Bool(javascript),
	})
}

func (h *browserHarness) newPageWithOptions(t *testing.T, options playwright.BrowserNewContextOptions) playwright.Page {
	t.Helper()

	context, err := h.browser.NewContext(options)
	if err != nil {
		t.Fatalf("create browser context: %v", err)
	}
	t.Cleanup(func() {
		if err := context.Close(); err != nil {
			t.Errorf("close browser context: %v", err)
		}
	})

	page, err := context.NewPage()
	if err != nil {
		t.Fatalf("create browser page: %v", err)
	}
	page.SetDefaultTimeout(5000)
	page.SetDefaultNavigationTimeout(5000)
	t.Cleanup(func() {
		if err := page.Close(); err != nil {
			t.Errorf("close browser page: %v", err)
		}
	})
	return page
}

func gotoFamilyPage(t *testing.T, page playwright.Page, baseURL, path string) {
	t.Helper()
	response, err := page.Goto(baseURL+path, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	if response == nil || response.Status() != 200 {
		t.Fatalf("GET %s status = %v, want 200", path, response)
	}
	if err := page.Locator("#main-content").WaitFor(); err != nil {
		t.Fatalf("wait for main content at %s: %v", path, err)
	}
}

func TestFamilyNavigationVisualMatrix(t *testing.T) {
	requireE2E(t)
	harness := newBrowserHarness(t)

	for _, width := range familyWidths {
		for _, theme := range familyThemes {
			for _, dark := range []bool{false, true} {
				name := fmt.Sprintf("%d/%s/%s", width, theme, map[bool]string{false: "light", true: "dark"}[dark])
				t.Run(name, func(t *testing.T) {
					page := harness.newPage(t, true)
					if err := page.SetViewportSize(width, 900); err != nil {
						t.Fatal(err)
					}
					seed := fmt.Sprintf(`localStorage.setItem("theme", %q); localStorage.setItem("darkMode", %q);`, theme, fmt.Sprintf("%t", dark))
					if err := page.AddInitScript(playwright.Script{Content: playwright.String(seed)}); err != nil {
						t.Fatal(err)
					}
					failures := watchBrowserFailures(page)
					failures.setPhase("initial-load")
					gotoFamilyPage(t, page, harness.baseURL, "/components")
					setMatrixAppearance(t, page, theme, dark)
					prepareMatrixSurfaces(t, page, width)
					focusMetrics := matrixFocusMetrics(t, page, width)
					metrics := collectMatrixMetrics(t, page, width, theme, dark, focusMetrics)
					touchTargets := collectMatrixTouchTargetMetrics(t, page, width)
					metrics["touchTargetAudit"] = touchTargets
					metrics["touchTargetsAtLeast44"] = touchTargets["allVisibleIntendedTargetsAtLeast44"]
					metrics["touchTargetScopeComplete"] = touchTargets["scopeComplete"]
					page.WaitForTimeout(50)
					metrics["browserErrors"] = failures.snapshot()
					assertMatrixMetrics(t, metrics)
				})
			}
		}
	}
}

func TestScopeModuleExternalLinkAffordance(t *testing.T) {
	requireE2E(t)
	harness := newBrowserHarness(t)
	page := harness.newPage(t, true)
	if err := page.SetViewportSize(390, 720); err != nil {
		t.Fatal(err)
	}
	gotoFamilyPage(t, page, harness.baseURL, "/components/button")
	if err := page.Locator(".component-doc-shell__menu-button").Click(); err != nil {
		t.Fatal(err)
	}
	waitForDrawer(t, page, true)

	module := page.Locator("a.component-doc-shell__scope-module")
	initial, err := module.Evaluate(`element => {
		const probe = document.createElement('span');
		probe.className = 'text-primary dark:text-primary-dark';
		document.body.append(probe);
		const primary = getComputedStyle(probe).color;
		probe.remove();
		const icon = element.querySelector('.component-doc-shell__scope-external-icon');
		return {
			linkColorIsPrimary: getComputedStyle(element).color === primary,
			iconColorIsPrimary: getComputedStyle(icon.querySelector('svg')).color === primary,
			iconOpacity: getComputedStyle(icon).opacity,
			iconHref: icon.querySelector('use')?.getAttribute('href'),
			iconDecorative: icon.querySelector('svg')?.getAttribute('aria-hidden') === 'true',
		};
	}`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := page.Evaluate(`() => document.documentElement.classList.add('dark')`); err != nil {
		t.Fatal(err)
	}
	darkColors, err := module.Evaluate(`element => {
		const probe = document.createElement('span');
		probe.className = 'text-primary dark:text-primary-dark';
		document.body.append(probe);
		const primary = getComputedStyle(probe).color;
		probe.remove();
		return {
			link: getComputedStyle(element).color === primary,
			icon: getComputedStyle(element.querySelector('svg')).color === primary,
		};
	}`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := module.Hover(); err != nil {
		t.Fatal(err)
	}
	if _, err := page.WaitForFunction(`() => getComputedStyle(document.querySelector('a.component-doc-shell__scope-module .component-doc-shell__scope-external-icon')).opacity === '1'`, nil); err != nil {
		t.Fatalf("wait for scope module hover affordance: %v", err)
	}
	hoverOpacity, err := module.Locator(".component-doc-shell__scope-external-icon").Evaluate(`element => getComputedStyle(element).opacity`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := page.Locator(".component-doc-shell__menu-button").Hover(); err != nil {
		t.Fatal(err)
	}
	if err := module.Focus(); err != nil {
		t.Fatal(err)
	}
	if err := page.Keyboard().Press("Tab"); err != nil {
		t.Fatal(err)
	}
	if err := page.Keyboard().Press("Shift+Tab"); err != nil {
		t.Fatal(err)
	}
	if _, err := page.WaitForFunction(`() => getComputedStyle(document.querySelector('a.component-doc-shell__scope-module .component-doc-shell__scope-external-icon')).opacity === '1'`, nil); err != nil {
		t.Fatalf("wait for scope module focus affordance: %v", err)
	}
	focusOpacity, err := module.Locator(".component-doc-shell__scope-external-icon").Evaluate(`element => getComputedStyle(element).opacity`, nil)
	if err != nil {
		t.Fatal(err)
	}
	metrics := initial.(map[string]any)
	metrics["darkColors"] = darkColors
	metrics["hoverOpacity"] = hoverOpacity
	metrics["focusOpacity"] = focusOpacity
	darkMetrics := darkColors.(map[string]any)
	if metrics["linkColorIsPrimary"] != true || metrics["iconColorIsPrimary"] != true || darkMetrics["link"] != true || darkMetrics["icon"] != true || metrics["iconOpacity"] != "0" || metrics["hoverOpacity"] != "1" || metrics["focusOpacity"] != "1" ||
		metrics["iconHref"] != "/assets/icons/heroicons.svg#hi-16-solid-arrow-top-right-on-square" || metrics["iconDecorative"] != true {
		failWithMetrics(t, "scope module external-link affordance", metrics)
	}
}

type browserFailures struct {
	mu       sync.Mutex
	messages []string
	phase    string
	started  time.Time
}

func watchBrowserFailures(page playwright.Page) *browserFailures {
	failures := &browserFailures{phase: "unassigned", started: time.Now()}
	add := func(message string) {
		failures.mu.Lock()
		defer failures.mu.Unlock()
		failures.messages = append(failures.messages, fmt.Sprintf("phase=%s elapsed=%s %s", failures.phase, time.Since(failures.started).Round(time.Millisecond), message))
	}
	page.OnPageError(func(err error) {
		pageError := &playwright.Error{}
		if errors.As(err, &pageError) {
			add(fmt.Sprintf("page error: name=%s message=%s stack=%s", pageError.Name, pageError.Message, pageError.Stack))
			return
		}
		add("page error: " + err.Error())
	})
	page.OnConsole(func(message playwright.ConsoleMessage) {
		if message.Type() == "error" {
			add("console error: " + message.Text())
		}
	})
	return failures
}

func (f *browserFailures) setPhase(phase string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.phase = phase
}

func (f *browserFailures) snapshot() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.messages...)
}

func setMatrixAppearance(t *testing.T, page playwright.Page, theme string, dark bool) {
	t.Helper()
	if _, err := page.WaitForFunction(`() => Boolean(window.Alpine && window.Alpine.$data(document.documentElement)?.setTheme)`, nil); err != nil {
		t.Fatalf("wait for shell appearance state: %v", err)
	}
	if _, err := page.Evaluate(`({theme, dark}) => {
		const state = window.Alpine.$data(document.documentElement);
		state.setTheme(theme);
		if (Boolean(state.dark) !== dark) state.toggleDark();
	}`, map[string]any{"theme": theme, "dark": dark}); err != nil {
		t.Fatalf("set matrix appearance: %v", err)
	}
	if _, err := page.WaitForFunction(`({theme, dark}) =>
		document.documentElement.dataset.theme === theme &&
		document.documentElement.classList.contains('dark') === dark`,
		map[string]any{"theme": theme, "dark": dark}); err != nil {
		t.Fatalf("wait for matrix appearance: %v", err)
	}
}

func prepareMatrixSurfaces(t *testing.T, page playwright.Page, width int) {
	t.Helper()
	if width >= 1024 {
		return
	}
	if err := page.Locator(".component-doc-shell__menu-button").Click(); err != nil {
		t.Fatalf("open local drawer: %v", err)
	}
	if _, err := page.WaitForFunction(`() => {
		const sidebar = document.querySelector('.component-doc-shell__sidebar');
		const backdrop = document.querySelector('.component-doc-shell__backdrop');
		const search = sidebar?.querySelector('input[type="search"]');
		const sidebarRect = sidebar?.getBoundingClientRect();
		const searchRect = search?.getBoundingClientRect();
		return sidebar?.classList.contains('is-open') && backdrop && getComputedStyle(backdrop).display !== 'none' &&
			sidebarRect && sidebarRect.left >= -0.5 && sidebarRect.right > 0 &&
			searchRect && searchRect.left >= 0 && searchRect.right <= innerWidth;
		}`, nil); err != nil {
		t.Fatalf("wait for local drawer: %v", err)
	}
	if width >= 720 {
		return
	}
	if err := page.Locator("#componentdocshell-family-trigger").Click(); err != nil {
		t.Fatalf("open family select: %v", err)
	}
	if err := page.Locator("#componentdocshell-family-listbox").WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}); err != nil {
		t.Fatalf("wait for family select: %v", err)
	}
	if _, err := page.WaitForFunction(`() => {
		const popup = document.querySelector('#componentdocshell-family-listbox')?.parentElement;
		if (!popup) return false;
		const style = getComputedStyle(popup);
		return style.opacity === '1' && style.transform === 'none';
	}`, nil); err != nil {
		t.Fatalf("wait for family select transition: %v", err)
	}
}

func matrixFocusMetrics(t *testing.T, page playwright.Page, width int) map[string]any {
	t.Helper()
	if width != 1024 && width != 1439 {
		return map[string]any{"applicable": false, "trusted": true}
	}
	if _, err := page.Evaluate(`() => document.activeElement?.blur()`); err != nil {
		t.Fatal(err)
	}
	focused := false
	for range 16 {
		if err := page.Keyboard().Press("Tab"); err != nil {
			t.Fatalf("keyboard Tab through header: %v", err)
		}
		value, err := page.Evaluate(`() => document.activeElement?.matches('.component-doc-shell__family-links a') || false`)
		if err != nil {
			t.Fatal(err)
		}
		if value == true {
			focused = true
			break
		}
	}
	if !focused {
		return map[string]any{"applicable": true, "trusted": false, "reason": "keyboard Tab did not reach inline family link"}
	}
	result, err := page.Evaluate(`() => {
		const link = document.activeElement;
		const nav = link.closest('.component-doc-shell__family-links');
		const style = getComputedStyle(link);
		const linkRect = link.getBoundingClientRect();
		const navRect = nav.getBoundingClientRect();
		const width = parseFloat(style.outlineWidth) || 0;
		const offset = parseFloat(style.outlineOffset) || 0;
		const extension = Math.max(0, width + offset);
		return {
			applicable: true,
			activeText: link.textContent.trim(),
			outlineWidth: width,
			outlineOffset: offset,
			outlineStyle: style.outlineStyle,
			outlineColor: style.outlineColor,
			trusted: width >= 2 && style.outlineStyle !== 'none' &&
				linkRect.left - extension >= navRect.left - 0.5 &&
				linkRect.right + extension <= navRect.right + 0.5 &&
				linkRect.top - extension >= navRect.top - 0.5 &&
				linkRect.bottom + extension <= navRect.bottom + 0.5,
		};
	}`)
	if err != nil {
		t.Fatalf("measure keyboard focus treatment: %v", err)
	}
	return result.(map[string]any)
}

func collectMatrixTouchTargetMetrics(t *testing.T, page playwright.Page, width int) map[string]any {
	t.Helper()
	base, err := page.Evaluate(`() => {
		const header = document.querySelector('.component-doc-shell__header');
		const sidebar = document.querySelector('.component-doc-shell__sidebar');
		const visible = element => {
			if (!element) return false;
			const style = getComputedStyle(element);
			if (style.display === 'none' || style.visibility === 'hidden') return false;
			const rect = element.getBoundingClientRect();
			return rect.width > 0 && rect.height > 0 && rect.right > 0 && rect.bottom > 0 && rect.left < innerWidth && rect.top < innerHeight;
		};
		const label = element => (element.getAttribute('aria-label') || element.textContent || element.getAttribute('placeholder') || '').trim();
		const candidates = [...new Set([
			...header.querySelectorAll('a[href], button, [role="option"]'),
			...sidebar.querySelectorAll('a[href], button, input:not([type="hidden"]), summary'),
		])].filter(element => visible(element) && !element.disabled && element.getAttribute('aria-disabled') !== 'true');
		const localLinks = Array.from(sidebar.querySelectorAll('nav[aria-label="sidebar navigation"] a[href]')).filter(visible);
		const searchControls = Array.from(sidebar.querySelectorAll('input[type="search"], [role="search"] input, [role="search"] button')).filter(visible);
		const familyTargets = Array.from(header.querySelectorAll('.component-doc-shell__family-links a[href], #componentdocshell-family-listbox [role="option"]')).filter(visible);
		const metrics = candidates.map(element => {
			const rect = element.getBoundingClientRect();
			const categories = [];
			if (localLinks.includes(element)) categories.push('local-sidebar-link');
			if (searchControls.includes(element)) categories.push('search-control');
			if (familyTargets.includes(element)) categories.push('family-navigation');
			if (header.contains(element) && categories.length === 0) categories.push('header-control');
			if (sidebar.contains(element) && categories.length === 0) categories.push('sidebar-control');
			return {
				origin: 'base-shell',
				categories,
				tag: element.tagName.toLowerCase(),
				role: element.getAttribute('role') || '',
				id: element.id,
				href: element.getAttribute('href') || '',
				name: element.getAttribute('name') || '',
				label: label(element),
				width: rect.width,
				height: rect.height,
			};
		});
		return {
			targetMetrics: metrics,
			visibleIntendedTargetCount: candidates.length,
			visibleFamilyTargetCount: familyTargets.length,
			visibleLocalSidebarLinkCount: localLinks.length,
			visibleSearchControlCount: searchControls.length,
			auditedFamilyTargetCount: metrics.filter(item => item.categories.includes('family-navigation')).length,
			auditedLocalSidebarLinkCount: metrics.filter(item => item.categories.includes('local-sidebar-link')).length,
			auditedSearchControlCount: metrics.filter(item => item.categories.includes('search-control')).length,
		};
	}`)
	if err != nil {
		t.Fatalf("collect visible shell touch targets: %v", err)
	}

	result, err := page.Evaluate(`base => {
		const targetMetrics = base.targetMetrics;
		const scopeComplete =
			base.visibleIntendedTargetCount === base.targetMetrics.length &&
			base.visibleFamilyTargetCount === 6 && base.auditedFamilyTargetCount === 6 &&
			base.visibleLocalSidebarLinkCount > 0 && base.auditedLocalSidebarLinkCount === base.visibleLocalSidebarLinkCount &&
			base.visibleSearchControlCount > 0 && base.auditedSearchControlCount === base.visibleSearchControlCount;
		return {
			...base,
			targetMetrics,
			scopeComplete,
			allVisibleIntendedTargetsAtLeast44: scopeComplete && targetMetrics.every(item => item.width >= 43.5 && item.height >= 43.5),
		};
	}`, base)
	if err != nil {
		t.Fatalf("collect visible shell touch targets: %v", err)
	}
	metrics, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("touch target metrics should be an object, got %#v", result)
	}
	return metrics
}

func collectMatrixMetrics(t *testing.T, page playwright.Page, width int, theme string, dark bool, focus map[string]any) map[string]any {
	t.Helper()
	expectedHeader := 64
	if width >= 720 && width < 1440 {
		expectedHeader = 108
	}
	result, err := page.Evaluate(`({width, theme, dark, expectedHeader, focus}) => {
		const root = document.documentElement;
		const body = document.body;
		const header = document.querySelector('.component-doc-shell__header');
		const sidebar = document.querySelector('.component-doc-shell__sidebar');
		const backdrop = document.querySelector('.component-doc-shell__backdrop');
		const toc = document.querySelector('.component-doc-shell__toc-inner');
		const inline = document.querySelector('.component-doc-shell__family-links');
		const disclosure = document.querySelector('.component-doc-shell__family-menu');
		const disclosureLinks = document.querySelector('#componentdocshell-family-listbox');
		const familySelectRoot = document.querySelector('.component-doc-shell__family-select [data-select-config]');
		const familyTrigger = document.querySelector('#componentdocshell-family-trigger');
		const familyPopup = disclosureLinks?.parentElement;
		const scope = document.querySelector('.component-doc-shell__scope');
		const scopeDetails = document.querySelector('.component-doc-shell__scope-details');
		const scopeModule = document.querySelector('.component-doc-shell__scope-module');
		const scopeVersion = document.querySelector('.component-doc-shell__scope-version');
		const scopeVersionBadge = document.querySelector('.component-doc-shell__scope-version-badge');
		const compactFamily = width < 720;
		const drawerSidebar = width < 1024;
		const expectedLabels = ['Components', 'Charts', 'App Shells', 'Icons', 'LLMs', 'Examples'];
		const expectedHrefs = ['/components', '/charts', '/app-shells', '/icons', '/llms', '/examples'];
		const visible = (element) => {
			if (!element) return false;
			const style = getComputedStyle(element);
			if (style.display === 'none' || style.visibility === 'hidden') return false;
			const rect = element.getBoundingClientRect();
			return rect.width > 0 && rect.height > 0 && rect.right > 0 && rect.bottom > 0 && rect.left < innerWidth && rect.top < innerHeight;
		};
		const label = (element) => (element?.getAttribute('aria-label') || element?.textContent || element?.querySelector('img')?.alt || '').trim();
		const surface = compactFamily ? disclosureLinks : inline;
		const familyLinks = Array.from(surface?.querySelectorAll(compactFamily ? '[role="option"]' : 'a') || []);
		const selectOptions = compactFamily && familySelectRoot ? window.goshtosoParseData(familySelectRoot.dataset.selectConfig, {}).options || [] : [];
		const familyHrefs = compactFamily ? selectOptions.map(item => item.value) : familyLinks.map(element => element.getAttribute('href'));
		const activeFamilies = familyLinks.filter(element => compactFamily ? element.getAttribute('aria-selected') === 'true' : element.getAttribute('aria-current') === 'location');
		const localNavigation = sidebar?.querySelector('nav[aria-label="sidebar navigation"]');
		const localLinks = Array.from(localNavigation?.querySelectorAll('a[href]') || []);
		const desktopTheme = document.querySelector('#componentdocshell-theme-trigger');
		const mobileTheme = document.querySelector('#componentdocshell-theme-mobile-trigger');
		const darkButtons = Array.from(document.querySelectorAll('#componentdocshell-dark-mode')).filter(visible);
		const required = [
			document.querySelector('.component-doc-shell__brand'),
			document.querySelector('#componentdocshell-dark-mode'),
			drawerSidebar ? document.querySelector('.component-doc-shell__menu-button') : null,
			compactFamily ? document.querySelector('#componentdocshell-family-trigger') : null,
			compactFamily ? mobileTheme : desktopTheme,
			compactFamily ? null : document.querySelector('.component-doc-shell__repository'),
			document.querySelector('.component-doc-shell__sidebar input[type="search"]'),
			document.querySelector('a.component-doc-shell__scope-module'),
			document.querySelector('.component-doc-shell__scope-version'),
		].filter(Boolean).concat(familyLinks, localLinks);
		const headerRect = header.getBoundingClientRect();
		const familyNavigation = document.querySelector('.component-doc-shell__family-navigation');
		const familyNavigationRect = familyNavigation.getBoundingClientRect();
		const sidebarRect = sidebar.getBoundingClientRect();
		const backdropRect = backdrop.getBoundingClientRect();
		const scopeModuleRect = scopeModule.getBoundingClientRect();
		const scopeVersionRect = scopeVersion.getBoundingClientRect();
		const scopeRect = scope.getBoundingClientRect();
		const localNavigationRect = localNavigation.getBoundingClientRect();
		const sidebarStyle = getComputedStyle(sidebar);
		const localNavigationStyle = getComputedStyle(localNavigation);
		const familyLabelMetrics = familyLinks.map(element => ({
			label: label(element),
			clientWidth: element.clientWidth,
			scrollWidth: element.scrollWidth,
			visible: visible(element),
		}));
		const familyTriggerRect = familyTrigger?.getBoundingClientRect();
		const familyTriggerLabelRect = familyTrigger?.querySelector(':scope > span')?.getBoundingClientRect();
		const familyTriggerStyle = familyTrigger ? getComputedStyle(familyTrigger) : null;
		const familyPopupRect = familyPopup?.getBoundingClientRect();
		const familyOptionLabelsCentered = familyLinks.every(element => {
			const optionRect = element.getBoundingClientRect();
			const optionLabelRect = element.querySelector(':scope > span')?.getBoundingClientRect();
			return optionLabelRect && Math.abs(
				(optionLabelRect.left + optionLabelRect.right) / 2 -
				(optionRect.left + optionRect.right) / 2
			) <= 1;
		});
		const selectedOption = activeFamilies[0];
		const selectedOptionRect = selectedOption?.getBoundingClientRect();
		const selectedIndicator = selectedOption?.querySelector(':scope > svg');
		const selectedIndicatorRect = selectedIndicator?.getBoundingClientRect();
		return {
			width,
			theme,
			dark,
			expectedHeader,
			headerHeight: headerRect.height,
			viewportMatches: innerWidth === width,
			headerHeightMatches: Math.abs(headerRect.height - expectedHeader) <= 0.5,
			familyNavigationContainedByHeader: familyNavigationRect.top >= headerRect.top - 0.5 && familyNavigationRect.bottom <= headerRect.bottom + 0.5,
			familyNavigationUsesHeaderSurface: getComputedStyle(familyNavigation).backgroundColor === 'rgba(0, 0, 0, 0)',
			familyNavigationHasNoDivider: getComputedStyle(familyNavigation).borderTopWidth === '0px',
			scopeHasNoBottomBorder: getComputedStyle(scope).borderBottomWidth === '0px',
			scopeIsOpaqueInDrawer: !drawerSidebar || getComputedStyle(scope).backgroundColor !== 'rgba(0, 0, 0, 0)',
			sidebarEdgeContinuousFromHeader: drawerSidebar
				? parseFloat(sidebarStyle.borderRightWidth) === 1 && parseFloat(getComputedStyle(scope).borderRightWidth) === 0 && Math.abs(scopeRect.top - headerRect.bottom) <= 0.5 && Math.abs(scopeRect.right - (sidebarRect.right - 1)) <= 0.5 && Math.abs(scopeRect.bottom - localNavigationRect.top) <= 0.5
				: parseFloat(getComputedStyle(scope).borderRightWidth) === 1 && getComputedStyle(scope).borderRightColor === localNavigationStyle.borderRightColor && Math.abs(scopeRect.top - headerRect.bottom) <= 0.5 && Math.abs(scopeRect.right - localNavigationRect.right) <= 0.5 && Math.abs(scopeRect.bottom - localNavigationRect.top) <= 0.5,
			mobileDrawerOwnsSurface: !drawerSidebar || (sidebarRect.width <= 320.5 && sidebarStyle.backgroundColor !== 'rgba(0, 0, 0, 0)' && parseFloat(sidebarStyle.borderRightWidth) === 1 && sidebarStyle.boxShadow !== 'none' && parseFloat(localNavigationStyle.borderRightWidth) === 0 && getComputedStyle(scope).backgroundColor !== 'rgba(0, 0, 0, 0)'),
			scopeDetailsUseTwoColumns: getComputedStyle(scopeDetails).display === 'grid' && scopeModuleRect.right <= scopeVersionRect.left + 0.5 && Math.abs(scopeModuleRect.top - scopeVersionRect.top) <= 0.5,
			scopeModuleIsLinkedSlug: scopeModule.tagName === 'A' && scopeModule.getAttribute('href') === 'https://github.com/araihu/goshtoso' && scopeModule.textContent.trim() === 'araihu/goshtoso',
			scopeVersionIsNeutralBadge: scopeVersionBadge.tagName === 'SPAN' && parseFloat(getComputedStyle(scopeVersionBadge).borderTopWidth) === 1 && getComputedStyle(scopeVersionBadge).backgroundColor !== 'rgba(0, 0, 0, 0)',
			documentOverflow: root.scrollWidth > innerWidth,
			bodyOverflow: body.scrollWidth > innerWidth,
			noHorizontalOverflow: root.scrollWidth <= innerWidth && body.scrollWidth <= innerWidth,
			themeMatches: root.dataset.theme === theme,
			darkMatches: root.classList.contains('dark') === dark,
			familyInlineVisible: visible(inline),
			familyDisclosureVisible: visible(disclosure),
			familySurfaceCount: Number(visible(inline)) + Number(visible(disclosure)),
			expectedFamilySurface: compactFamily ? visible(disclosure) && !visible(inline) : visible(inline) && !visible(disclosure),
			localMenuTriggerVisible: visible(document.querySelector('.component-doc-shell__menu-button')),
			expectedLocalMenuTrigger: visible(document.querySelector('.component-doc-shell__menu-button')) === drawerSidebar,
			sidebarVisible: visible(sidebar),
			sidebarPersistent: drawerSidebar ? getComputedStyle(sidebar).position === 'fixed' && sidebar.classList.contains('is-open') : getComputedStyle(sidebar).position === 'static',
			backdropVisible: visible(backdrop),
			expectedBackdrop: visible(backdrop) === drawerSidebar,
			themeSelectorVisibleCount: Number(visible(desktopTheme)) + Number(visible(mobileTheme)),
			themeSelectorsRemoved: !desktopTheme && !mobileTheme,
			darkModeVisibleCount: darkButtons.length,
			exactlyOneVisibleDarkMode: darkButtons.length === 1,
			sidebarTop: sidebarRect.top,
			sidebarTopMatches: Math.abs(sidebarRect.top - headerRect.bottom) <= 0.5,
			backdropTop: backdropRect.top,
			backdropTopMatches: !drawerSidebar || Math.abs(backdropRect.top - headerRect.bottom) <= 0.5,
			tocComputedTop: parseFloat(getComputedStyle(toc).top),
			tocTopMatches: Math.abs(parseFloat(getComputedStyle(toc).top) - expectedHeader) <= 0.5,
			familyLabelMetrics,
			familyLabelsNotClipped: familyLabelMetrics.every(item => item.visible && item.scrollWidth <= item.clientWidth + 0.5),
			familyPopupMatchesTrigger: !compactFamily || (
				visible(familyTrigger) && visible(familyPopup) &&
				Math.abs(familyPopupRect.left - familyTriggerRect.left) <= 0.5 &&
				Math.abs(familyPopupRect.right - familyTriggerRect.right) <= 0.5 &&
				Math.abs(familyPopupRect.width - familyTriggerRect.width) <= 0.5
			),
			familyTriggerLabelCentered: !compactFamily || Boolean(familyTriggerLabelRect && Math.abs(
				(familyTriggerLabelRect.left + familyTriggerLabelRect.right) / 2 -
				(familyTriggerRect.left + familyTriggerRect.right) / 2
			) <= 1),
			familyTriggerSignalsInteractivity: !compactFamily || Boolean(
				familyTriggerStyle &&
				parseFloat(familyTriggerStyle.borderTopWidth) >= 1 &&
				familyTriggerStyle.borderTopColor !== 'rgba(0, 0, 0, 0)' &&
				familyTriggerStyle.backgroundColor !== 'rgba(0, 0, 0, 0)'
			),
			familyTriggerSignalsOpenState: !compactFamily || Boolean(
				familyTrigger.getAttribute('aria-expanded') === 'true' &&
				familyTriggerStyle.backgroundColor !== getComputedStyle(header).backgroundColor
			),
			familyOptionLabelsCentered: !compactFamily || familyOptionLabelsCentered,
			familySelectedIndicatorHasFixedTrailingSlot: !compactFamily || Boolean(
				selectedIndicatorRect && selectedOptionRect &&
				getComputedStyle(selectedIndicator).position === 'absolute' &&
				Math.abs(selectedOptionRect.right - selectedIndicatorRect.right - 16) <= 1
			),
			familyPopupContainedByViewport: !compactFamily || Boolean(
				familyPopupRect && familyPopupRect.left >= -0.5 && familyPopupRect.right <= innerWidth + 0.5
			),
			familyLabels: familyLinks.map(label),
			familyHrefs,
			allSixFamilies: familyLinks.length === 6 && expectedLabels.every((value, index) => label(familyLinks[index]) === value) && expectedHrefs.every((value, index) => familyHrefs[index] === value),
			activeLocationCount: activeFamilies.length,
			uniqueActiveLocation: activeFamilies.length === 1,
			activePageCount: localLinks.filter(element => element.getAttribute('aria-current') === 'page').length,
			uniqueActivePage: localLinks.filter(element => element.getAttribute('aria-current') === 'page').length === 1,
			requiredControlMetrics: required.map(element => ({label: label(element), visible: visible(element), tag: element.tagName.toLowerCase()})),
			visibleLabelledRequiredControls: required.every(element => visible(element) && label(element).length > 0),
			focus,
			focusTreatmentTrusted: focus.trusted === true,
		};
	}`, map[string]any{
		"width": width, "theme": theme, "dark": dark, "expectedHeader": expectedHeader, "focus": focus,
	})
	if err != nil {
		t.Fatalf("collect matrix metrics: %v", err)
	}
	metrics, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("matrix metrics should be an object, got %#v", result)
	}
	return metrics
}

func assertMatrixMetrics(t *testing.T, metrics map[string]any) {
	t.Helper()
	requiredTrue := []string{
		"viewportMatches", "headerHeightMatches", "familyNavigationContainedByHeader", "familyNavigationUsesHeaderSurface", "familyNavigationHasNoDivider", "scopeHasNoBottomBorder", "scopeIsOpaqueInDrawer", "sidebarEdgeContinuousFromHeader", "mobileDrawerOwnsSurface", "scopeDetailsUseTwoColumns", "scopeModuleIsLinkedSlug", "scopeVersionIsNeutralBadge", "noHorizontalOverflow", "themeMatches", "darkMatches",
		"expectedFamilySurface", "expectedLocalMenuTrigger", "sidebarVisible", "sidebarPersistent", "expectedBackdrop",
		"themeSelectorsRemoved", "exactlyOneVisibleDarkMode", "sidebarTopMatches", "backdropTopMatches", "tocTopMatches",
		"familyLabelsNotClipped", "familyPopupMatchesTrigger", "familyTriggerLabelCentered", "familyTriggerSignalsInteractivity", "familyTriggerSignalsOpenState", "familyOptionLabelsCentered", "familySelectedIndicatorHasFixedTrailingSlot", "familyPopupContainedByViewport", "allSixFamilies", "uniqueActiveLocation", "uniqueActivePage",
		"visibleLabelledRequiredControls", "touchTargetsAtLeast44", "touchTargetScopeComplete", "focusTreatmentTrusted",
	}
	var failed []string
	for _, key := range requiredTrue {
		if metrics[key] != true {
			failed = append(failed, key)
		}
	}
	if errors, ok := metrics["browserErrors"].([]string); ok && len(errors) > 0 {
		failed = append(failed, "browserErrors")
	}
	if len(failed) == 0 {
		return
	}
	encoded, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		t.Fatalf("matrix failed fields=%v metrics=%#v", failed, metrics)
	}
	t.Fatalf("matrix failed fields=%v; metrics=%s", failed, encoded)
}

func metricNumber(value any) float64 {
	switch number := value.(type) {
	case int:
		return float64(number)
	case int64:
		return float64(number)
	case float64:
		return number
	default:
		return -1
	}
}

func failWithMetrics(t *testing.T, label string, metrics any) {
	t.Helper()
	encoded, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		t.Fatalf("%s failed: %#v", label, metrics)
	}
	t.Fatalf("%s failed; metrics=%s", label, encoded)
}

func assertThemeSelectorAbsent(t *testing.T, page playwright.Page, phase string) {
	t.Helper()
	result, err := page.Evaluate(`() => ({
		triggers: document.querySelectorAll('[id^="componentdocshell-theme"][id$="-trigger"]').length,
		listboxes: document.querySelectorAll('[id^="componentdocshell-theme"][id$="-listbox"]').length,
	})`)
	if err != nil {
		t.Fatalf("%s theme selector absence evaluation: %v", phase, err)
	}
	metrics := result.(map[string]any)
	if metricNumber(metrics["triggers"]) != 0 || metricNumber(metrics["listboxes"]) != 0 {
		failWithMetrics(t, phase+" theme selector absence", metrics)
	}
}

func assertNoBrowserFailures(t *testing.T, page playwright.Page, failures *browserFailures, label string) {
	t.Helper()
	page.WaitForTimeout(50)
	if messages := failures.snapshot(); len(messages) > 0 {
		failWithMetrics(t, label+" browser errors", messages)
	}
}

func familyIdentity(t *testing.T, page playwright.Page) map[string]any {
	t.Helper()
	result, err := page.Evaluate(`() => ({
		title: document.title,
		heading: document.querySelector('#main-content h1')?.textContent.trim(),
		family: document.querySelector('.component-doc-shell__family-links [aria-current="location"]')?.textContent.trim(),
		scope: document.querySelector('.component-doc-shell__scope-family')?.textContent.trim(),
		focus: document.activeElement?.textContent.trim(),
		path: location.pathname,
		activePages: document.querySelectorAll('.component-doc-shell__sidebar [aria-current="page"]').length,
		pageScroll: document.querySelector('#page-scroll')?.scrollTop,
	})`)
	if err != nil {
		t.Fatalf("read family identity: %v", err)
	}
	return result.(map[string]any)
}

func waitForFamilyIdentity(t *testing.T, page playwright.Page, label, path string) {
	t.Helper()
	if _, err := page.WaitForFunction(`([label, path]) =>
		location.pathname === path &&
		document.querySelector('#main-content h1')?.textContent.trim() === label &&
		document.querySelector('.component-doc-shell__family-links [aria-current="location"]')?.textContent.trim() === label &&
		document.querySelector('.component-doc-shell__scope-family')?.textContent.trim() === label &&
		document.activeElement === document.querySelector('#main-content h1')`, []any{label, path}); err != nil {
		failWithMetrics(t, "wait for family identity", familyIdentity(t, page))
	}
}

func waitForHeadingFocus(t *testing.T, page playwright.Page, heading, path string) {
	t.Helper()
	if _, err := page.WaitForFunction(`([heading, path]) =>
		location.pathname === path &&
		document.querySelector('#main-content h1')?.textContent.trim() === heading &&
		document.activeElement === document.querySelector('#main-content h1')`, []any{heading, path}); err != nil {
		failWithMetrics(t, "wait for page heading focus", familyIdentity(t, page))
	}
}

func assertFamilyIdentity(t *testing.T, page playwright.Page, label, path string) {
	t.Helper()
	got := familyIdentity(t, page)
	wantTitle := label + " Documentation - Goshtoso"
	if got["title"] != wantTitle || got["heading"] != label || got["family"] != label || got["scope"] != label || got["focus"] != label || got["path"] != path || metricNumber(got["activePages"]) != 1 {
		failWithMetrics(t, "family identity", got)
	}
}

func assertPageScrollReset(t *testing.T, page playwright.Page) {
	t.Helper()
	identity := familyIdentity(t, page)
	if metricNumber(identity["pageScroll"]) != 0 {
		failWithMetrics(t, "page scroll reset", identity)
	}
}

func closedDrawerFocusMetrics(t *testing.T, page playwright.Page) map[string]any {
	t.Helper()
	if err := page.Locator("#componentdocshell-dark-mode").Focus(); err != nil {
		t.Fatal(err)
	}
	if err := page.Keyboard().Press("Tab"); err != nil {
		t.Fatal(err)
	}
	result, err := page.Evaluate(`() => {
		const sidebar = document.querySelector('.component-doc-shell__sidebar');
		const active = document.activeElement;
		const candidates = Array.from(sidebar.querySelectorAll('a[href], button:not([disabled]), input:not([disabled]), summary, [tabindex]'))
			.filter(element => element.tabIndex >= 0);
		return {
			drawerOpen: sidebar.classList.contains('is-open'),
			potentiallyFocusableCount: candidates.length,
			activeInsideClosedDrawer: sidebar.contains(active),
			activeLabel: (active.getAttribute?.('aria-label') || active.textContent || '').trim(),
			hiddenItemsExcluded: !sidebar.contains(active),
		};
	}`)
	if err != nil {
		t.Fatal(err)
	}
	return result.(map[string]any)
}

func assertFamilySelectState(t *testing.T, page playwright.Page, open, triggerFocused bool) {
	t.Helper()
	result, err := page.Evaluate(`() => ({
		open: document.querySelector('#componentdocshell-family-trigger').getAttribute('aria-expanded') === 'true',
		triggerFocused: document.activeElement === document.querySelector('#componentdocshell-family-trigger'),
	})`)
	if err != nil {
		t.Fatal(err)
	}
	metrics := result.(map[string]any)
	if metrics["open"] != open || metrics["triggerFocused"] != triggerFocused {
		failWithMetrics(t, "Goshtoso family select", metrics)
	}
}

func waitForDrawer(t *testing.T, page playwright.Page, open bool) {
	t.Helper()
	if _, err := page.WaitForFunction(`open => {
		const sidebar = document.querySelector('.component-doc-shell__sidebar');
		const rect = sidebar.getBoundingClientRect();
		return sidebar.classList.contains('is-open') === open && (open ? rect.left >= -0.5 : rect.right <= 0.5);
	}`, open); err != nil {
		state, _ := page.Evaluate(`() => {
			const sidebar = document.querySelector('.component-doc-shell__sidebar');
			const rect = sidebar.getBoundingClientRect();
			return {className: sidebar.className, left: rect.left, right: rect.right, active: document.activeElement?.outerHTML};
		}`)
		failWithMetrics(t, "wait for drawer", state)
	}
}

func activeMatches(t *testing.T, page playwright.Page, selector string) bool {
	t.Helper()
	result, err := page.Evaluate(`selector => document.activeElement?.matches(selector) || false`, selector)
	if err != nil {
		t.Fatal(err)
	}
	return result == true
}

func drawerTrapAndScrollMetrics(t *testing.T, page playwright.Page) map[string]any {
	t.Helper()
	lastFocusable := page.Locator(`.component-doc-shell__sidebar a[href], .component-doc-shell__sidebar button:not([disabled]), .component-doc-shell__sidebar input:not([disabled])`).Last()
	if err := lastFocusable.ScrollIntoViewIfNeeded(); err != nil {
		t.Fatal(err)
	}
	if err := lastFocusable.Focus(); err != nil {
		t.Fatal(err)
	}
	reachability, err := page.Evaluate(`() => {
		const content = document.querySelector('.component-doc-shell__sidebar-content');
		const sidebar = document.querySelector('.component-doc-shell__sidebar');
		const focusables = Array.from(sidebar.querySelectorAll('a[href], button:not([disabled]), input:not([disabled])')).filter(element => element.tabIndex >= 0);
		const lastTarget = focusables.at(-1);
		const sidebarRect = sidebar.getBoundingClientRect();
		const lastTargetRect = lastTarget.getBoundingClientRect();
		const documentScroll = Math.max(document.documentElement.scrollTop, document.body.scrollTop);
		const maxScroll = content.scrollHeight - content.clientHeight;
		return {
			outerMaxScroll: maxScroll,
			outerScrollTop: content.scrollTop,
			outerOwnerScrollable: maxScroll > 0,
			outerOwnerSupportsScrolling: ['auto', 'scroll'].includes(getComputedStyle(content).overflowY),
			outerOwnerScrolled: content.scrollTop > 0,
			lastTargetFullyWithinSidebar: lastTargetRect.top >= sidebarRect.top && lastTargetRect.bottom <= sidebarRect.bottom + 0.5,
			lastTargetFullyWithinViewport: lastTargetRect.top >= 0 && lastTargetRect.bottom <= innerHeight + 0.5,
			documentScroll,
			documentStayedFixed: documentScroll === 0,
		};
	}`)
	if err != nil {
		t.Fatal(err)
	}
	if err := page.Keyboard().Press("Tab"); err != nil {
		t.Fatal(err)
	}
	forwardWrap, err := page.Evaluate(`() => {
		const sidebar = document.querySelector('.component-doc-shell__sidebar');
		const documentScroll = Math.max(document.documentElement.scrollTop, document.body.scrollTop);
		return {
			activeLabel: (document.activeElement?.getAttribute('aria-label') || document.activeElement?.textContent || '').trim(),
			focusStayedInside: sidebar.contains(document.activeElement),
			wrappedFromLastTarget: document.activeElement !== Array.from(sidebar.querySelectorAll('a[href], button:not([disabled]), input:not([disabled])')).filter(element => element.tabIndex >= 0).at(-1),
			documentScroll,
			documentStayedFixed: documentScroll === 0,
		};
	}`)
	if err != nil {
		t.Fatal(err)
	}
	firstFocusable := page.Locator("a.component-doc-shell__scope-module")
	if err := firstFocusable.Focus(); err != nil {
		t.Fatal(err)
	}
	if err := page.Keyboard().Press("Shift+Tab"); err != nil {
		t.Fatal(err)
	}
	reverseWrap, err := page.Evaluate(`() => {
		const sidebar = document.querySelector('.component-doc-shell__sidebar');
		const lastTarget = Array.from(sidebar.querySelectorAll('a[href], button:not([disabled]), input:not([disabled])')).filter(element => element.tabIndex >= 0).at(-1);
		const documentScroll = Math.max(document.documentElement.scrollTop, document.body.scrollTop);
		return {
			activeLabel: (document.activeElement?.getAttribute('aria-label') || document.activeElement?.textContent || '').trim(),
			focusStayedInside: sidebar.contains(document.activeElement),
			wrappedToLastTarget: document.activeElement === lastTarget,
			documentScroll,
			documentStayedFixed: documentScroll === 0,
		};
	}`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := page.Evaluate(`() => document.querySelector('#main-content').focus()`); err != nil {
		t.Fatal(err)
	}
	page.WaitForTimeout(50)
	outsideFocusRecapture, err := page.Evaluate(`() => {
		const sidebar = document.querySelector('.component-doc-shell__sidebar');
		const outsideTarget = document.querySelector('#main-content');
		const documentScroll = Math.max(document.documentElement.scrollTop, document.body.scrollTop);
		return {
			attemptedTarget: outsideTarget.id,
			activeLabel: (document.activeElement?.getAttribute('aria-label') || document.activeElement?.textContent || '').trim(),
			outsideTargetRejected: document.activeElement !== outsideTarget,
			focusRecapturedInside: sidebar.contains(document.activeElement),
			documentScroll,
			documentStayedFixed: documentScroll === 0,
		};
	}`)
	if err != nil {
		t.Fatal(err)
	}
	return map[string]any{
		"utilityReachability":   reachability,
		"forwardWrap":           forwardWrap,
		"reverseWrap":           reverseWrap,
		"outsideFocusRecapture": outsideFocusRecapture,
	}
}

func responsiveTrapMetrics(t *testing.T, page playwright.Page) map[string]any {
	t.Helper()
	if err := page.SetViewportSize(1023, 900); err != nil {
		t.Fatal(err)
	}
	if err := page.Locator(".component-doc-shell__menu-button").Click(); err != nil {
		t.Fatal(err)
	}
	waitForDrawer(t, page, true)
	innerLink := page.Locator(`.component-doc-shell__sidebar a[href="/components/button"]`)
	if err := innerLink.Focus(); err != nil {
		t.Fatal(err)
	}
	if err := page.SetViewportSize(1024, 900); err != nil {
		t.Fatal(err)
	}
	if _, err := page.WaitForFunction(`() => {
		const state = window.Alpine?.$data(document.documentElement);
		const sidebar = document.querySelector('.component-doc-shell__sidebar');
		const rect = sidebar?.getBoundingClientRect();
		return matchMedia('(min-width: 1024px)').matches &&
			state?.sidebarPersistent === true &&
			getComputedStyle(sidebar).position === 'static' &&
			rect.left >= -0.5;
	}`, nil); err != nil {
		t.Fatal(err)
	}

	persistentBeforeFocus, err := page.Evaluate(`() => {
		const state = window.Alpine.$data(document.documentElement);
		const sidebar = document.querySelector('.component-doc-shell__sidebar');
		const trigger = document.querySelector('.component-doc-shell__menu-button');
		return {
			sidebarPersistent: state.sidebarPersistent,
			sidebarOpen: state.sidebarOpen,
			sidebarInert: sidebar.inert,
			triggerVisible: trigger.getClientRects().length > 0 && getComputedStyle(trigger).display !== 'none',
			triggerFocused: document.activeElement === trigger,
			activeInsideSidebar: sidebar.contains(document.activeElement),
		};
	}`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := page.Evaluate(`() => document.querySelector('#main-content h1').focus()`); err != nil {
		t.Fatal(err)
	}
	page.WaitForTimeout(50)
	persistentAfterFocus, err := page.Evaluate(`() => {
		const state = window.Alpine.$data(document.documentElement);
		const sidebar = document.querySelector('.component-doc-shell__sidebar');
		const trigger = document.querySelector('.component-doc-shell__menu-button');
		const heading = document.querySelector('#main-content h1');
		return {
			sidebarPersistent: state.sidebarPersistent,
			sidebarOpen: state.sidebarOpen,
			sidebarInert: sidebar.inert,
			headingFocused: document.activeElement === heading,
			activeInsideSidebar: sidebar.contains(document.activeElement),
			activeLabel: (document.activeElement?.getAttribute('aria-label') || document.activeElement?.textContent || '').trim(),
			triggerVisible: trigger.getClientRects().length > 0 && getComputedStyle(trigger).display !== 'none',
			triggerFocused: document.activeElement === trigger,
		};
	}`)
	if err != nil {
		t.Fatal(err)
	}

	if err := page.SetViewportSize(1023, 900); err != nil {
		t.Fatal(err)
	}
	if _, err := page.WaitForFunction(`() => window.Alpine?.$data(document.documentElement)?.sidebarPersistent === false`, nil); err != nil {
		t.Fatal(err)
	}
	returnedToMobile, err := page.Evaluate(`() => {
		const state = window.Alpine.$data(document.documentElement);
		const sidebar = document.querySelector('.component-doc-shell__sidebar');
		return {
			sidebarOpen: state.sidebarOpen,
			sidebarInert: sidebar.inert,
			closedAndInert: state.sidebarOpen === false && sidebar.inert === true,
		};
	}`)
	if err != nil {
		t.Fatal(err)
	}
	return map[string]any{
		"persistentBeforeFocus": persistentBeforeFocus,
		"persistentAfterFocus":  persistentAfterFocus,
		"returnedToMobile":      returnedToMobile,
	}
}

func chooseTheme(t *testing.T, page playwright.Page, selectorID, label, value string) {
	t.Helper()
	trigger := page.Locator("#" + selectorID + "-trigger")
	if err := trigger.Click(); err != nil {
		t.Fatalf("open %s theme selector: %v", selectorID, err)
	}
	list := page.Locator("#" + selectorID + "-listbox")
	if err := list.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}); err != nil {
		t.Fatalf("wait for %s theme list: %v", selectorID, err)
	}
	options := list.Locator(`[role="option"]`)
	labels, err := options.AllTextContents()
	if err != nil {
		t.Fatal(err)
	}
	index := -1
	for candidate, text := range labels {
		if strings.TrimSpace(text) == label {
			index = candidate
			break
		}
	}
	if index < 0 {
		t.Fatalf("theme %q absent from %s options: %v", label, selectorID, labels)
	}
	option := options.Nth(index)
	if err := option.ScrollIntoViewIfNeeded(); err != nil {
		t.Fatal(err)
	}
	if err := option.Click(); err != nil {
		t.Fatal(err)
	}
	if _, err := page.WaitForFunction(`value => document.documentElement.dataset.theme === value`, value); err != nil {
		t.Fatalf("wait for theme %s: %v", value, err)
	}
}

func testThemeSynchronization(t *testing.T, harness *browserHarness) {
	t.Helper()
	page := harness.newPage(t, true)
	if err := page.SetViewportSize(390, 900); err != nil {
		t.Fatal(err)
	}
	failures := watchBrowserFailures(page)
	gotoFamilyPage(t, page, harness.baseURL, "/__e2e/persistent")
	if err := page.Locator(".component-doc-shell__menu-button").Click(); err != nil {
		t.Fatal(err)
	}
	waitForDrawer(t, page, true)
	chooseTheme(t, page, "componentdocshell-theme-mobile", "Minimal", "minimal")
	if err := page.Locator(".component-doc-shell__menu-button").Click(); err != nil {
		t.Fatal(err)
	}
	waitForDrawer(t, page, false)
	if err := page.SetViewportSize(841, 900); err != nil {
		t.Fatal(err)
	}
	if _, err := page.WaitForFunction(`() => {
		const trigger = document.querySelector('#componentdocshell-theme-trigger');
		return trigger?.textContent.includes('Minimal') && getComputedStyle(trigger).display !== 'none';
	}`, nil); err != nil {
		t.Fatal(err)
	}
	chooseTheme(t, page, "componentdocshell-theme", "Goshtoso", "goshtoso")
	if err := page.SetViewportSize(390, 900); err != nil {
		t.Fatal(err)
	}
	if err := page.Locator(".component-doc-shell__menu-button").Click(); err != nil {
		t.Fatal(err)
	}
	waitForDrawer(t, page, true)
	result, err := page.Evaluate(`() => ({
		rootTheme: document.documentElement.dataset.theme,
		mobileLabel: document.querySelector('#componentdocshell-theme-mobile-trigger')?.textContent.trim(),
		desktopVisible: document.querySelector('#componentdocshell-theme-trigger')?.getClientRects().length > 0,
		mobileVisible: document.querySelector('#componentdocshell-theme-mobile-trigger')?.getClientRects().length > 0,
	})`)
	if err != nil {
		t.Fatal(err)
	}
	metrics := result.(map[string]any)
	if metrics["rootTheme"] != "goshtoso" || !strings.Contains(fmt.Sprint(metrics["mobileLabel"]), "Goshtoso") || metrics["desktopVisible"] != false || metrics["mobileVisible"] != true {
		failWithMetrics(t, "theme selector synchronization", metrics)
	}
	assertNoBrowserFailures(t, page, failures, "theme synchronization")
}

func testSystemPreference(t *testing.T, harness *browserHarness, scheme *playwright.ColorScheme, dark bool) {
	t.Helper()
	page := harness.newPageWithOptions(t, playwright.BrowserNewContextOptions{
		JavaScriptEnabled: playwright.Bool(true),
		ColorScheme:       scheme,
	})
	failures := watchBrowserFailures(page)
	gotoFamilyPage(t, page, harness.baseURL, "/components")
	result, err := page.Evaluate(`dark => ({
		mediaMatches: matchMedia('(prefers-color-scheme: dark)').matches === dark,
		rootMatches: document.documentElement.classList.contains('dark') === dark,
		contentPresent: document.querySelector('#main-content h1')?.textContent.trim() === 'Components',
	})`, dark)
	if err != nil {
		t.Fatal(err)
	}
	metrics := result.(map[string]any)
	if metrics["mediaMatches"] != true || metrics["rootMatches"] != true || metrics["contentPresent"] != true {
		failWithMetrics(t, "system color preference", metrics)
	}
	assertNoBrowserFailures(t, page, failures, "system color preference")
}

func testThrowingStorage(t *testing.T, harness *browserHarness) {
	t.Helper()
	page := harness.newPage(t, true)
	if err := page.SetViewportSize(841, 900); err != nil {
		t.Fatal(err)
	}
	script := `
		window.__storageCalls = [];
		Storage.prototype.getItem = function (key) { window.__storageCalls.push('get:' + key); throw new Error('storage disabled'); };
		Storage.prototype.setItem = function (key) { window.__storageCalls.push('set:' + key); throw new Error('storage disabled'); };
	`
	if err := page.AddInitScript(playwright.Script{Content: playwright.String(script)}); err != nil {
		t.Fatal(err)
	}
	failures := watchBrowserFailures(page)
	gotoFamilyPage(t, page, harness.baseURL, "/__e2e/persistent")
	chooseTheme(t, page, "componentdocshell-theme", "Minimal", "minimal")
	if err := page.Locator("#componentdocshell-dark-mode").Click(); err != nil {
		t.Fatal(err)
	}
	if err := page.Locator(`.component-doc-shell__family-links a[href="/charts"]`).Click(); err != nil {
		t.Fatal(err)
	}
	waitForFamilyIdentity(t, page, "Charts", "/charts")
	result, err := page.Evaluate(`() => ({
		storageCalls: window.__storageCalls,
		theme: document.documentElement.dataset.theme,
		dark: document.documentElement.classList.contains('dark'),
		heading: document.querySelector('#main-content h1')?.textContent.trim(),
		families: document.querySelectorAll('.component-doc-shell__family-links a').length,
		themeVisible: document.querySelector('#componentdocshell-theme-trigger')?.getClientRects().length > 0,
	})`)
	if err != nil {
		t.Fatal(err)
	}
	metrics := result.(map[string]any)
	calls := fmt.Sprint(metrics["storageCalls"])
	if !strings.Contains(calls, "get:theme") || !strings.Contains(calls, "set:theme") || !strings.Contains(calls, "set:darkMode") || metrics["theme"] != "minimal" || metrics["dark"] != true || metrics["heading"] != "Charts" || metricNumber(metrics["families"]) != 6 || metrics["themeVisible"] != true {
		failWithMetrics(t, "throwing storage fallback", metrics)
	}
	assertNoBrowserFailures(t, page, failures, "throwing storage fallback")
}

func testTextReflow(t *testing.T, harness *browserHarness, width int, rootFontSize, label string) {
	t.Helper()
	t.Run(label, func(t *testing.T) {
		page := harness.newPage(t, true)
		if err := page.SetViewportSize(width, 900); err != nil {
			t.Fatal(err)
		}
		failures := watchBrowserFailures(page)
		failures.setPhase("initial-load")
		gotoFamilyPage(t, page, harness.baseURL, "/components")
		if rootFontSize != "" {
			if _, err := page.Evaluate(`size => document.documentElement.style.fontSize = size`, rootFontSize); err != nil {
				t.Fatal(err)
			}
		}
		if err := page.Locator("#componentdocshell-family-trigger").Click(); err != nil {
			t.Fatal(err)
		}
		if err := page.Locator(".component-doc-shell__menu-button").Click(); err != nil {
			t.Fatal(err)
		}
		waitForDrawer(t, page, true)
		result, err := page.Evaluate(`({width, label, rootFontSize}) => {
			const root = document.documentElement;
			const body = document.body;
			const header = document.querySelector('.component-doc-shell__header');
			const headerRect = header.getBoundingClientRect();
			const familyTrigger = document.querySelector('#componentdocshell-family-trigger');
			const familyLabel = familyTrigger.querySelector('span');
			const familyChevron = familyTrigger.querySelector('svg');
			const familyTriggerRect = familyTrigger.getBoundingClientRect();
			const familyLabelRect = familyLabel.getBoundingClientRect();
			const familyChevronRect = familyChevron.getBoundingClientRect();
			const compactMark = document.querySelector('.component-doc-shell__brand-compact-mark');
			const compactLogo = compactMark?.querySelector('.component-doc-shell__brand-logo');
			const compactMarkRect = compactMark?.getBoundingClientRect();
			const visibleCompactImage = Array.from(compactLogo?.querySelectorAll('img') || []).find(image => getComputedStyle(image).display !== 'none');
			const visibleCompactImageRect = visibleCompactImage?.getBoundingClientRect();
			const familyPanel = document.querySelector('#componentdocshell-family-listbox');
			const familyRect = familyPanel.getBoundingClientRect();
			const controls = ['.component-doc-shell__menu-button', '#componentdocshell-family-trigger', '#componentdocshell-dark-mode']
				.map(selector => document.querySelector(selector).getBoundingClientRect());
			const familyLinks = Array.from(familyPanel.querySelectorAll('[role="option"]'));
			return {
				label,
				width,
				rootFontSize: getComputedStyle(root).fontSize,
				requestedRootFontSize: rootFontSize,
				familyTrigger: {
					fontSize: getComputedStyle(familyTrigger).fontSize,
					fontScalesWithRoot: Math.abs(parseFloat(getComputedStyle(familyTrigger).fontSize) - parseFloat(getComputedStyle(root).fontSize)) <= 0.5,
					label: familyLabel.textContent.trim(),
					labelVisible: familyLabel.textContent.trim().length > 0 && familyLabelRect.width > 0 && familyLabelRect.height > 0,
					labelNotClipped: familyLabel.scrollWidth <= familyLabel.clientWidth + 0.5 && familyLabel.scrollHeight <= familyLabel.clientHeight + 0.5,
					chevronVisible: getComputedStyle(familyChevron).visibility !== 'hidden' && getComputedStyle(familyChevron).display !== 'none' && familyChevronRect.width > 0 && familyChevronRect.height > 0,
					chevronWithinTrigger: familyChevronRect.left >= familyTriggerRect.left - 0.5 && familyChevronRect.right <= familyTriggerRect.right + 0.5 && familyChevronRect.top >= familyTriggerRect.top - 0.5 && familyChevronRect.bottom <= familyTriggerRect.bottom + 0.5,
					chevronWithinViewport: familyChevronRect.left >= 0 && familyChevronRect.right <= innerWidth + 0.5 && familyChevronRect.top >= 0 && familyChevronRect.bottom <= innerHeight + 0.5,
				},
				compactLogo: {
					present: Boolean(compactLogo),
					visible: Boolean(compactMarkRect && compactMarkRect.width === 32 && compactMarkRect.height === 32 && getComputedStyle(compactMark).display !== 'none'),
					imageVisible: Boolean(visibleCompactImageRect && visibleCompactImageRect.width > 0 && visibleCompactImageRect.height > 0),
					imageContained: Boolean(visibleCompactImageRect && compactMarkRect && visibleCompactImageRect.width <= compactMarkRect.width + 0.5 && visibleCompactImageRect.height <= compactMarkRect.height + 0.5),
					boundedByHeader: Boolean(compactMarkRect && compactMarkRect.top >= headerRect.top - 0.5 && compactMarkRect.bottom <= headerRect.bottom + 0.5),
				},
				familyRect: {top: familyRect.top, right: familyRect.right, bottom: familyRect.bottom, left: familyRect.left, width: familyRect.width, height: familyRect.height},
				familyScroll: {clientHeight: familyPanel.clientHeight, scrollHeight: familyPanel.scrollHeight, scrollTop: familyPanel.scrollTop},
				noHorizontalOverflow: root.scrollWidth <= innerWidth && body.scrollWidth <= innerWidth,
				familyPanelBounded: familyRect.left >= 0 && familyRect.right <= innerWidth + 0.5 && familyRect.top >= 0 && familyRect.bottom <= innerHeight + 0.5,
				themeSelectorsRemoved: document.querySelectorAll('[id^="componentdocshell-theme"][id$="-trigger"]').length === 0,
				controlsRetained: controls.every(rect => rect.width > 0 && rect.height > 0 && rect.left >= 0 && rect.right <= innerWidth + 0.5),
				familyDestinationsReachable: familyLinks.length === 6 && familyLinks.every(link => link.scrollWidth <= link.clientWidth + 0.5) && familyPanel.scrollHeight >= familyLinks.at(-1).offsetTop + familyLinks.at(-1).offsetHeight,
			};
		}`, map[string]any{"width": width, "label": label, "rootFontSize": rootFontSize})
		if err != nil {
			t.Fatal(err)
		}
		metrics := result.(map[string]any)
		if metrics["noHorizontalOverflow"] != true || metrics["familyPanelBounded"] != true || metrics["themeSelectorsRemoved"] != true || metrics["controlsRetained"] != true || metrics["familyDestinationsReachable"] != true {
			failWithMetrics(t, "text/reflow containment", metrics)
		}
		if rootFontSize != "" {
			familyTrigger := metrics["familyTrigger"].(map[string]any)
			compactLogo := metrics["compactLogo"].(map[string]any)
			if familyTrigger["fontScalesWithRoot"] != true || familyTrigger["labelVisible"] != true || familyTrigger["labelNotClipped"] != true || familyTrigger["chevronVisible"] != true || familyTrigger["chevronWithinTrigger"] != true || familyTrigger["chevronWithinViewport"] != true || compactLogo["present"] != true || compactLogo["visible"] != true || compactLogo["imageVisible"] != true || compactLogo["imageContained"] != true || compactLogo["boundedByHeader"] != true {
				failWithMetrics(t, "200% family trigger and compact logo geometry", metrics)
			}
		}
		assertNoBrowserFailures(t, page, failures, "text/reflow containment")
	})
}

func testAccessibilitySemantics(t *testing.T, harness *browserHarness) {
	t.Helper()
	t.Run("chromium_semantics", func(t *testing.T) {
		page := harness.newPage(t, true)
		if err := page.SetViewportSize(841, 900); err != nil {
			t.Fatal(err)
		}
		failures := watchBrowserFailures(page)
		gotoFamilyPage(t, page, harness.baseURL, "/components")
		snapshot, err := page.Locator("body").AriaSnapshot()
		if err != nil {
			t.Fatal(err)
		}
		result, err := page.Evaluate(`() => {
			const visible = element => Boolean(element && element.getClientRects().length && getComputedStyle(element).display !== 'none' && getComputedStyle(element).visibility !== 'hidden');
			const ids = Array.from(document.querySelectorAll('[id]')).map(element => element.id);
			const duplicates = ids.filter((id, index) => ids.indexOf(id) !== index);
			const labelled = element => Boolean((element.getAttribute('aria-label') || element.textContent || element.getAttribute('alt') || '').trim());
			const required = Array.from(document.querySelectorAll('.component-doc-shell__header a, .component-doc-shell__header button, .component-doc-shell__family-links a, .component-doc-shell__sidebar input, .component-doc-shell__sidebar a')).filter(visible);
			const hiddenSurfaces = [document.querySelector('#componentdocshell-family-listbox'), document.querySelector('.component-doc-shell__mobile-utilities')];
			const hiddenFocusable = hiddenSurfaces.flatMap(surface => Array.from(surface?.querySelectorAll('a[href],button,input,summary,[tabindex]') || []))
				.filter(element => element.tabIndex >= 0 && visible(element));
			const visibleFamilyLandmarks = Array.from(document.querySelectorAll('nav[aria-label="Documentation families"]')).filter(visible);
			return {
				duplicateIDs: [...new Set(duplicates)],
				uniqueIDs: duplicates.length === 0,
				visibleFamilyLandmarks: visibleFamilyLandmarks.length,
				oneVisibleFamilyLandmark: visibleFamilyLandmarks.length === 1,
				activeLocations: visibleFamilyLandmarks[0]?.querySelectorAll('[aria-current="location"]').length || 0,
				activePages: document.querySelectorAll('.component-doc-shell__sidebar [aria-current="page"]').length,
				unlabelledRequired: required.filter(element => !labelled(element)).map(element => element.outerHTML),
				allRequiredLabelled: required.every(labelled),
				hiddenFocusable: hiddenFocusable.map(element => element.outerHTML),
				hiddenSurfacesExcluded: hiddenFocusable.length === 0,
			};
		}`)
		if err != nil {
			t.Fatal(err)
		}
		metrics := result.(map[string]any)
		metrics["ariaFamilyLandmarkCount"] = strings.Count(snapshot, `navigation "Documentation families"`)
		metrics["ariaSidebarLandmarkCount"] = strings.Count(snapshot, `navigation "sidebar navigation"`)
		if metrics["uniqueIDs"] != true || metrics["oneVisibleFamilyLandmark"] != true || metricNumber(metrics["activeLocations"]) != 1 || metricNumber(metrics["activePages"]) != 1 || metrics["allRequiredLabelled"] != true || metrics["hiddenSurfacesExcluded"] != true || metricNumber(metrics["ariaFamilyLandmarkCount"]) != 1 || metricNumber(metrics["ariaSidebarLandmarkCount"]) != 1 {
			metrics["ariaSnapshot"] = snapshot
			failWithMetrics(t, "Chromium accessibility semantics", metrics)
		}
		assertNoBrowserFailures(t, page, failures, "Chromium accessibility semantics")
		t.Log("Playwright-Go provides ARIA snapshots but no serious/critical rule scanner; no remote axe runtime or unapproved dependency was added")
	})
}

func TestFamilyNavigationHTMXHistoryAndFocus(t *testing.T) {
	requireE2E(t)
	harness := newBrowserHarness(t)

	t.Run("history_identity", func(t *testing.T) {
		page := harness.newPage(t, true)
		if err := page.SetViewportSize(841, 900); err != nil {
			t.Fatal(err)
		}
		failures := watchBrowserFailures(page)
		failures.setPhase("initial-load")
		gotoFamilyPage(t, page, harness.baseURL, "/components")
		if _, err := page.WaitForFunction(`() => Boolean(window.componentDocShell?.focusMain)`, nil); err != nil {
			t.Fatal(err)
		}
		if _, err := page.Evaluate(`() => window.componentDocShell.focusMain()`); err != nil {
			t.Fatal(err)
		}
		assertFamilyIdentity(t, page, "Components", "/components")
		assertThemeSelectorAbsent(t, page, "initial Components")
		if _, err := page.Evaluate(`() => {
			const content = document.querySelector('#main-content');
			content.style.minHeight = '2000px';
			const scroller = document.querySelector('#page-scroll');
			scroller.scrollTop = 400;
			return scroller.scrollTop;
		}`); err != nil {
			t.Fatal(err)
		}

		failures.setPhase("charts-swap")
		if err := page.Locator(`.component-doc-shell__family-links a[href="/charts"]`).Click(); err != nil {
			t.Fatal(err)
		}
		waitForFamilyIdentity(t, page, "Charts", "/charts")
		assertFamilyIdentity(t, page, "Charts", "/charts")
		assertPageScrollReset(t, page)
		assertThemeSelectorAbsent(t, page, "Charts swap")

		failures.setPhase("history-back")
		if _, err := page.GoBack(); err != nil {
			t.Fatal(err)
		}
		waitForFamilyIdentity(t, page, "Components", "/components")
		assertFamilyIdentity(t, page, "Components", "/components")
		assertPageScrollReset(t, page)
		assertThemeSelectorAbsent(t, page, "history Back")

		failures.setPhase("history-forward")
		if _, err := page.GoForward(); err != nil {
			t.Fatal(err)
		}
		waitForFamilyIdentity(t, page, "Charts", "/charts")
		assertFamilyIdentity(t, page, "Charts", "/charts")
		assertPageScrollReset(t, page)
		assertThemeSelectorAbsent(t, page, "history Forward")
		assertNoBrowserFailures(t, page, failures, "HTMX history")
	})

	t.Run("small_disclosure_and_drawer", func(t *testing.T) {
		page := harness.newPage(t, true)
		if err := page.SetViewportSize(390, 420); err != nil {
			t.Fatal(err)
		}
		failures := watchBrowserFailures(page)
		gotoFamilyPage(t, page, harness.baseURL, "/components")
		if _, err := page.WaitForFunction(`() => Boolean(window.Alpine)`, nil); err != nil {
			t.Fatal(err)
		}

		hiddenFocus := closedDrawerFocusMetrics(t, page)
		trigger := page.Locator("#componentdocshell-family-trigger")
		if err := trigger.Click(); err != nil {
			t.Fatal(err)
		}
		assertFamilySelectState(t, page, true, true)
		if err := trigger.Press("Escape"); err != nil {
			t.Fatal(err)
		}
		assertFamilySelectState(t, page, false, true)
		if err := trigger.Press("Space"); err != nil {
			t.Fatal(err)
		}
		assertFamilySelectState(t, page, true, true)
		darkButton := page.Locator("#componentdocshell-dark-mode")
		if err := darkButton.Click(); err != nil {
			t.Fatal(err)
		}
		outside, err := page.Evaluate(`() => ({
			open: document.querySelector('#componentdocshell-family-trigger').getAttribute('aria-expanded') === 'true',
			clickedTargetKeptFocus: document.activeElement === document.querySelector('#componentdocshell-dark-mode'),
		})`)
		if err != nil {
			t.Fatal(err)
		}
		outsideMetrics := outside.(map[string]any)

		menuButton := page.Locator(".component-doc-shell__menu-button")
		if err := menuButton.Click(); err != nil {
			t.Fatal(err)
		}
		waitForDrawer(t, page, true)
		trapMetrics := drawerTrapAndScrollMetrics(t, page)
		if err := page.Locator(".component-doc-shell__backdrop").Click(playwright.LocatorClickOptions{Position: &playwright.Position{X: 380, Y: 10}}); err != nil {
			t.Fatal(err)
		}
		waitForDrawer(t, page, false)
		overlayFocus := activeMatches(t, page, ".component-doc-shell__menu-button")

		if err := menuButton.Click(); err != nil {
			t.Fatal(err)
		}
		waitForDrawer(t, page, true)
		if err := page.Keyboard().Press("Escape"); err != nil {
			t.Fatal(err)
		}
		waitForDrawer(t, page, false)
		escapeFocus := activeMatches(t, page, ".component-doc-shell__menu-button")

		if err := menuButton.Click(); err != nil {
			t.Fatal(err)
		}
		waitForDrawer(t, page, true)
		buttonLink := page.Locator(`.component-doc-shell__sidebar a[href="/components/button"]`)
		if err := buttonLink.ScrollIntoViewIfNeeded(); err != nil {
			t.Fatal(err)
		}
		if err := buttonLink.Click(); err != nil {
			t.Fatal(err)
		}
		waitForHeadingFocus(t, page, "Button", "/components/button")
		localNavigation, err := page.Evaluate(`() => ({
			drawerClosed: !document.querySelector('.component-doc-shell__sidebar').classList.contains('is-open'),
			headingFocused: document.activeElement === document.querySelector('#main-content h1'),
			activePages: document.querySelectorAll('.component-doc-shell__sidebar [aria-current="page"]').length,
		})`)
		if err != nil {
			t.Fatal(err)
		}
		assertNoBrowserFailures(t, page, failures, "Small disclosure and drawer")
		metrics := map[string]any{
			"hiddenFocus": hiddenFocus, "outside": outsideMetrics, "trap": trapMetrics,
			"overlayFocusReturn": overlayFocus, "escapeFocusReturn": escapeFocus,
			"localNavigation": localNavigation,
		}
		reachability := trapMetrics["utilityReachability"].(map[string]any)
		forwardWrap := trapMetrics["forwardWrap"].(map[string]any)
		reverseWrap := trapMetrics["reverseWrap"].(map[string]any)
		outsideFocusRecapture := trapMetrics["outsideFocusRecapture"].(map[string]any)
		if hiddenFocus["hiddenItemsExcluded"] != true || outsideMetrics["open"] != false || outsideMetrics["clickedTargetKeptFocus"] != true ||
			reachability["outerOwnerSupportsScrolling"] != true || reachability["lastTargetFullyWithinSidebar"] != true || reachability["lastTargetFullyWithinViewport"] != true || reachability["documentStayedFixed"] != true ||
			forwardWrap["focusStayedInside"] != true || forwardWrap["wrappedFromLastTarget"] != true || forwardWrap["documentStayedFixed"] != true ||
			reverseWrap["focusStayedInside"] != true || reverseWrap["wrappedToLastTarget"] != true || reverseWrap["documentStayedFixed"] != true ||
			outsideFocusRecapture["outsideTargetRejected"] != true || outsideFocusRecapture["focusRecapturedInside"] != true || outsideFocusRecapture["documentStayedFixed"] != true ||
			overlayFocus != true || escapeFocus != true ||
			localNavigation.(map[string]any)["drawerClosed"] != true || localNavigation.(map[string]any)["headingFocused"] != true || metricNumber(localNavigation.(map[string]any)["activePages"]) != 1 {
			failWithMetrics(t, "Small disclosure/drawer behavior", metrics)
		}
	})

	t.Run("small_family_select_navigation", func(t *testing.T) {
		page := harness.newPage(t, true)
		if err := page.SetViewportSize(390, 720); err != nil {
			t.Fatal(err)
		}
		failures := watchBrowserFailures(page)
		gotoFamilyPage(t, page, harness.baseURL, "/components")
		trigger := page.Locator("#componentdocshell-family-trigger")
		if err := trigger.Click(); err != nil {
			t.Fatal(err)
		}
		if err := page.Locator(".component-doc-shell__menu-button").Click(); err != nil {
			t.Fatal(err)
		}
		if _, err := page.WaitForFunction(`() => document.querySelector('.component-doc-shell__sidebar').classList.contains('is-open') && document.querySelector('#componentdocshell-family-trigger').getAttribute('aria-expanded') === 'false'`, nil); err != nil {
			t.Fatal(err)
		}
		if err := page.Locator(".component-doc-shell__backdrop").Click(playwright.LocatorClickOptions{Position: &playwright.Position{X: 380, Y: 10}}); err != nil {
			t.Fatal(err)
		}
		waitForDrawer(t, page, false)
		if err := trigger.Click(); err != nil {
			t.Fatal(err)
		}
		if err := page.Locator("#componentdocshell-family-option-1").Click(); err != nil {
			t.Fatal(err)
		}
		if err := page.WaitForURL("**/charts"); err != nil {
			t.Fatal(err)
		}
		if _, err := page.WaitForFunction(`() => document.querySelector('#main-content h1')?.textContent.trim() === 'Charts'`, nil); err != nil {
			t.Fatal(err)
		}
		metrics, err := page.Evaluate(`() => ({
			role: document.querySelector('#componentdocshell-family-trigger')?.getAttribute('role'),
			selected: document.querySelector('#componentdocshell-family-listbox [aria-selected="true"]')?.textContent.trim(),
			path: location.pathname,
			customDetailsCount: document.querySelectorAll('.component-doc-shell__family-menu details').length,
		})`)
		if err != nil {
			t.Fatal(err)
		}
		state := metrics.(map[string]any)
		if state["role"] != "combobox" || state["selected"] != "Charts" || state["path"] != "/charts" || metricNumber(state["customDetailsCount"]) != 0 {
			failWithMetrics(t, "small Goshtoso family select navigation", state)
		}
		assertNoBrowserFailures(t, page, failures, "small Goshtoso family select navigation")
	})

	t.Run("responsive_trap_release", func(t *testing.T) {
		page := harness.newPage(t, true)
		failures := watchBrowserFailures(page)
		gotoFamilyPage(t, page, harness.baseURL, "/components")
		metrics := responsiveTrapMetrics(t, page)
		assertNoBrowserFailures(t, page, failures, "responsive drawer trap release")
		before := metrics["persistentBeforeFocus"].(map[string]any)
		after := metrics["persistentAfterFocus"].(map[string]any)
		returned := metrics["returnedToMobile"].(map[string]any)
		if before["sidebarPersistent"] != true || before["sidebarOpen"] != false || before["sidebarInert"] != false || before["triggerVisible"] != false || before["triggerFocused"] != false ||
			after["sidebarPersistent"] != true || after["sidebarOpen"] != false || after["sidebarInert"] != false || after["headingFocused"] != true || after["activeInsideSidebar"] != false || after["triggerVisible"] != false || after["triggerFocused"] != false ||
			returned["closedAndInert"] != true {
			failWithMetrics(t, "responsive drawer trap release", metrics)
		}
	})

	t.Run("theme_system_and_throwing_storage", func(t *testing.T) {
		testThemeSynchronization(t, harness)
		testSystemPreference(t, harness, playwright.ColorSchemeLight, false)
		testSystemPreference(t, harness, playwright.ColorSchemeDark, true)
		testThrowingStorage(t, harness)
	})

	t.Run("reflow_and_accessibility", func(t *testing.T) {
		testTextReflow(t, harness, 390, "32px", "200-percent-text")
		testTextReflow(t, harness, 320, "", "bounded-400-percent-reflow-proxy")
		testAccessibilitySemantics(t, harness)
	})
}

func TestFamilyNavigationMaximumTextReflow(t *testing.T) {
	requireE2E(t)
	harness := newBrowserHarness(t)
	page := harness.newPage(t, true)
	if err := page.SetViewportSize(320, 900); err != nil {
		t.Fatal(err)
	}
	gotoFamilyPage(t, page, harness.baseURL, "/components")
	if _, err := page.Evaluate(`() => document.documentElement.style.fontSize = '32px'`); err != nil {
		t.Fatal(err)
	}
	page.WaitForTimeout(50)
	result, err := page.Evaluate(`() => {
		const root = document.documentElement;
		const body = document.body;
		const rect = selector => document.querySelector(selector).getBoundingClientRect();
		const header = rect('.component-doc-shell__header');
		const trigger = document.querySelector('#componentdocshell-family-trigger');
		const triggerRect = trigger.getBoundingClientRect();
		const label = trigger.querySelector('span');
		const labelRect = label.getBoundingClientRect();
		const chevron = trigger.querySelector('svg');
		const chevronRect = chevron.getBoundingClientRect();
		const compactMark = document.querySelector('.component-doc-shell__brand-compact-mark');
		const compactRect = compactMark.getBoundingClientRect();
		const targets = ['.component-doc-shell__menu-button', '.component-doc-shell__brand', '#componentdocshell-dark-mode'].map(selector => rect(selector));
		const range = document.createRange();
		range.selectNodeContents(label);
		const textRect = range.getBoundingClientRect();
		return {
			rootFontSize: getComputedStyle(root).fontSize,
			label: label.textContent.trim(),
			labelVisible: labelRect.width > 0 && labelRect.height > 0,
			labelNotClipped: label.scrollWidth <= label.clientWidth + 0.5 && label.scrollHeight <= label.clientHeight + 0.5,
			textWithinLabel: textRect.left >= labelRect.left - 0.5 && textRect.right <= labelRect.right + 0.5,
			chevronVisible: chevronRect.width === 16 && chevronRect.height === 16,
			chevronWithinTrigger: chevronRect.left >= triggerRect.left - 0.5 && chevronRect.right <= triggerRect.right + 0.5 && chevronRect.top >= triggerRect.top - 0.5 && chevronRect.bottom <= triggerRect.bottom + 0.5,
			compactMarkVisible: compactRect.width === 32 && compactRect.height === 32 && getComputedStyle(compactMark).display !== 'none',
			headerHeight: header.height,
			targetsAtLeast44: targets.every(target => target.width >= 44 && target.height >= 44),
			noHorizontalOverflow: root.scrollWidth <= innerWidth && body.scrollWidth <= innerWidth,
		};
	}`)
	if err != nil {
		t.Fatal(err)
	}
	metrics := result.(map[string]any)
	if metrics["rootFontSize"] != "32px" || metrics["label"] != "Components" || metrics["labelVisible"] != true || metrics["labelNotClipped"] != true || metrics["textWithinLabel"] != true || metrics["chevronVisible"] != true || metrics["chevronWithinTrigger"] != true || metrics["compactMarkVisible"] != true || metricNumber(metrics["headerHeight"]) != 64 || metrics["targetsAtLeast44"] != true || metrics["noHorizontalOverflow"] != true {
		failWithMetrics(t, "maximum family text reflow", metrics)
	}
}

func TestFamilyNavigationPreservesLegacySmallBranding(t *testing.T) {
	requireE2E(t)
	harness := newBrowserHarness(t)
	for _, path := range []string{"/__e2e/legacy-managed", "/__e2e/legacy-custom"} {
		t.Run(path, func(t *testing.T) {
			page := harness.newPage(t, true)
			if err := page.SetViewportSize(390, 900); err != nil {
				t.Fatal(err)
			}
			response, err := page.Goto(harness.baseURL+path, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
			if err != nil {
				t.Fatal(err)
			}
			if response == nil || response.Status() != 200 {
				t.Fatalf("GET %s status = %v, want 200", path, response)
			}
			result, err := page.Evaluate(`() => {
				const root = document.documentElement;
				const source = document.querySelector('.component-doc-shell__brand-logo-source');
				const logo = document.querySelector('.component-doc-shell__managed-logo, .legacy-custom-logo');
				const compact = document.querySelector('.component-doc-shell__brand-compact-mark');
				const sourceStyle = getComputedStyle(source);
				const compactStyle = getComputedStyle(compact);
				const logoRect = logo.getBoundingClientRect();
				const compactRect = compact.getBoundingClientRect();
				return {
					familyNavigation: document.body.dataset.familyNavigation,
					sourceDisplay: sourceStyle.display,
					sourceOpacity: sourceStyle.opacity,
					logoVisible: logoRect.width > 0 && logoRect.height > 0 && sourceStyle.opacity !== '0' && sourceStyle.visibility !== 'hidden',
					compactDisplay: compactStyle.display,
					compactVisible: compactRect.width > 0 && compactRect.height > 0 && compactStyle.display !== 'none',
					brandLabel: document.querySelector('.component-doc-shell__brand').getAttribute('aria-label'),
				};
			}`)
			if err != nil {
				t.Fatal(err)
			}
			metrics := result.(map[string]any)
			if metrics["familyNavigation"] != "false" || metrics["logoVisible"] != true || metrics["compactVisible"] != false || metrics["brandLabel"] == "" {
				failWithMetrics(t, "legacy small branding", metrics)
			}
		})
	}
}

func TestFamilyNavigationWithoutJavaScript(t *testing.T) {
	requireE2E(t)
	harness := newBrowserHarness(t)
	page := harness.newPage(t, false)
	if err := page.SetViewportSize(390, 900); err != nil {
		t.Fatal(err)
	}
	gotoFamilyPage(t, page, harness.baseURL, "/components")

	failures := watchBrowserFailures(page)
	fallback := page.Locator(".component-doc-shell__family-menu-links")
	hrefs, err := fallback.Locator("a[href]").EvaluateAll(`elements => elements.map(element => element.getAttribute('href'))`)
	if err != nil {
		t.Fatal(err)
	}
	wantHrefs := []any{"/components", "/charts", "/app-shells", "/icons", "/llms", "/examples"}
	gotHrefs, ok := hrefs.([]any)
	if !ok || fmt.Sprint(gotHrefs) != fmt.Sprint(wantHrefs) {
		t.Fatalf("no-JavaScript family hrefs = %#v, want %#v", hrefs, wantHrefs)
	}
	selectHidden, err := page.Evaluate(`() => {
		const select = document.querySelector('.component-doc-shell__family-select');
		return Boolean(select && getComputedStyle(select).display === 'none');
	}`)
	if err != nil || selectHidden != true {
		t.Fatalf("no-JavaScript Goshtoso select hidden = %#v, err=%v", selectHidden, err)
	}
	if err := fallback.Locator(`a[href="/charts"]`).Click(); err != nil {
		t.Fatal(err)
	}
	if err := page.WaitForURL("**/charts"); err != nil {
		t.Fatal(err)
	}
	identity, err := page.Evaluate(`() => ({
		title: document.title,
		heading: document.querySelector('#main-content h1')?.textContent.trim(),
		scope: document.querySelector('.component-doc-shell__scope-family')?.textContent.trim(),
		family: document.querySelector('.component-doc-shell__family-menu-links [aria-current="location"]')?.textContent.trim(),
		path: location.pathname,
		fullDocument: document.doctype?.name === 'html' && Boolean(document.querySelector('html > head')) && Boolean(document.querySelector('html > body .component-doc-shell__header')),
	})`)
	if err != nil {
		t.Fatal(err)
	}
	got := identity.(map[string]any)
	if got["title"] != "Charts Documentation - Goshtoso" || got["heading"] != "Charts" || got["scope"] != "Charts" || got["family"] != "Charts" || got["path"] != "/charts" || got["fullDocument"] != true {
		failWithMetrics(t, "no-JavaScript full-document identity", got)
	}
	assertNoBrowserFailures(t, page, failures, "no-JavaScript navigation")
}
