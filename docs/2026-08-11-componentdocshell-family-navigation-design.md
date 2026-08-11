# Component documentation family navigation design

Status: approved in design discussion and written-spec review on 2026-08-11.

## Context

The Goshtoso documentation sidebar currently carries introductory guides, the
component catalog, module links, examples, and legal pages in one continuous
navigation surface. That model is already dense and will become harder to scan
when Goshtoso Charts and Goshtoso App Shells documentation join the site.

The shell needs a global product-family layer above a local documentation layer.
The global layer answers “which Goshtoso family am I browsing?” The sidebar then
answers “where am I inside that family?”

## Goals

- Provide six stable families in this order: Components, Charts, App Shells,
  Icons, LLMs, and Examples.
- Keep every family visible when horizontal space permits.
- Move family navigation to a second header row before it can overflow.
- Use a compact family dropdown on small screens while retaining one local
  sidebar trigger.
- Scope sidebar content and package metadata to the active family.
- Preserve server rendering, ordinary links, HTMX navigation, themes, keyboard
  access, and current component documentation behavior.
- Keep legal and attribution links globally reachable without treating them as
  product families.

## Non-goals

- This change does not migrate every Charts or App Shells documentation page.
- It does not copy producer-owned pages into the Goshtoso repository.
- It does not define the future cross-repository documentation contribution
  format.
- It does not remove `BrandBadge` from the public shell API. Goshtoso stops using
  a global release badge because its families release independently.
- It does not authorize merge, tag, release, deployment, worktree removal, or
  other lifecycle actions.

## Information architecture

Family overview routes are:

| Family | Route | Initial scope |
| --- | --- | --- |
| Components | `/components` | Core Goshtoso setup, model, patterns, themes, and component catalog |
| Charts | `/charts` | Existing Charts module overview; full catalog follows separately |
| App Shells | `/app-shells` | Existing App Shells module overview; full package docs follow separately |
| Icons | `/icons` | Icon component, icon packs, catalogs, and sources |
| LLMs | `/llms` | Agent skill, integration guidance, and future agent tooling |
| Examples | `/examples` | Complete applications spanning one or more families |

Switching family always opens that family’s overview. The site does not attempt
to map the current local page to an analogous page in another family.

The active family determines the sidebar items and sections. The sidebar begins
with active-family context: family label, optional Go module path, and optional
independent release version linking to that release. Families without one Go
module or version, such as Examples, omit those fields.

Search remains application-owned through `Navigation.SearchSlot`. Goshtoso keeps
one global search index and groups results by family. Every result identifies its
family so global discovery does not weaken local orientation.

The site footer owns Attributions, License, Privacy, organization, and repository
links. Source-specific chart and icon provenance remains beside the relevant
content; the global Attributions page aggregates it.

## Public componentdocshell API

`componentdocshell` owns the header, responsive family presentation, mobile
drawer, sidebar frame, and HTMX shell lifecycle. It must not embed or compose the
general Goshtoso `Navbar`, which would introduce another responsive menu and a
second mobile navigation trigger.

The public model adds these concepts:

```go
type FamilyLink struct {
	ID        string
	Label     string
	Href      string
	LinkAttrs templ.Attributes
}

type ScopeMetadata struct {
	ModulePath string
	Version    string
	VersionURL string
}
```

`Navigation` gains `Families []FamilyLink` and `Scope *ScopeMetadata`. `Page`
gains `ActiveFamily string`. Existing `Navigation.Items`, `Sections`, and
`SearchSlot` remain the active family’s local navigation.

An empty `Families` slice preserves existing consumer behavior: no family
navigation, no family dropdown, and no new active-family requirement. This keeps
the API additive in behavior and for zero-value/keyed literals. Adding exported
struct fields is not source-compatible with external positional literals, so all
supported examples use keyed literals.

When families are configured, validation requires:

- non-empty, unique family IDs;
- family IDs with no leading or trailing whitespace;
- non-empty labels and overview URLs, with URLs restricted to root-relative or
  absolute HTTPS values;
- `FamilyLink.LinkAttrs` with no `id` key, case-insensitively, because one link
  renders on both desktop and mobile surfaces;
- `Page.ActiveFamily` matching one configured ID;
- no `ScopeMetadata.VersionURL` without `ScopeMetadata.Version`;
- root-relative or absolute HTTPS version URLs when supplied;
- no existing sidebar navigation ID regressions.

The shell derives the active family label from `FamilyLink`; scope metadata does
not duplicate it.

Validation completes before template output begins and returns a descriptive
error without partial document or fragment bytes.

## Rendering and navigation lifecycle

Full-page rendering includes the family navigation, active scope, sidebar, main
content, optional table of contents, and footer.

Family links remain ordinary anchors. Without JavaScript they perform normal
full-page navigation. With HTMX enabled, the shell adds the same navigation
attributes it owns for sidebar links: fetch the destination, target main content,
and push the URL.

Fragment responses contain:

- the document title;
- an out-of-band main-content replacement;
- an out-of-band scoped-sidebar replacement; and
- an out-of-band family-navigation replacement.

The family replacement updates desktop active state and the mobile dropdown’s
current label together. The active family uses `aria-current="location"`; the
exact local sidebar page retains `aria-current="page"`.

