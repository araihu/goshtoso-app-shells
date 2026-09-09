package consoleshell

import (
	"bytes"
	"context"
	"github.com/araihu/goshtoso/components/head"
	"strings"
	"testing"
)

func TestDrawerIconMenu(t *testing.T) {
	cfg := validConfig()
	cfg.Navigation.Drawer = true
	cfg.Navigation.IconOnlyMenu = true
	html := render(t, Layout(cfg, validPage()))
	for _, want := range []string{`console-shell-root--drawer`, `id="consoleshell-menu"`, `hi-16-solid-bars-3`, `consoleshell-menu-tooltip`, `aria-label="Open navigation"`} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %q", want)
		}
	}
	if strings.Count(html, `id="consoleshell-menu"`) != 1 {
		t.Fatal("duplicate menu control")
	}
}

func TestPageMetadataRenderedOnceAndExcludedFromFragments(t *testing.T) {
	cfg, page := validConfig(), validPage()
	page.Description = "Review jobs."
	page.CanonicalURL = "https://example.com/jobs"
	page.Metadata = &head.MetadataConfig{Image: head.SocialImage{URL: "https://example.com/og.png", MIMEType: "image/png", Width: 1280, Height: 640, Alt: "Jobs"}}
	html := render(t, Layout(cfg, page))
	for _, tag := range []string{`<title>Jobs · Console</title>`, `name="description"`, `rel="canonical"`, `property="og:url"`, `name="twitter:card"`} {
		if strings.Count(html, tag) != 1 {
			t.Errorf("expected one %s", tag)
		}
	}
	cfg.Interactions.EnableHTMX = true
	if fragment := render(t, Fragment(cfg, page)); strings.Contains(fragment, `property="og:`) {
		t.Fatal("metadata leaked into fragment")
	}
	page.Metadata.Image.URL = "/relative.png"
	var output bytes.Buffer
	if err := Layout(cfg, page).Render(context.Background(), &output); err == nil || output.Len() != 0 {
		t.Fatal("invalid metadata must fail before writing")
	}
}
