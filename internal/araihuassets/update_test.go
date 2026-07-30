package araihuassets

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestUpdateCopiesAllowlistedCatalogAssetsInStableOrderAndIsIdempotent(t *testing.T) {
	repoRoot := t.TempDir()
	releaseRoot := t.TempDir()
	release := writeReleaseFixture(t, releaseRoot)
	manifest := fixtureManifest(release.releaseJSONSHA256)
	writeJSON(t, filepath.Join(repoRoot, "araihu-assets.json"), manifest)

	result, err := Update(Options{RepoRoot: repoRoot, ReleaseRoot: releaseRoot})
	if err != nil {
		t.Fatal(err)
	}
	wantChanged := []string{
		"componentdocshell/assets/araihu.css",
		"componentdocshell/assets/goshtoso-favicon.svg",
		"componentdocshell/assets/goshtoso-logo.svg",
		"componentdocshell/assets/goshtoso-mark-reverse.svg",
		"componentdocshell/assets/goshtoso-mark.svg",
	}
	if !slices.Equal(result.Changed, wantChanged) {
		t.Fatalf("changed paths = %q, want stable order %q", result.Changed, wantChanged)
	}
	for _, mapping := range manifest.Mappings {
		got, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(mapping.Destination)))
		if err != nil {
			t.Fatal(err)
		}
		want, err := os.ReadFile(filepath.Join(releaseRoot, filepath.FromSlash(mapping.Source)))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(want) {
			t.Fatalf("%s bytes differ from allowlisted source %s", mapping.Destination, mapping.Source)
		}
	}
	if _, err := os.Stat(filepath.Join(repoRoot, "unlisted.txt")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("unlisted release file copied: %v", err)
	}

	second, err := Update(Options{RepoRoot: repoRoot, ReleaseRoot: releaseRoot})
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Changed) != 0 {
		t.Fatalf("second update changed %q, want clean", second.Changed)
	}
}

