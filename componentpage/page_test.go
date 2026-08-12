package componentpage

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func TestPageRendersSharedComponentReferenceStructure(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	err := Page(Config{
		Title: "Line Chart", Description: "An SSR chart.",
		Primary:  Example{Preview: templ.Raw(`<svg aria-label="Line"></svg>`), Code: `@line.Line(cfg)`},
		Sections: []Example{{Title: "Empty State", Description: "No observations.", Preview: templ.Raw(`<p>None</p>`), Code: `@line.Line(empty)`}},
	}).Render(context.Background(), &output)
	if err != nil {
		t.Fatalf("Page().Render() error = %v", err)
	}
	for _, want := range []string{
		`data-component-page`, `id="line-chart"`, `data-component-description`,
		`data-component-preview`, `Usage Example`, `id="empty-state"`, `component-example-empty-state`,
		`data-component-preview class="component-page__preview rounded-radius border border-outline dark:border-outline-dark"`,
		`data-component-preview-header`, `Default`, `Rendered example`,
		`data-component-example-body class="component-page__example-body space-y-3"`, `data-component-example class="mt-12"`,
	} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("page missing %q", want)
		}
	}
}

func TestNormalizedExampleUsesExplicitAndContextualPreviewLabels(t *testing.T) {
	t.Parallel()
	if got := normalizedExample("Button", Example{}, true).PreviewLabel; got != "Default" {
		t.Fatalf("primary PreviewLabel = %q, want Default", got)
	}
	if got := normalizedExample("", Example{Title: "Disabled"}, false).PreviewLabel; got != "Disabled" {
		t.Fatalf("section PreviewLabel = %q, want Disabled", got)
	}
	if got := normalizedExample("", Example{PreviewLabel: "Loading"}, false).PreviewLabel; got != "Loading" {
		t.Fatalf("explicit PreviewLabel = %q, want Loading", got)
	}
}

func TestSlugKeepsUnicodeLettersAndCollapsesSeparators(t *testing.T) {
	t.Parallel()
	if got := slug("Ação / Empty State"); got != "ação-empty-state" {
		t.Fatalf("slug() = %q", got)
	}
}

func TestSectionRendersReusableVariantStructure(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	err := Section(Example{
		Title:       "Soft appearance",
		Description: "A tinted variant.",
		RootAttrs:   templ.Attributes{"data-demo-section": true},
		Preview:     templ.Raw(`<span>Soft badge</span>`),
		Code:        `@badge.Badge(cfg)`,
	}).Render(context.Background(), &output)
	if err != nil {
		t.Fatalf("Section().Render() error = %v", err)
	}
	for _, want := range []string{
		`section data-component-example class="mt-12" data-demo-section`,
		`id="soft-appearance"`,
		`data-component-example-body`,
		`Soft badge`,
		`component-example-soft-appearance`,
	} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("section missing %q", want)
		}
	}
}
