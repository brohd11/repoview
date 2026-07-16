package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// nestedRepoTree builds base/<name>/.git for each name (real git init), so FindGitRepos sees
// them as depth-1 checkouts. Mirrors the fixture style in internal/app/screen_test.go.
func nestedRepoTree(t *testing.T, names ...string) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	base := t.TempDir()
	for _, name := range names {
		dir := filepath.Join(base, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command("git", "-C", dir, "init", "-q", "-b", "main")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git init %s: %v\n%s", name, err, out)
		}
	}
	return base
}

// TestReposListMode drives the repos subcommand with no command (list mode) and captures its
// output, asserting every nested repo path is printed.
func TestReposListMode(t *testing.T) {
	base := nestedRepoTree(t, "alpha", "beta")

	// Set the package-level flag vars the way cobra would, then run through RunE with a captured
	// writer so the full scan→print path is exercised.
	reposDir, reposRaw, reposDirty, reposDepth = base, false, false, 1
	var buf bytes.Buffer
	reposCmd.SetOut(&buf)
	reposCmd.SetErr(&buf)
	t.Cleanup(func() { reposCmd.SetOut(nil); reposCmd.SetErr(nil) })

	if err := runRepos(reposCmd, nil); err != nil {
		t.Fatalf("runRepos: %v", err)
	}

	got := buf.String()
	for _, want := range []string{"alpha", "beta"} {
		if !strings.Contains(got, want) {
			t.Errorf("list output missing %q:\n%s", want, got)
		}
	}
}
