package componentdocshell

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func renderLayout(t *testing.T, cfg Config, page Page) string {
	t.Helper()
	var buffer bytes.Buffer
	if err := Layout(cfg, page).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	return buffer.String()
}

func renderFragment(t *testing.T, cfg Config, page Page) string {
	t.Helper()
	var buffer bytes.Buffer
	if err := Fragment(cfg, page).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Fragment().Render() error = %v", err)
	}
	return buffer.String()
}

func TestLayoutRendersFamilyNavigationAndScope(t *testing.T) {
	t.Parallel()
	cfg, page := validFamilyConfig(), validFamilyPage()
	body := renderLayout(t, cfg, page)
	for _, want := range []string{
		`id="componentdocshell-family-navigation"`,
		`aria-label="Documentation families"`,
		`component-doc-shell__family-select`,
		`id="componentdocshell-family-select-control"`,
		`hx-preserve`,
		`id="componentdocshell-family-trigger"`,
		`role="combobox"`,
		`x-data="goshtosoSelect($el)"`,
		`aria-current="location"`,
		`github.com/araihu/goshtoso`,
		`/assets/icons/heroicons.svg#hi-16-solid-arrow-top-right-on-square`,
		`component-doc-shell__scope-external-icon`,
		`size-4 text-primary dark:text-primary-dark`,
		`href="https://github.com/araihu/goshtoso/releases/tag/v0.1.6"`,
		`v0.1.6`,
		`<div class="component-doc-shell__mobile-utilities">`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("layout missing %q", want)
		}
	}
	if strings.Contains(body, `<details class="component-doc-shell__family-menu"`) {
		t.Fatal("family navigation still uses the custom details disclosure")
	}
	if strings.Contains(body, `class="component-doc-shell__mobile-utilities" aria-label=`) {
		t.Fatal("mobile utilities retains a semantic label")
	}
	if strings.Contains(body, `component-doc-shell__mobile-repository`) || strings.Contains(body, `>Source repository</a>`) {
		t.Fatal("mobile drawer still renders the repository link")
	}
	if got := strings.Count(body, `href="/components"`); got != 2 {
		t.Fatalf("Components family link count = %d, want one desktop and one mobile link", got)
	}
	if got := strings.Count(body, `aria-current="location"`); got != 2 {
		t.Fatalf("active family marker count = %d, want 2", got)
	}
}

func TestLayoutPreservesThemeSelectorRoots(t *testing.T) {
	t.Parallel()
	cfg, page := validFamilyConfig(), validFamilyPage()
	cfg.Appearance.ThemeSelectorID = "docs-theme"
	body := renderLayout(t, cfg, page)
	if got := strings.Count(body, `hx-preserve="true"`); got != 2 {
		t.Fatalf("preserved theme roots = %d, want 2", got)
	}
	for _, id := range []string{"docs-theme-root", "docs-theme-mobile-root"} {
		if got := strings.Count(body, `id="`+id+`"`); got != 1 {
			t.Errorf("preserved theme root %q count = %d, want 1", id, got)
		}
	}
	if !strings.Contains(body, `class="component-doc-shell__mobile-theme"`) {
		t.Fatal("mobile theme selector missing dedicated preserve wrapper")
	}

	cfg.Appearance.DisableThemeSelector = true
	body = renderLayout(t, cfg, page)
	if got := strings.Count(body, `hx-preserve="true"`); got != 0 {
		t.Fatalf("disabled theme selector preserved roots = %d, want 0", got)
	}

	cfg = validConfig()
	cfg.Appearance.ThemeSelectorID = "docs-theme"
	body = renderLayout(t, cfg, validPage())
	if got := strings.Count(body, `hx-preserve="true"`); got != 1 {
		t.Fatalf("family-free preserved theme roots = %d, want desktop only", got)
	}
	if strings.Contains(body, `id="docs-theme-mobile-root"`) {
		t.Fatal("family-free layout rendered mobile theme preserve root")
	}
}