func TestUpdateRequiresExactCatalogRolesAndReleaseHashes(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(t *testing.T, releaseRoot string, manifest *Manifest)
		want   string
	}{
		{
			name: "catalog role",
			mutate: func(_ *testing.T, _ string, manifest *Manifest) {
				manifest.Mappings[2].Appearance = "adaptive"
			},
			want: `catalog role appearance = "light", want "adaptive"`,
		},
		{
			name: "canonical name",
			mutate: func(_ *testing.T, _ string, manifest *Manifest) {
				manifest.Mappings[3].CanonicalName = "goshtoso-icon-reverse"
			},
			want: `catalog canonicalName "goshtoso-icon-reverse" not found`,
		},
		{
			name: "release json identity",
			mutate: func(_ *testing.T, _ string, manifest *Manifest) {
				manifest.ReleaseJSONSHA256 = strings.Repeat("0", 64)
			},
			want: "release.json SHA-256",
		},
		{
			name: "selected file bytes",
			mutate: func(t *testing.T, releaseRoot string, _ *Manifest) {
				t.Helper()
				if err := os.WriteFile(filepath.Join(releaseRoot, "themes", "araihu.css"), []byte("thEme\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			want: "SHA-256",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repoRoot := t.TempDir()
			releaseRoot := t.TempDir()
			release := writeReleaseFixture(t, releaseRoot)
			manifest := fixtureManifest(release.releaseJSONSHA256)
			test.mutate(t, releaseRoot, &manifest)
			writeJSON(t, filepath.Join(repoRoot, "araihu-assets.json"), manifest)

			_, err := Update(Options{RepoRoot: repoRoot, ReleaseRoot: releaseRoot})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Update() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestUpdateRejectsTraversalSymlinksAndDestinationCollisions(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(t *testing.T, repoRoot, releaseRoot string, manifest *Manifest)
		want   string
	}{
		{
			name: "source traversal",
			mutate: func(_ *testing.T, _, _ string, manifest *Manifest) {
				manifest.Mappings[0].Source = "../araihu.css"
			},
			want: "unsafe source",
		},
		{
			name: "destination traversal",
			mutate: func(_ *testing.T, _, _ string, manifest *Manifest) {
				manifest.Mappings[0].Destination = "../araihu.css"
			},
			want: "unsafe destination",
		},
		{
			name: "case folded destination collision",
			mutate: func(_ *testing.T, _, _ string, manifest *Manifest) {
				manifest.Mappings[1].Destination = strings.ToUpper(manifest.Mappings[0].Destination)
			},
			want: "destination collision",
		},
		{
			name: "source symlink",
			mutate: func(t *testing.T, _, releaseRoot string, _ *Manifest) {
				t.Helper()
				target := filepath.Join(releaseRoot, "themes", "araihu.css")
				if err := os.Remove(target); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink("../unlisted.txt", target); err != nil {
					t.Fatal(err)
				}
			},
			want: "symbolic link",
		},
		{
			name: "destination parent symlink",
			mutate: func(t *testing.T, repoRoot, _ string, _ *Manifest) {
				t.Helper()
				if err := os.Symlink(t.TempDir(), filepath.Join(repoRoot, "componentdocshell")); err != nil {
					t.Fatal(err)
				}
			},
			want: "symbolic link",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repoRoot := t.TempDir()
			releaseRoot := t.TempDir()
			release := writeReleaseFixture(t, releaseRoot)
			manifest := fixtureManifest(release.releaseJSONSHA256)
			test.mutate(t, repoRoot, releaseRoot, &manifest)
			writeJSON(t, filepath.Join(repoRoot, "araihu-assets.json"), manifest)

			_, err := Update(Options{RepoRoot: repoRoot, ReleaseRoot: releaseRoot})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Update() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestUpdateRollsBackEveryFileWhenReleaseUpgradeFailsMidApply(t *testing.T) {
	repoRoot := t.TempDir()
	releaseRoot := t.TempDir()
	release := writeReleaseFixture(t, releaseRoot)
	oldManifest := fixtureManifest(release.releaseJSONSHA256)
	oldManifest.AssetsRevision = strings.Repeat("1", 40)
	oldManifest.Release = "v1.2.2"
	oldManifest.ReleaseURL = "https://github.com/araihu/assets/releases/download/v1.2.2/araihu-assets-v1.2.2.tar.gz"
	oldManifest.ReleaseSHA256 = strings.Repeat("2", 64)
	oldManifest.ReleaseJSONSHA256 = strings.Repeat("3", 64)
	writeJSON(t, filepath.Join(repoRoot, "araihu-assets.json"), oldManifest)
	writeOldFallbacks(t, repoRoot, oldManifest.Mappings)

	paths := []string{"araihu-assets.json"}
	for _, mapping := range oldManifest.Mappings {
		paths = append(paths, mapping.Destination)
	}
	before := snapshotFiles(t, repoRoot, paths)
	newManifest := fixtureManifest(release.releaseJSONSHA256)
	opts := Options{
		RepoRoot:    repoRoot,
		ReleaseRoot: releaseRoot,
		Identity: &ReleaseIdentity{
			AssetsRepository:  newManifest.AssetsRepository,
			AssetsRevision:    newManifest.AssetsRevision,
			Release:           newManifest.Release,
			ReleaseURL:        newManifest.ReleaseURL,
			ReleaseSHA256:     newManifest.ReleaseSHA256,
			ReleaseJSONSHA256: newManifest.ReleaseJSONSHA256,
		},
		beforeReplace: func(index int, _ string) error {
			if index == 2 {
				return errors.New("injected replacement failure")
			}
			return nil
		},
	}

	_, err := Update(opts)
	var applyErr *ApplyError
	if !errors.As(err, &applyErr) {
		t.Fatalf("Update() error = %v, want *ApplyError", err)
	}
	if !applyErr.RollbackComplete() {
		t.Fatalf("rollback errors = %v, want complete rollback", applyErr.RollbackErrors)
	}
	if applyErr.FailedPath == "" || len(applyErr.AppliedPaths) != 2 {
		t.Fatalf("failure state = %#v, want failed path after two replacements", applyErr)
	}
	if slices.Contains(applyErr.AppliedPaths, "araihu-assets.json") {
		t.Fatalf("manifest applied before fallback failure: %q", applyErr.AppliedPaths)
	}
	assertFilesMatch(t, repoRoot, paths, before)

	opts.beforeReplace = nil
	result, err := Update(opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Changed) != len(paths) {
		t.Fatalf("successful upgrade changed %q, want every fallback plus manifest", result.Changed)
	}
	if result.Changed[len(result.Changed)-1] != "araihu-assets.json" {
		t.Fatalf("upgrade order = %q, want manifest last", result.Changed)
	}
}

type fixtureRelease struct {
	releaseJSONSHA256 string
}

func fixtureManifest(releaseJSONSHA256 string) Manifest {
	return Manifest{
		SchemaVersion:     1,
		AssetsRepository:  "araihu/assets",
		AssetsRevision:    strings.Repeat("a", 40),
		Release:           "v1.2.3",
		ReleaseURL:        "https://github.com/araihu/assets/releases/download/v1.2.3/araihu-assets-v1.2.3.tar.gz",
		ReleaseSHA256:     strings.Repeat("b", 64),
		ReleaseJSONSHA256: releaseJSONSHA256,
		Mappings: []Mapping{
			{Source: "themes/araihu.css", Destination: "componentdocshell/assets/araihu.css"},
			{Source: "brand/goshtoso/logo/adaptive-transparent-optical.svg", Destination: "componentdocshell/assets/goshtoso-logo.svg", CanonicalName: "goshtoso-logo-adaptive-transparent-optical", Namespace: "brand", Product: "goshtoso", Artwork: "logo", Appearance: "adaptive", Surface: "transparent", Framing: "optical", Format: "svg"},
			{Source: "icons/brand/goshtoso-icon-light-transparent-optical.svg", Destination: "componentdocshell/assets/goshtoso-mark.svg", CanonicalName: "goshtoso-icon-light-transparent-optical", Namespace: "brand", Product: "goshtoso", Artwork: "icon", Appearance: "light", Surface: "transparent", Framing: "optical", Format: "svg"},
			{Source: "icons/brand/goshtoso-icon-dark-transparent-optical.svg", Destination: "componentdocshell/assets/goshtoso-mark-reverse.svg", CanonicalName: "goshtoso-icon-dark-transparent-optical", Namespace: "brand", Product: "goshtoso", Artwork: "icon", Appearance: "dark", Surface: "transparent", Framing: "optical", Format: "svg"},
			{Source: "platform/web/goshtoso/favicon.svg", Destination: "componentdocshell/assets/goshtoso-favicon.svg", CanonicalName: "platform-web-goshtoso-favicon-svg", Namespace: "brand", Product: "goshtoso", Artwork: "icon", Appearance: "adaptive", Surface: "transparent", Framing: "optical", Format: "svg"},
		},
	}
}

func writeReleaseFixture(t *testing.T, root string) fixtureRelease {
	t.Helper()
	files := map[string][]byte{
		"themes/araihu.css": []byte("theme\n"),
		"brand/goshtoso/logo/adaptive-transparent-optical.svg":    []byte("<svg>logo</svg>\n"),
		"icons/brand/goshtoso-icon-light-transparent-optical.svg": []byte("<svg>mark</svg>\n"),
		"icons/brand/goshtoso-icon-dark-transparent-optical.svg":  []byte("<svg>reverse</svg>\n"),
		"platform/web/goshtoso/favicon.svg":                       []byte("<svg>favicon</svg>\n"),
		"unlisted.txt":                                            []byte("do not copy\n"),
	}
	catalog := map[string]any{
		"schemaVersion": 1, "release": "v1.2.3", "identityRevision": 11,
		"assets": []map[string]any{
			fixtureCatalogAsset("goshtoso-logo-adaptive-transparent-optical", "brand/goshtoso/logo/adaptive-transparent-optical.svg", "logo", "adaptive", files),
			fixtureCatalogAsset("goshtoso-icon-light-transparent-optical", "icons/brand/goshtoso-icon-light-transparent-optical.svg", "icon", "light", files),
			fixtureCatalogAsset("goshtoso-icon-dark-transparent-optical", "icons/brand/goshtoso-icon-dark-transparent-optical.svg", "icon", "dark", files),
			fixtureCatalogAsset("platform-web-goshtoso-favicon-svg", "platform/web/goshtoso/favicon.svg", "icon", "adaptive", files),
		},
	}
	catalogBytes := marshalJSON(t, catalog)
	files["catalog.json"] = catalogBytes
	for name, contents := range files {
		writeFile(t, filepath.Join(root, filepath.FromSlash(name)), contents)
	}
	inventory := make([]map[string]any, 0, len(files))
	for name, contents := range files {
		inventory = append(inventory, map[string]any{"path": name, "sha256": fixtureDigest(contents), "size": len(contents)})
	}
	releaseDocument := map[string]any{
		"schemaVersion": 1, "release": "v1.2.3", "identityRevision": 11, "runtimeVersion": 1,
		"catalogSha256": fixtureDigest(catalogBytes), "themesSha256": strings.Repeat("c", 64),
		"campaignsSha256": strings.Repeat("d", 64), "files": inventory,
	}
	releaseBytes := marshalJSON(t, releaseDocument)
	writeFile(t, filepath.Join(root, "release.json"), releaseBytes)
	return fixtureRelease{releaseJSONSHA256: fixtureDigest(releaseBytes)}
}

func fixtureCatalogAsset(name, sourcePath, artwork, appearance string, files map[string][]byte) map[string]any {
	return map[string]any{
		"canonicalName": name, "namespace": "brand", "path": sourcePath, "product": "goshtoso",
		"artwork": artwork, "appearance": appearance, "surface": "transparent", "framing": "optical", "format": "svg",
		"dimensions": map[string]any{"viewBox": "0 0 16 16"}, "spriteSymbol": "", "colorBehavior": "protected",
		"license": "Arai Hu Brand Terms", "source": "fixture", "sha256": fixtureDigest(files[sourcePath]),
	}
}

func writeOldFallbacks(t *testing.T, root string, mappings []Mapping) {
	t.Helper()
	for index, mapping := range mappings {
		writeFile(t, filepath.Join(root, filepath.FromSlash(mapping.Destination)), []byte(fmt.Sprintf("old-%d\n", index)))
	}
}

func snapshotFiles(t *testing.T, root string, paths []string) map[string][]byte {
	t.Helper()
	snapshot := make(map[string][]byte, len(paths))
	for _, name := range paths {
		contents, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		snapshot[name] = contents
	}
	return snapshot
}

func assertFilesMatch(t *testing.T, root string, paths []string, want map[string][]byte) {
	t.Helper()
	for _, name := range paths {
		got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(want[name]) {
			t.Fatalf("%s changed after rollback\ngot: %q\nwant: %q", name, got, want[name])
		}
	}
}

func writeJSON(t *testing.T, name string, value any) {
	t.Helper()
	writeFile(t, name, marshalJSON(t, value))
}

func writeFile(t *testing.T, name string, contents []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, contents, 0o644); err != nil {
		t.Fatal(err)
	}
}

func marshalJSON(t *testing.T, value any) []byte {
	t.Helper()
	contents, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return append(contents, '\n')
}

func fixtureDigest(contents []byte) string {
	sum := sha256.Sum256(contents)
	return hex.EncodeToString(sum[:])
}
