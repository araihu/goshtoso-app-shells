package assets

import (
	"os/exec"
	"testing"
)

func TestSystemModeBehavior(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is required for runtime unit tests")
	}
	output, err := exec.Command(node, "--test", "system-mode.test.mjs").CombinedOutput()
	if err != nil {
		t.Fatalf("System mode runtime: %v\n%s", err, output)
	}
}