func TestLayoutRendersFamilyCallerAttributesWithoutMutation(t *testing.T) {
	t.Parallel()
	cfg, page := validFamilyConfig(), validFamilyPage()
	cfg.Navigation.Families[0].LinkAttrs = templ.Attributes{"data-analytics": "components-family"}
	body := renderLayout(t, cfg, page)
	if got := strings.Count(body, `data-analytics="components-family"`); got != 2 {
		t.Fatalf("family caller attribute count = %d, want 2", got)
	}
	if got := cfg.Navigation.Families[0].LinkAttrs["data-analytics"]; got != "components-family" {
		t.Fatalf("layout mutated family caller attribute = %#v", got)
	}
}

func TestLayoutGuardsSidebarEscapeHandlerWithOpenState(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("layout.templ")
	if err != nil {
		t.Fatalf("ReadFile(layout.templ) error = %v", err)
	}
	want := `x-on:keydown.escape.window="if (sidebarOpen) { sidebarOpen = false; $nextTick(() => $refs.sidebarTrigger.focus()) }"`
	if !strings.Contains(string(source), want) {
		t.Fatalf("layout source missing guarded sidebar Escape handler %q", want)
	}
	if !strings.Contains(renderLayout(t, validFamilyConfig(), validFamilyPage()), `x-on:keydown.escape.window="if (sidebarOpen) { sidebarOpen = false;`) {
		t.Fatal("rendered layout missing guarded sidebar Escape handler")
	}
}

func TestLayoutClosesFamilySelectBeforeOpeningMobileDrawer(t *testing.T) {
	t.Parallel()
	body := renderLayout(t, validFamilyConfig(), validFamilyPage())
	want := `x-on:click="$dispatch('componentdocshell:close-family-select'); sidebarOpen = !sidebarOpen"`
	if !strings.Contains(body, want) {
		t.Fatalf("mobile drawer trigger missing family-select exclusion %q", want)
	}
}

func TestLayoutBindsSidebarInertToResponsiveState(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("layout.templ")
	if err != nil {
		t.Fatalf("ReadFile(layout.templ) error = %v", err)
	}
	wantSource := `x-bind:inert="!sidebarOpen && !sidebarPersistent"`
	if !strings.Contains(string(source), wantSource) {
		t.Fatalf("layout source missing responsive sidebar inert binding %q", wantSource)
	}
	body := renderLayout(t, validFamilyConfig(), validFamilyPage())
	if !strings.Contains(body, `x-bind:inert="!sidebarOpen && !sidebarPersistent"`) {
		t.Fatal("rendered layout missing responsive sidebar inert binding")
	}
}

func TestLayoutDelegatesDrawerContainmentToRuntime(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("layout.templ")
	if err != nil {
		t.Fatalf("ReadFile(layout.templ) error = %v", err)
	}
	if strings.Contains(string(source), `x-trap`) {
		t.Fatal("layout source must not install Alpine Focus global trap")
	}
	if strings.Contains(renderLayout(t, validFamilyConfig(), validFamilyPage()), `x-trap`) {
		t.Fatal("rendered layout must not install Alpine Focus global trap")
	}
}

func TestLayoutOmitsFamilySurfacesWhenFamiliesEmpty(t *testing.T) {
	t.Parallel()
	body := renderLayout(t, validConfig(), validPage())
	for _, absent := range []string{
		`id="componentdocshell-family-navigation"`,
		`component-doc-shell__family-menu`,
		`component-doc-shell__scope`,
		`component-doc-shell__mobile-utilities`,
	} {
		if strings.Contains(body, absent) {
			t.Errorf("legacy layout unexpectedly contains %q", absent)
		}
	}
}

func TestLayoutRendersLinkedScopeMetadataAsOneRow(t *testing.T) {
	t.Parallel()
	body := renderLayout(t, validFamilyConfig(), validFamilyPage())
	for _, want := range []string{
		`class="component-doc-shell__scope-details"`,
		`class="component-doc-shell__scope-module" href="https://github.com/araihu/goshtoso"`,
		`<code>araihu/goshtoso</code>`,
		`class="component-doc-shell__scope-version-badge">v0.1.6</span>`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("linked scope metadata missing %q", want)
		}
	}
}

