package app

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/brohd11/bubblestack/components"
	"github.com/brohd11/bubblestack/core"
	"github.com/brohd11/gitstack/repo"
	"github.com/brohd11/gitstack/repoui"
)

func TestTargetedRefreshKeepsOtherReposCached(t *testing.T) {
	base := twoRepoTree(t)
	checkoutBase(t, base)
	c := New(base, 5, "dev", false)
	c.Scan()
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
	drain(sh, screen, screen.Receive(sh, repoui.RepoRefreshMsg{Dir: base}))
	if !c.Repos[0].Dirty {
		t.Fatal("root action must perform a full scan")
	}
}

// drain runs a refresh to completion, feeding its broadcasts back to the screen, and returns
// the status lines it wrote.
func drain(sh *core.Shared, screen core.Receiver, act core.Action) []string {
	return core.DrainBroadcasts(act, func(p any) core.Action { return screen.Receive(sh, p) })
}

// TestAsyncRefreshLandsRows: startup only discovers (rows pending, nothing described), a
// rescan returns at once with every row muted, and the results land each repo's state,
// clear the pending marks and finish with "Refreshed".
func TestAsyncRefreshLandsRows(t *testing.T) {
	base := twoRepoTree(t)
	checkoutBase(t, base)
	c := New(base, 5, "dev", false)
	if len(c.Repos) != 2 || !c.Pending[c.Repos[1].Dir] || !c.Pending[base] || c.Repos[1].Dirty {
		t.Fatal("New must discover both repos and the root as pending without describing them")
	}
	sh := core.NewShared(c)
	screen := NewReposScreen(sh)
	screen.SetSize(sh, 80, 24)
	act := screen.Receive(sh, RescanMsg{})
	for _, it := range screen.List().Items() {
		if row := it.(components.Item); !row.Pending || !strings.Contains(row.Desc, "refreshing") {
			t.Fatalf("row %q must be pending until its result lands", row.Name)
		}
	}
	statuses := drain(sh, screen, act)
	if len(statuses) == 0 || statuses[len(statuses)-1] != "Refreshed" {
		t.Fatalf("refresh must finish with Refreshed, got %q", statuses)
	}
	if len(c.Pending) != 0 || c.Refresh.Running() {
		t.Fatal("every pending mark must clear once the pass lands")
	}
	if !c.Repos[1].Dirty || c.Repos[1].Branch == "" {
		t.Fatal("beta's described state must land")
	}
	for _, it := range screen.List().Items() {
		if it.(components.Item).Pending {
			t.Fatal("rows must un-mute once described")
		}
	}
}

// TestAsyncRefreshDropsStalePass: a rescan superseding a running one drops the first pass's
// results rather than counting them toward the second.
func TestAsyncRefreshDropsStalePass(t *testing.T) {
	c := New(twoRepoTree(t), 5, "dev", false)
	sh := core.NewShared(c)
	screen := NewReposScreen(sh)
	first := screen.Receive(sh, RescanMsg{})
	second := screen.Receive(sh, RescanMsg{})
	drain(sh, screen, first)
	if !c.Refresh.Running() || len(c.Pending) == 0 {
		t.Fatal("a superseded pass's results must not complete the running one")
	}
	drain(sh, screen, second)
	if c.Refresh.Running() || len(c.Pending) != 0 {
		t.Fatal("the current pass must complete")
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
