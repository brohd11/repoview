package app

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/brohd11/bubblestack/core"
	"github.com/brohd11/gitstack/repo"
	"github.com/brohd11/gitstack/repoui"
)

func TestTargetedRefreshKeepsOtherReposCached(t *testing.T) {
	base := twoRepoTree(t)
	checkoutBase(t, base)
	c := New(base, 5, "dev")
	sh := core.NewShared(c)
	screen := NewReposScreen(sh)
	screen.SetSize(sh, 80, 24)
	screen.List().SetFilterText("beta")
	screen.SetCompact(true)
	alpha, beta := filepath.Join(base, "alpha"), filepath.Join(base, "beta")
	if err := os.WriteFile(filepath.Join(alpha, "new.txt"), []byte("dirty"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(beta, "wip.txt")); err != nil {
		t.Fatal(err)
	}
	// The root is intentionally stale to verify parent-marker refresh as well.
	c.RootRepo.Dirty = false
	screen.Receive(sh, repoui.RepoRefreshMsg{Dir: beta})
	if c.Repos[0].Dirty || c.Repos[1].Dirty {
		t.Fatal("target must become clean while alpha stays cached clean")
	}
	if !c.RootRepo.Dirty {
		t.Fatal("enclosing root marker was not refreshed")
	}
	if screen.List().FilterValue() != "beta" || !screen.Compact() {
		t.Fatal("refresh dropped filter/density")
	}
	if strings.Contains(screen.List().VisibleItems()[0].FilterValue(), "uncommitted") {
		t.Fatal("target row is stale")
	}
	for _, msg := range []repoui.RepoRefreshMsg{{}, {Dir: filepath.Join(base, "unknown")}} {
		screen.Receive(sh, msg)
		if c.Repos[0].Dirty {
			t.Fatal("unknown target caused a full scan")
		}
	}
	screen.Receive(sh, repoui.RepoRefreshMsg{Dir: base})
	if !c.Repos[0].Dirty {
		t.Fatal("root action must perform a full scan")
	}
}

func TestTargetedRefreshGitQueriesStayScoped(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell git probe")
	}
	base, bin := t.TempDir(), t.TempDir()
	log := filepath.Join(t.TempDir(), "calls")
	// Counting at the executable boundary catches accidental scans in row rebuilding.
	script := "#!/bin/sh\nprintf '%s\\n' \"$2\" >> \"$REFRESH_TEST_LOG\"\n"
	if err := os.WriteFile(filepath.Join(bin, "git"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("REFRESH_TEST_LOG", log)
	c := &Ctx{Root: base, Depth: 5}
	for i := 0; i < 30; i++ {
		dir := filepath.Join(base, fmt.Sprintf("repo%02d", i))
		if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
			t.Fatal(err)
		}
		c.Repos = append(c.Repos, repo.Repo{Name: filepath.Base(dir), Dir: dir})
	}
	if err := os.Mkdir(filepath.Join(base, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	c.RootRepo = &repo.Repo{Name: "root", Dir: base, Root: true}
	sh := core.NewShared(c)
	screen := NewReposScreen(sh)
	screen.Receive(sh, repoui.RepoRefreshMsg{Dir: c.Repos[0].Dir})
	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	calls := strings.Fields(string(data))
	if len(calls) != 6 {
		t.Fatalf("got %d git queries, want 3 each for target and root: %s", len(calls), data)
	}
	for _, dir := range calls {
		if dir != base && dir != c.Repos[0].Dir {
			t.Fatalf("queried unrelated repo %s", dir)
		}
	}
}