func TestFragmentRendersAtomicFamilyIdentity(t *testing.T) {
	t.Parallel()
	cfg, page := validFamilyConfig(), validFamilyPage()
	cfg.Interactions.EnableHTMX = true
	body := renderFragment(t, cfg, page)
	for _, target := range []string{
		`outerHTML:#main-content`,
		`outerHTML:#componentdocshell-sidebar-content`,
		`outerHTML:#componentdocshell-family-navigation`,
	} {
		if got := strings.Count(body, target); got != 1 {
			t.Errorf("fragment target %q count = %d, want 1", target, got)
		}
	}
	if !strings.Contains(body, `<title>Line · Reference</title>`) {
		t.Fatal("fragment missing title")
	}
}

func TestLayoutRendersComponentDocsShellContract(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.RepositoryURL = "https://github.com/araihu/reference"
	cfg.Interactions.EnableHTMX = true
	page := validPage()
	page.Description = "Reference components"
	page.Content = templ.Raw(`<h1 id="line" data-toc-heading>Line</h1><h2 id="usage" data-toc-heading>Usage</h2>`)
	page.EnableTOC = true

	var buffer bytes.Buffer
	if err := Layout(cfg, page).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	body := buffer.String()
	for _, want := range []string{
		"<!doctype html>",
		`<html lang="en" class="component-doc-shell-root"`,
		`href="#main-content"`,
		`class="component-doc-shell__header"`,
		`aria-label="Open navigation"`,
		`aria-controls="componentdocshell-sidebar"`,
		`x-bind:aria-expanded="sidebarOpen"`,
		`aria-label="Theme"`,
		`aria-label="Switch to dark mode"`,
		`Switch to light mode`,
		`x-on:componentdocshell:navigated.window="sidebarOpen = false"`,
		`aria-current="page"`,
		`id="componentdocshell-sidebar-content"`,
		`id="main-content"`,
		`id="componentdocshell-toc"`,
		`/componentdocshell/assets/shell.css`,
		`/componentdocshell/assets/shell.js`,
		`localStorage.getItem("theme")`,
		`github.com/araihu/reference`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("layout missing %q", want)
		}
	}
	if shellIndex := strings.Index(body, `/componentdocshell/assets/shell.js`); shellIndex < 0 || shellIndex > strings.Index(body, `/assets/js/dependency-loader.js`) {
		t.Fatal("shell registration script must run before the Goshtoso dependency loader")
	}
}

func TestLayoutBootstrapsPersistedAppearanceBeforeRuntime(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Appearance.PersistPreferences = true
	cfg.Appearance.DefaultTheme = "minimal"
	cfg.Appearance.InitialColorScheme = ColorSchemeDark
	var buffer bytes.Buffer
	if err := Layout(cfg, validPage()).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	body := buffer.String()
	for _, want := range []string{`"persist":true`, `"persistTheme":true`, `"theme":"minimal"`, `"colorScheme":"dark"`, `document.documentElement.setAttribute("data-theme",theme)`} {
		if !strings.Contains(body, want) {
			t.Errorf("layout appearance bootstrap missing %q", want)
		}
	}
}