After navigation settles, the shell closes the mobile sidebar and family menu,
updates history, resets the main scroll position, and moves focus using the
existing shell focus contract. Back and Forward must restore title, family,
sidebar, URL, and content as one identity.

The mobile family menu uses semantic link markup with a no-JavaScript disclosure
fallback. Enhancement adds outside-click and Escape closure plus focus return;
it does not replace link semantics.

## Responsive layout

Responsive states are deterministic CSS layouts, not JavaScript measurements:

- **Small, below 720px:** one 64px row containing local-sidebar trigger,
  Goshtoso mark, current-family dropdown, and dark-mode control. The built-in
  theme selector and repository link move to a drawer utility region.
- **Medium, 720px through 1199px:** 64px brand/control row plus a 44px family
  navigation row. Persistent sidebar begins below the combined 108px header.
- **Wide, 1200px and above:** one 64px row containing brand, inline family
  navigation, and controls.

The shell exposes its current header height through an internal CSS custom
property used consistently by frame height, fixed mobile sidebar, backdrop, and
scroll offsets. No element may retain a hard-coded 64px offset while the medium
header is 108px.

All configured controls remain reachable at every width. Responsive behavior may
relocate low-priority controls, but must not hide a configured destination
without another visible, labelled access path.

The built-in theme selector may render desktop and mobile instances with unique
DOM IDs, one visible at each breakpoint, both bound to the same theme state.
Repository links follow the same visibility rule. The shell does not clone the
arbitrary `HeaderActions` slot because consumer components may contain IDs or
state; consumers keep responsibility for making that slot responsive.

## Goshtoso site adoption

Goshtoso supplies the six family links, active family for every route, local
sidebar model, active scope metadata, global search index, and structured footer.
Its current top-level sidebar guides move under Components, Icons, or LLMs.
Attributions and License leave the sidebar.

The current header-wide release badge is removed from Goshtoso configuration.
Components, Charts, and App Shells display their own package/version context in
the scoped sidebar or page metadata.

Existing Charts and App Shells module guides become the first `/charts` and
`/app-shells` overview pages. Full producer catalogs require a separate design.
That future work must keep producer repositories authoritative, consume a
versioned public contribution contract, and avoid importing any `site/internal`
package or copying pages that can drift.

## Acceptance contract

### App Shell package

- Existing consumers render unchanged when `Families` is empty.
- Config validation covers empty, duplicate, and unknown family IDs plus invalid
  scope metadata.
- Full documents render all configured family links exactly once per visible
  navigation surface with correct semantics.
- Fragments contain the title and exactly one OOB replacement for main, sidebar,
  and family navigation.
- HTMX attributes preserve consumer-provided link attributes without mutating
  caller-owned slices or maps.

### Goshtoso consumer

- Direct loads and JavaScript-disabled navigation reach every family root and
  every local destination.
- HTMX navigation keeps title, URL, active family, active sidebar item, content,
  and focus synchronized.
- Back and Forward restore the same synchronized identity.
- Global search results identify family and navigate to correct scoped sidebar.
- Footer links expose Attributions, License, Privacy, organization, and source
  repositories.

### Visual and accessibility matrix

- Exercise 390px, 719px, 720px, 841px, 1199px, 1200px, 1280px, and 1440px.
- Assert no horizontal overflow, clipped family labels, unreachable controls, or
  sidebar/backdrop gaps at either side of each breakpoint.
- Verify Arai Hû, Goshtoso, and Minimal themes in light and dark mode.
- Test keyboard order, visible focus, family menu open/close, Escape, focus
  return, local drawer closure, and current-state semantics.
- Test system light/dark preferences and throwing browser storage; navigation and
  visible controls must continue working without page errors.
- When the theme selector is enabled, assert desktop/mobile instances never show
  together and remain synchronized through the shared theme state.
- Check browser console and run the project accessibility scan.

## Delivery sequence

1. Start App Shell work from fetched, verified `origin/main` in an isolated
   worktree. Implement API, rendering, CSS, unit tests, and example coverage.
2. Prove the App Shell change with its own tests and a clean external consumer.
3. Review and release App Shells only after separate authorization.
4. Start Goshtoso adoption from its then-current `origin/main` in a separate
   isolated worktree. Consume the released App Shell version rather than a local
   replacement for final proof.
5. Run Goshtoso root and site checks, current-source integration, standalone
   pinned-dependency deployability with `GOWORK=off`, and browser acceptance.
6. Treat merge, tag, release, deployment, and cleanup as distinct approval gates.

## Rejected approaches

- **Keep one mixed sidebar:** does not solve density or prepare for extension
  catalogs.
- **Compose Goshtoso Navbar inside componentdocshell:** duplicates shell and
  mobile-menu ownership.
- **Always use two header rows:** wastes reading height on wide screens.
- **Use only a product dropdown:** hides ecosystem breadth on screens that can
  show it.
- **Use horizontal mobile overflow:** makes unequal family labels partially
  hidden and swipe-dependent.
- **Copy Charts/App Shells pages into Goshtoso:** creates immediate source drift
  and unclear release ownership.
