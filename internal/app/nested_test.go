package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/brohd11/bubblestack/core"
)

func nestedRepoTree(t *testing.T) string {
	t.Helper()
	base := twoRepoTree(t)
	dir := filepath.Join(base, "alpha", "child")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "-C", dir, "init", "-q", "-b", "main").CombinedOutput(); err != nil {
		t.Fatalf("git init nested repo: %v\n%s", err, out)
	}
	checkoutBase(t, base)
	return base
}

func repoNames(c *Ctx) []string {
	names := make([]string, len(c.Repos))
	for i, r := range c.Repos {
		names[i] = r.Name
	}
	return names
}

func TestNestedDiscoveryModes(t *testing.T) {
	base := nestedRepoTree(t)
	for _, nested := range []bool{false, true} {
		c := New(base, 5, "dev", nested)
		want := []string{"alpha", "beta"}
		if nested {
			want = []string{"alpha", filepath.Join("alpha", "child"), "beta"}
		}
		if got := repoNames(c); !reflect.DeepEqual(got, want) {
			t.Fatalf("New(nested=%v) found %v, want %v", nested, got, want)
		}
		if c.RootRepo == nil || !c.Pending[base] {
			t.Fatal("base checkout must stay separate and pending in both modes")
		}
		if err := c.Scan(); err != nil {
			t.Fatal(err)
		}
		if got := repoNames(c); !reflect.DeepEqual(got, want) {
			t.Fatalf("Scan(nested=%v) found %v, want %v", nested, got, want)
		}
	}
}

func TestNestedTogglePreservesListState(t *testing.T) {
	c := New(nestedRepoTree(t), 5, "dev", false)
	sh := core.NewShared(c)
	screen := NewReposScreen(sh)
	sized(core.NewRouter(sh, []core.TabEntry{{Title: "Repos", New: func(*core.Shared) core.Screen { return screen }}}))
	screen.Update(sh, keyMsg("s"))
	screen.SetCompact(true)
	screen.List().SetFilterText("alpha")
	for _, nested := range []bool{true, false} {
		_, act := screen.Update(sh, keyMsg("n"))
		if c.Nested != nested {
			t.Fatalf("n should set nested=%v", nested)
		}
		drain(sh, screen, act)
		want := 1
		if nested {
			want = 2
		}
		if len(screen.List().VisibleItems()) != want {
			t.Fatalf("nested=%v: got %d filtered rows, want %d", nested, len(screen.List().VisibleItems()), want)
		}
		if !screen.Compact() || screen.List().FilterValue() != "alpha" || !strings.Contains(screen.List().Title, "Z→A") {
			t.Fatal("nested toggle lost density, filter or sorting")
		}
		header := Header(sh)
		if strings.Contains(header, " · Nested") != nested {
			t.Fatalf("nested=%v: unexpected header: %s", nested, header)
		}
		if nested && !strings.Contains(header, "depth ≤ 5 · Nested") {
			t.Fatalf("Nested must follow depth: %s", header)
		}
	}
	// A rebuilt screen reads the session mode from the same app context.
	screen.Update(sh, keyMsg("n"))
	rebuilt := NewReposScreen(sh)
	drain(sh, rebuilt, rebuilt.Receive(sh, RescanMsg{}))
	if !c.Nested || len(c.Repos) != 3 {
		t.Fatal("screen rebuild lost nested discovery")
	}
	foundHelp := false
	for _, binding := range rebuilt.List().AdditionalFullHelpKeys() {
		foundHelp = foundHelp || binding.Help().Key == "n"
	}
	if !foundHelp {
		t.Fatal("full help must include the nested toggle")
	}
}

func TestNestedToggleDropsStalePass(t *testing.T) {
	c := New(nestedRepoTree(t), 5, "dev", false)
	sh := core.NewShared(c)
	screen := NewReposScreen(sh)
	_, first := screen.Update(sh, keyMsg("n"))
	var stale []repoDescribedMsg
	core.DrainBroadcasts(first, func(payload any) core.Action {
		if m, ok := payload.(repoDescribedMsg); ok {
			stale = append(stale, m)
		}
		return core.Action{}
	})
	_, second := screen.Update(sh, keyMsg("n"))
	// Even a result already emitted by the old pass must not reintroduce its repo.
	for _, msg := range stale {
		if ok, _ := c.applyDescribed(msg); ok {
			t.Fatal("accepted a result from the superseded pass")
		}
	}
	if !c.Refresh.Running() {
		t.Fatal("stale pass completed the current refresh")
	}
	drain(sh, screen, second)
	if c.Nested || len(c.Repos) != 2 || c.Refresh.Running() || len(c.Pending) != 0 {
		t.Fatal("latest top-level refresh did not complete cleanly")
	}
}

func TestNestedToggleWhileFilteringAndEmpty(t *testing.T) {
	c := New(t.TempDir(), 5, "dev", false)
	sh := core.NewShared(c)
	screen := NewReposScreen(sh)
	screen.SetSize(sh, 100, 30)
	_, act := screen.Update(sh, keyMsg("n"))
	drain(sh, screen, act)
	if !c.Nested || len(c.Repos) != 0 {
		t.Fatal("n must work with no discovered repos")
	}
	// Filtering dispatch must treat n as input, rather than a mode toggle.
	c = New(nestedRepoTree(t), 5, "dev", false)
	sh = core.NewShared(c)
	screen = NewReposScreen(sh)
	screen.SetSize(sh, 100, 30)
	screen.Update(sh, keyMsg("/"))
	screen.Update(sh, keyMsg("n"))
	if c.Nested || screen.List().FilterValue() != "n" {
		t.Fatal("typing n into a filter changed nested mode")
	}
}