func TestLayoutRendersManagedPresentationChannel(t *testing.T) {
	t.Parallel()
	zeroValueHTML := renderValid(t, validConfig())
	for _, hook := range []string{"data-asset-brand", "data-campaign-toggle", "data-use-campaign-label", "data-channel"} {
		if strings.Contains(zeroValueHTML, hook) {
			t.Fatalf("zero-value configuration rendered %q", hook)
		}
	}

	cfg := validConfig()
	cfg.Brand.ManagedLogo = &ManagedBrandAsset{URL: "/assets/brand/logo.svg", Alt: "Reference", Width: 120, Height: 32}
	cfg.Brand.ManageFavicon = true
	cfg.Brand.FaviconURL = "/assets/brand/icon.svg"
	cfg.Interactions.PresentationChannel = &PresentationChannelConfig{
		RuntimeURL:       "/assets/campaign/v1.js",
		ChannelURL:       "/assets/releases/current",
		Integrity:        "sha384-campaign",
		UseCampaignLabel: "Use campaign",
		UseBaselineLabel: "Use baseline",
	}
	html := renderValid(t, cfg)
	for _, want := range []string{
		`data-asset-brand="logo"`, `width="120"`, `height="32"`,
		`data-asset-brand="icon"`, `data-campaign-toggle`, `data-campaign-toggle-icon`,
		`data-use-campaign-label="Use campaign"`, `data-use-baseline-label="Use baseline"`,
		`data-channel="/assets/releases/current"`, `integrity="sha384-campaign"`, `crossorigin="anonymous"`, `defer`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("layout missing %q", want)
		}
	}
	if strings.Count(html, "data-campaign-toggle") != 2 {
		t.Fatalf("campaign toggle hooks = %d, want button and icon only", strings.Count(html, "data-campaign-toggle"))
	}
	for _, want := range []string{`type="button"`, `hidden`, `aria-pressed="false"`, `class="sr-only"`} {
		if !strings.Contains(html, want) {
			t.Errorf("campaign toggle missing %q", want)
		}
	}
	if bootstrap, runtime := strings.Index(html, "dataset.themeSource"), strings.Index(html, `src="/assets/campaign/v1.js"`); bootstrap < 0 || runtime < 0 || bootstrap > runtime {
		t.Fatal("first-paint bootstrap must precede campaign runtime")
	}
	if stylesheet, documentEnd := strings.Index(html, `/componentdocshell/assets/shell.css`), strings.Index(html, "</html>"); stylesheet < 0 || documentEnd < 0 || stylesheet > documentEnd {
		t.Fatal("baseline stylesheet must precede document end")
	}
	if !strings.Contains(html, `class="component-doc-shell__brand-name">Reference</span>`) {
		t.Fatal("managed logo must preserve brand name by default")
	}

	cfg.Brand.HideName = true
	html = renderValid(t, cfg)
	if strings.Contains(html, `class="component-doc-shell__brand-name">Reference</span>`) {
		t.Fatal("hidden managed brand name rendered")
	}
}

func TestLayoutProvidesCompactBrandFallbackForLogoBackedBrands(t *testing.T) {
	t.Parallel()
	cfg := validFamilyConfig()
	cfg.Brand.ManagedLogo = &ManagedBrandAsset{URL: "/assets/brand/logo.svg", Alt: "Reference", Width: 120, Height: 32}
	body := renderLayout(t, cfg, validFamilyPage())
	for _, want := range []string{
		`href="/" aria-label="Reference home"`,
		`class="component-doc-shell__brand-logo-source" aria-hidden="true"`,
		`class="component-doc-shell__managed-logo"`,
		`class="component-doc-shell__brand-compact-mark" aria-hidden="true">R</span>`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("managed logo layout missing compact-brand contract %q", want)
		}
	}

	cfg = validConfig()
	cfg.Brand.Logo = templ.Raw(`<img src="/wordmark.svg" alt="Reference">`)
	body = renderLayout(t, cfg, validPage())
	if !strings.Contains(body, `class="component-doc-shell__brand-compact-mark" aria-hidden="true">R</span>`) {
		t.Fatal("custom logo layout missing compact initial mark")
	}
}

func TestLayoutUsesOptionalCompactBrandAsset(t *testing.T) {
	t.Parallel()
	cfg := validFamilyConfig()
	cfg.Brand.Logo = templ.Raw(`<svg class="wordmark" viewBox="0 0 120 32" aria-label="Reference"></svg>`)
	cfg.Brand.CompactLogo = templ.Raw(`<svg class="compact-logo" viewBox="0 0 32 32" aria-label="Reference mark"></svg>`)
	body := renderLayout(t, cfg, validFamilyPage())
	if !strings.Contains(body, `class="component-doc-shell__brand-compact-mark" aria-hidden="true"><svg class="compact-logo"`) {
		t.Fatal("layout did not render optional compact brand asset")
	}
	if strings.Contains(body, `class="component-doc-shell__brand-compact-mark" aria-hidden="true">R</span>`) {
		t.Fatal("layout rendered first-rune fallback alongside configured compact asset")
	}
}

