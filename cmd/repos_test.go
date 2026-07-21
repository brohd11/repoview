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
	reposDir, reposRaw, reposDirty, reposDepth, reposIncludeRoot = base, false, false, 1, false
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

// TestReposIncludeRoot: --include-root opts the scanned base itself into the listing as ".", but
// only when it is a git checkout, and never without the flag.
func TestReposIncludeRoot(t *testing.T) {
	base := nestedRepoTree(t, "alpha", "beta")
	// Make base itself a checkout too.
	if out, err := exec.Command("git", "-C", base, "init", "-q", "-b", "main").CombinedOutput(); err != nil {
		t.Fatalf("git init base: %v\n%s", err, out)
	}

	run := func(includeRoot bool) string {
		t.Helper()
		reposDir, reposRaw, reposDirty, reposDepth, reposIncludeRoot = base, false, false, 1, includeRoot
		var buf bytes.Buffer
		reposCmd.SetOut(&buf)
		reposCmd.SetErr(&buf)
		t.Cleanup(func() { reposCmd.SetOut(nil); reposCmd.SetErr(nil) })
		if err := runRepos(reposCmd, nil); err != nil {
			t.Fatalf("runRepos: %v", err)
		}
		return buf.String()
	}

	// Without the flag, the base itself is excluded (no "." line).
	for _, line := range strings.Split(run(false), "\n") {
		if strings.TrimSpace(line) == "." {
			t.Errorf("root should be excluded without --include-root")
		}
	}

	// With it, "." appears alongside the nested repos.
	got := run(true)
	var haveRoot bool
	for _, line := range strings.Split(got, "\n") {
		if strings.TrimSpace(line) == "." {
			haveRoot = true
		}
	}
	if !haveRoot {
		t.Errorf("--include-root should list the base as \".\":\n%s", got)
	}
	for _, want := range []string{"alpha", "beta"} {
		if !strings.Contains(got, want) {
			t.Errorf("--include-root output missing nested repo %q:\n%s", want, got)
		}
	}
}
