package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlaywrightDependencyUsesMaintainedModulePath(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Join("..", "..")
	for _, relativePath := range []string{
		"go.mod",
		filepath.Join(".dagger", "src", "index.ts"),
		filepath.Join("example", "e2e", "family_navigation_test.go"),
	} {
		body, err := os.ReadFile(filepath.Join(repoRoot, relativePath))
		if err != nil {
			t.Fatalf("read %s: %v", relativePath, err)
		}
		content := string(body)
		if !strings.Contains(content, "github.com/mxschmitt/playwright-go") {
			t.Errorf("%s must use maintained Playwright Go module path", relativePath)
		}
		if strings.Contains(content, "github.com/playwright-community/playwright-go") {
			t.Errorf("%s still uses deprecated Playwright Go module path", relativePath)
		}
	}
}