func TestLayoutUsesCompactBrandAssetWithoutFullLogo(t *testing.T) {
	t.Parallel()
	cfg := validFamilyConfig()
	cfg.Brand.CompactLogo = templ.Raw(`<svg id="compact-only-asset" class="compact-only-logo" viewBox="0 0 32 32" aria-label="Reference mark"></svg>`)
	body := renderLayout(t, cfg, validFamilyPage())
	if strings.Contains(body, `class="component-doc-shell__brand-mark"`) {
		t.Fatal("compact-only brand rendered first-rune mark alongside compact asset")
	}
	if !strings.Contains(body, `class="component-doc-shell__brand component-doc-shell__brand--compact-only"`) {
		t.Fatal("compact-only brand missing responsive modifier")
	}
	if got := strings.Count(body, `id="compact-only-asset"`); got != 1 {
		t.Fatalf("compact-only asset render count = %d, want one unique DOM instance", got)
	}
}

func TestLayoutKeepsLegacyBrandMarkupWithoutFamilies(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Brand.ManagedLogo = &ManagedBrandAsset{URL: "/assets/brand/logo.svg", Alt: "Reference", Width: 120, Height: 32}
	body := renderLayout(t, cfg, validPage())
	for _, want := range []string{
		`data-family-navigation="false"`,
		`class="component-doc-shell__brand-logo-source" aria-hidden="true"`,
		`class="component-doc-shell__managed-logo"`,
		`class="component-doc-shell__brand-compact-mark" aria-hidden="true">R</span>`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("legacy brand layout missing %q", want)
		}
	}
}

func TestLayoutRendersStableSidebarNavHookForDefaultSearch(t *testing.T) {
	t.Parallel()
	body := renderLayout(t, validConfig(), validPage())
	for _, want := range []string{
		`component-doc-shell__sidebar-nav`,
		`class="component-doc-shell__default-search shrink-0"`,
		`class="flex-1 overflow-y-auto sidebar-scroll`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("default sidebar missing stable gutter hook %q", want)
		}
	}
}

func TestLayoutLeavesCustomSearchSlotUnwrapped(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Navigation.SearchSlot = templ.Raw(`<section class="consumer-search-slot"><input aria-label="Consumer search"></section>`)
	body := renderLayout(t, cfg, validPage())
	if !strings.Contains(body, `<section class="consumer-search-slot"><input aria-label="Consumer search"></section>`) {
		t.Fatal("layout did not preserve consumer search slot root")
	}
	if strings.Contains(body, `component-doc-shell__default-search`) {
		t.Fatal("layout wrapped consumer search slot with default-search styling hook")
	}
}

func TestAppearanceBootstrapMarksThemeSource(t *testing.T) {
	t.Parallel()
	disabled := appearanceBootstrapScript(validConfig())
	if !strings.Contains(disabled, `"persist":false`) || !strings.Contains(disabled, `"persistTheme":false`) {
		t.Error("disabled persistence bootstrap enables a saved theme")
	}
	cfg := validConfig()
	cfg.Appearance.PersistPreferences = true
	script := appearanceBootstrapScript(cfg)
	for _, want := range []string{
		`var source="default"`,
		`var savedTheme=localStorage.getItem("theme");if(savedTheme){theme=savedTheme;source="preference"}`, // missing or empty key stays default; non-empty key is preference
		`catch(_){}`, // storage exceptions retain configured theme and default source
		`var saved=localStorage.getItem("darkMode");if(saved!==null)dark=saved==="true"`,
		`document.documentElement.dataset.themeSource=source`,
	} {
		if !strings.Contains(script, want) {
			t.Errorf("appearance bootstrap missing %q", want)
		}
	}
}

func TestLayoutLocksThemePersistenceWhenSelectorIsDisabled(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Appearance.PersistPreferences = true
	cfg.Appearance.DefaultTheme = "araihu"
	cfg.Appearance.DisableThemeSelector = true
	var buffer bytes.Buffer
	if err := Layout(cfg, validPage()).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	body := buffer.String()
	for _, want := range []string{`"persist":true`, `"persistTheme":false`, `"theme":"araihu"`} {
		if !strings.Contains(body, want) {
			t.Errorf("locked theme layout missing %q", want)
		}
	}
	if strings.Contains(body, `aria-label="Theme"`) {
		t.Error("locked theme layout rendered theme selector")
	}
}

func TestLayoutCanBindDarkModeControlToApplicationStore(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Appearance.DarkModeBinding = &DarkModeBinding{
		ButtonID:         "darkModeToggleBtn",
		StateExpression:  "$store.darkMode.on",
		ToggleExpression: "$store.darkMode.toggle()",
	}
	var buffer bytes.Buffer
	if err := Layout(cfg, validPage()).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	body := buffer.String()
	for _, want := range []string{
		`id="darkModeToggleBtn"`,
		`x-bind:aria-label="$store.darkMode.on ? &#39;Switch to light mode&#39; : &#39;Switch to dark mode&#39;"`,
		`x-on:click="$store.darkMode.toggle()"`,
		`x-show="!($store.darkMode.on)"`,
		`x-show="$store.darkMode.on"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("layout application dark-mode binding missing %q", want)
		}
	}
}

func TestLayoutCanUseWordmarkWithoutDuplicateBrandName(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Brand.Logo = templ.Raw(`<img src="/wordmark.svg" alt="">`)
	cfg.Brand.HideName = true
	var buffer bytes.Buffer
	if err := Layout(cfg, validPage()).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	body := buffer.String()
	if !strings.Contains(body, `/wordmark.svg`) {
		t.Error("wordmark layout missing logo")
	}
	if strings.Contains(body, `component-doc-shell__brand-name`) {
		t.Error("wordmark layout rendered duplicate brand name")
	}
}

func TestLayoutBrandBadgeIsOptional(t *testing.T) {
	t.Parallel()
	var buffer bytes.Buffer
	if err := Layout(validConfig(), validPage()).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	if strings.Contains(buffer.String(), `component-doc-shell__brand-badge`) {
		t.Fatal("layout rendered an unconfigured brand badge")
	}
}

func TestLayoutRendersLinkedBrandBadge(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Brand.Badge = &BrandBadge{
		Label:     "v1.2.3",
		AriaLabel: "Goshtoso release v1.2.3",
		Href:      "https://github.com/araihu/goshtoso/releases/tag/v1.2.3",
	}
	var buffer bytes.Buffer
	if err := Layout(cfg, validPage()).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	body := buffer.String()
	for _, want := range []string{
		`class="component-doc-shell__brand-badge"`,
		`href="https://github.com/araihu/goshtoso/releases/tag/v1.2.3"`,
		`aria-label="Goshtoso release v1.2.3"`,
		`>v1.2.3</a>`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("linked brand badge missing %q", want)
		}
	}
}

func TestLayoutRendersUnlinkedBrandBadge(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Brand.Badge = &BrandBadge{Label: "dev", AriaLabel: "Development build"}
	var buffer bytes.Buffer
	if err := Layout(cfg, validPage()).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	body := buffer.String()
	if !strings.Contains(body, `<span class="component-doc-shell__brand-badge" aria-label="Development build">dev</span>`) {
		t.Fatalf("unlinked brand badge markup missing: %s", body)
	}
	if strings.Contains(body, `href=""`) {
		t.Fatal("unlinked brand badge rendered an empty link")
	}
}

func TestLayoutCanPreserveApplicationThemeSelectorID(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Appearance.ThemeSelectorID = "site-theme"
	var buffer bytes.Buffer
	if err := Layout(cfg, validPage()).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	body := buffer.String()
	for _, want := range []string{`id="site-theme-trigger"`, `id="site-theme-listbox"`, `name="theme"`} {
		if !strings.Contains(body, want) {
			t.Errorf("layout application theme-selector ID missing %q", want)
		}
	}
}

func TestLayoutCanPreserveApplicationTOCIDs(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.TOC = TOCConfig{RailID: "toc-rail", ListID: "toc-list"}
	page := validPage()
	page.EnableTOC = true
	var buffer bytes.Buffer
	if err := Layout(cfg, page).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	body := buffer.String()
	for _, want := range []string{
		`id="toc-rail" class="component-doc-shell__toc" data-componentdocshell-toc`,
		`id="toc-list" class="component-doc-shell__toc-list" data-componentdocshell-toc-list`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("layout application TOC ID missing %q", want)
		}
	}
}

func TestLayoutCanExposeLocalHTMXBeforeBodyContent(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Interactions.LocalRuntime = true
	var buffer bytes.Buffer
	if err := Layout(cfg, validPage()).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	body := buffer.String()
	if strings.Contains(body, `/assets/js/dependency-loader.js`) {
		t.Fatal("local runtime layout contains dependency loader")
	}
	htmlHTMX := `<script src="/assets/js/runtime/htmx.org/`
	if !strings.Contains(body, htmlHTMX) {
		t.Fatalf("local runtime layout missing eager HTMX script %q", htmlHTMX)
	}
}

func TestLayoutRendersApplicationSearchOverlayAndRuntimeExtensions(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Navigation.SearchSlot = templ.Raw(`<button id="docs-search">Search docs</button>`)
	cfg.BodyEnd = templ.Raw(`<div id="storage-consent">Consent</div>`)
	cfg.Interactions.LocalRuntime = true
	cfg.Interactions.RuntimeScripts = []string{"/assets/js/runtime/htmx-ext-ws.js", "/assets/js/runtime/htmx-ext-sse.js"}
	var buffer bytes.Buffer
	if err := Layout(cfg, validPage()).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	body := buffer.String()
	for _, want := range []string{`id="docs-search"`, `id="storage-consent"`, `src="/assets/js/runtime/htmx-ext-ws.js"`, `src="/assets/js/runtime/htmx-ext-sse.js"`} {
		if !strings.Contains(body, want) {
			t.Errorf("layout missing application extension %q", want)
		}
	}
	if core := strings.Index(body, `/assets/js/runtime/htmx.org/`); core < 0 || core > strings.Index(body, `/assets/js/runtime/htmx-ext-ws.js`) {
		t.Fatal("HTMX core must render before runtime extensions")
	}
}

func TestLayoutDoesNotMutateNavigation(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	if cfg.Navigation.Sections[0].Items[0].Active {
		t.Fatal("test fixture unexpectedly active")
	}
	var buffer bytes.Buffer
	if err := Layout(cfg, validPage()).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	if cfg.Navigation.Sections[0].Items[0].Active {
		t.Fatal("Layout mutated caller-owned navigation")
	}
}

func TestLayoutAllowsExactDocumentTitle(t *testing.T) {
	t.Parallel()
	page := validPage()
	page.DocumentTitle = "Exact established SEO title"
	var buffer bytes.Buffer
	if err := Layout(validConfig(), page).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	if !strings.Contains(buffer.String(), "<title>Exact established SEO title</title>") {
		t.Fatalf("layout did not preserve the exact document title")
	}
}

func TestLayoutRendersCompleteSocialMetadataOnce(t *testing.T) {
	t.Parallel()
	page := validPage()
	page.DocumentTitle = "Line reference - Reference"
	page.Description = "Build accessible line charts with server-rendered Go components."
	page.CanonicalURL = "https://docs.example/components/line"
	page.SiteName = "Reference Docs"
	page.Locale = "en_US"
	page.SocialImage = SocialImage{
		URL:      "https://docs.example/assets/line-social.png",
		MIMEType: "image/png",
		Width:    1200,
		Height:   630,
		Alt:      "Line chart component preview",
	}
	body := renderLayout(t, validConfig(), page)
	for _, want := range []string{
		`<title>Line reference - Reference</title>`,
		`<meta name="description" content="Build accessible line charts with server-rendered Go components.">`,
		`<link rel="canonical" href="https://docs.example/components/line">`,
		`<meta property="og:url" content="https://docs.example/components/line">`,
		`<meta property="og:type" content="website">`,
		`<meta property="og:title" content="Line reference - Reference">`,
		`<meta property="og:description" content="Build accessible line charts with server-rendered Go components.">`,
		`<meta property="og:site_name" content="Reference Docs">`,
		`<meta property="og:locale" content="en_US">`,
		`<meta property="og:image" content="https://docs.example/assets/line-social.png">`,
		`<meta property="og:image:type" content="image/png">`,
		`<meta property="og:image:width" content="1200">`,
		`<meta property="og:image:height" content="630">`,
		`<meta property="og:image:alt" content="Line chart component preview">`,
		`<meta name="twitter:card" content="summary_large_image">`,
		`<meta name="twitter:title" content="Line reference - Reference">`,
		`<meta name="twitter:description" content="Build accessible line charts with server-rendered Go components.">`,
		`<meta name="twitter:image" content="https://docs.example/assets/line-social.png">`,
		`<meta name="twitter:image:alt" content="Line chart component preview">`,
	} {
		if got := strings.Count(body, want); got != 1 {
			t.Errorf("metadata %q count = %d, want 1", want, got)
		}
	}
}

func TestGoshtosoBrandUsesCanonicalMarkAndFavicon(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Brand = GoshtosoBrand("Goshtoso Charts", "/", "")
	var buffer bytes.Buffer
	if err := Layout(cfg, validPage()).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Layout().Render() error = %v", err)
	}
	for _, want := range []string{"goshtoso-mark.svg", "goshtoso-mark-reverse.svg", "goshtoso-favicon.svg"} {
		if !strings.Contains(buffer.String(), want) {
			t.Errorf("layout missing %q", want)
		}
	}
}

func TestFragmentRendersMainAndOutOfBandSidebar(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Interactions.EnableHTMX = true
	var buffer bytes.Buffer
	if err := Fragment(cfg, validPage()).Render(context.Background(), &buffer); err != nil {
		t.Fatalf("Fragment().Render() error = %v", err)
	}
	body := buffer.String()
	if strings.Contains(body, "<html") {
		t.Fatal("fragment contains complete document")
	}
	for _, want := range []string{`<title>Line · Reference</title>`, `id="main-content"`, `hx-swap-oob="outerHTML:#componentdocshell-sidebar-content"`, `aria-current="page"`} {
		if !strings.Contains(body, want) {
			t.Errorf("fragment missing %q", want)
		}
	}
}

func TestLayoutReportsValidationErrorAtRender(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Brand.Name = ""
	var buffer bytes.Buffer
	err := Layout(cfg, validPage()).Render(context.Background(), &buffer)
	if err == nil || !strings.Contains(err.Error(), "brand name is required") {
		t.Fatalf("Layout().Render() error = %v", err)
	}
}

func TestFamilyLinksCloneAttributesWithoutMutation(t *testing.T) {
	t.Parallel()
	cfg := validFamilyConfig()
	cfg.Interactions.EnableHTMX = true
	cfg.Navigation.Families[0].LinkAttrs = templ.Attributes{
		"data-analytics": "components-family",
		"hx-target":      "#consumer-target",
	}

	links := familyLinks(cfg, "components")
	if links[0].LinkAttrs["data-analytics"] != "components-family" {
		t.Fatal("familyLinks() dropped consumer attribute")
	}
	for key, want := range map[string]any{
		"hx-get":      "/components",
		"hx-target":   "#main-content",
		"hx-push-url": "true",
	} {
		if got := links[0].LinkAttrs[key]; got != want {
			t.Errorf("familyLinks()[0].LinkAttrs[%q] = %#v, want %#v", key, got, want)
		}
	}
	if got := cfg.Navigation.Families[0].LinkAttrs["hx-target"]; got != "#consumer-target" {
		t.Fatalf("familyLinks() mutated caller attribute: %#v", got)
	}
	if _, exists := cfg.Navigation.Families[0].LinkAttrs["hx-get"]; exists {
		t.Fatal("familyLinks() added hx-get to caller map")
	}
	if got := links[0].LinkAttrs["aria-current"]; got != "location" {
		t.Fatalf("familyLinks() active aria-current = %#v, want location", got)
	}
	if _, exists := links[1].LinkAttrs["aria-current"]; exists {
		t.Fatal("familyLinks() marked inactive family current")
	}
}

func TestFamilyLinksStayOrdinaryWithoutHTMX(t *testing.T) {
	t.Parallel()
	cfg := validFamilyConfig()
	links := familyLinks(cfg, "components")
	if _, exists := links[0].LinkAttrs["hx-get"]; exists {
		t.Fatal("familyLinks() added HTMX attributes while HTMX is disabled")
	}
}
