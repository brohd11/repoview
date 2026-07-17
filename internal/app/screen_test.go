package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/brohd11/bubblestack/components"
	"github.com/brohd11/bubblestack/core"
	"github.com/brohd11/gitstack/repoui"

	tea "github.com/charmbracelet/bubbletea"
)

// twoRepoTree builds a directory with two real git checkouts under it: "alpha" (clean, on
// main) and "beta" (a dirty working tree). It's the fixture the screen e2e drives.
func twoRepoTree(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	base := t.TempDir()
	git := func(dir string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	initRepo := func(name string) string {
		dir := filepath.Join(base, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		git(dir, "init", "-q", "-b", "main")
		git(dir, "remote", "add", "origin", "https://github.com/owner/"+name+".git")
		os.WriteFile(filepath.Join(dir, "f"), []byte("x"), 0o644)
		git(dir, "add", ".")
		git(dir, "commit", "-q", "-m", "init")
		return dir
	}
	initRepo("alpha")
	beta := initRepo("beta")
	os.WriteFile(filepath.Join(beta, "wip.txt"), []byte("uncommitted"), 0o644) // make beta dirty
	return base
}

// router builds the repoview router the same way Run does, rooted at a real scanned tree.
func router(root string) core.Router {
	r, _ := routerWithPane(root)
	return r
}

// routerWithPane is router, also handing back the log pane. Router keeps its Shared
// unexported, so a test that needs to assert on the pane (the w-key arbitration) has to
// hold the reference from construction.
func routerWithPane(root string) (core.Router, *components.LogPane) {
	pane := components.NewLogPane()
	sh := core.NewShared(New(root, 5))
	sh.Chrome = &core.Chrome{Header: core.NewHeaderPane(Header), Output: pane, Status: components.NewStatusLine()}
	return core.NewRouter(sh, []core.TabEntry{
		{Title: "Repos", New: func(sh *core.Shared) core.Screen { return NewReposScreen(sh) }},
	}), pane
}

func sized(tm tea.Model) tea.Model {
	tm, _ = tm.Update(tea.WindowSizeMsg{Width: 90, Height: 30})
	return tm
}

// pump delivers msg, then runs the returned command and feeds its (single, non-batch) result
// back — enough to drive the navigation commands (push/pop) and broadcasts.
func pump(tm tea.Model, msg tea.Msg) tea.Model {
	tm, cmd := tm.Update(msg)
	for i := 0; i < 8 && cmd != nil; i++ {
		out := cmd()
		if out == nil {
			break
		}
		if _, isBatch := out.(tea.BatchMsg); isBatch {
			break
		}
		tm, cmd = tm.Update(out)
	}
	return tm
}

// TestReposScreenWiring drives the whole tool: the scanned list renders with the right markers,
// enter opens a repo's shared git submenu, V the all-repos menu, a the Actions menu, and a git
// flow's RefreshMsg rebuilds the list.
func TestReposScreenWiring(t *testing.T) {
	tm := sized(router(twoRepoTree(t)))

	if _, ok := tm.(core.Router).Top().(*ReposScreen); !ok {
		t.Fatalf("want the Repos screen on top, got %T", tm.(core.Router).Top())
	}
	// The list shows both repos, and the dirty one carries the marker.
	if out := tm.View(); !strings.Contains(out, "alpha") || !strings.Contains(out, "beta") ||
		!strings.Contains(out, "uncommitted changes") {
		t.Errorf("repo list should show both repos and beta's dirty marker:\n%s", out)
	}

	// enter on the highlighted row (alpha, first in the sorted scan) opens its git submenu.
	tm = pump(tm, tea.KeyMsg{Type: tea.KeyEnter})
	if _, ok := tm.(core.Router).Top().(*components.PickerScreen); !ok {
		t.Fatalf("enter should open the per-repo Git submenu (PickerScreen), got %T", tm.(core.Router).Top())
	}
	if out := tm.View(); !strings.Contains(out, "alpha") || !strings.Contains(out, "Pull") {
		t.Errorf("submenu should be the repo's Git menu:\n%s", out)
	}

	// esc back, then "v" (the row-level alias, dispatched via Item.Keys) opens the same submenu.
	tm = pump(tm, tea.KeyMsg{Type: tea.KeyEsc})
	tm = pump(tm, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	if _, ok := tm.(core.Router).Top().(*components.PickerScreen); !ok {
		t.Fatalf("v should open the per-repo Git submenu, got %T", tm.(core.Router).Top())
	}
	if out := tm.View(); !strings.Contains(out, "Pull") {
		t.Errorf("v submenu should be the repo's Git menu:\n%s", out)
	}

	// esc back, then V opens the all-repos menu.
	tm = pump(tm, tea.KeyMsg{Type: tea.KeyEsc})
	tm = pump(tm, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'V'}})
	if _, ok := tm.(core.Router).Top().(*components.PickerScreen); !ok {
		t.Fatalf("V should open the all-repos Git menu, got %T", tm.(core.Router).Top())
	}
	if out := tm.View(); !strings.Contains(out, "all repos") {
		t.Errorf("all-repos menu title should say so:\n%s", out)
	}

	// esc back, then a opens the Actions menu with Theme + Refresh.
	tm = pump(tm, tea.KeyMsg{Type: tea.KeyEsc})
	tm = pump(tm, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if _, ok := tm.(core.Router).Top().(*components.PickerScreen); !ok {
		t.Fatalf("a should open the Actions menu, got %T", tm.(core.Router).Top())
	}
	if out := tm.View(); !strings.Contains(out, "Theme") || !strings.Contains(out, "Refresh") {
		t.Errorf("Actions menu should list Theme and Refresh:\n%s", out)
	}

	// esc back to the root; a git flow's RefreshMsg must rebuild the list without panicking.
	tm = pump(tm, tea.KeyMsg{Type: tea.KeyEsc})
	tm = pump(tm, core.PropagateAll(RescanMsg{}))
	if _, ok := tm.(core.Router).Top().(*ReposScreen); !ok {
		t.Fatalf("after refresh, want the Repos screen on top, got %T", tm.(core.Router).Top())
	}
	if out := tm.View(); !strings.Contains(out, "alpha") {
		t.Errorf("rebuilt list should still show the repos:\n%s", out)
	}
}

// dirtyRepoTree is twoRepoTree plus a real edit to a tracked file in beta, so there is an
// actual diff to render rather than only an untracked file.
func dirtyRepoTree(t *testing.T) string {
	t.Helper()
	base := twoRepoTree(t)
	beta := filepath.Join(base, "beta")
	// f was committed as "x"; rewrite it so `diff HEAD` has a deletion and an addition,
	// with a line long enough to exercise truncation and wrapping.
	os.WriteFile(filepath.Join(beta, "f"),
		[]byte("y"+strings.Repeat("longtoken", 30)+"\nsecond\n"), 0o644)
	return base
}

// TestDiffViewWiring drives the Diff flow the way a user does: down to the dirty repo, into
// its git menu, into Diff, into a file, then the layout and wrap toggles. It runs against a
// real git checkout through the real router and the real log pane, which is what makes the
// w-key arbitration meaningful — the pane implements Wrapper too, so the routing rule is
// under genuine contention here rather than against a stub.
func TestDiffViewWiring(t *testing.T) {
	r, pane := routerWithPane(dirtyRepoTree(t))
	tm := sized(r)

	// down to beta (the dirty one; alpha sorts first), then v for its git menu.
	tm = pump(tm, tea.KeyMsg{Type: tea.KeyDown})
	tm = pump(tm, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	if out := tm.View(); !strings.Contains(out, "Diff") {
		t.Fatalf("the git menu should offer a Diff row:\n%s", out)
	}

	// Diff sits under Status: down once from Fetch, twice to reach it.
	tm = pump(tm, tea.KeyMsg{Type: tea.KeyDown})
	tm = pump(tm, tea.KeyMsg{Type: tea.KeyDown})
	tm = pump(tm, tea.KeyMsg{Type: tea.KeyEnter})
	if _, ok := tm.(core.Router).Top().(*components.PickerScreen); !ok {
		t.Fatalf("Diff should open the file picker, got %T", tm.(core.Router).Top())
	}
	out := tm.View()
	if !strings.Contains(out, "beta") || !strings.Contains(out, "wip.txt") {
		t.Errorf("the picker should list the repo row and the changed files:\n%s", out)
	}

	// The top row is the whole repo's diff.
	tm = pump(tm, tea.KeyMsg{Type: tea.KeyEnter})
	top := tm.(core.Router).Top()
	if _, ok := top.(*repoui.DiffScreen); !ok {
		t.Fatalf("selecting a row should open the DiffScreen, got %T", top)
	}
	out = tm.View()
	if !strings.Contains(out, "second") {
		t.Errorf("the diff should show the added line:\n%s", out)
	}
	// It opens in auto, which names no layout — the title is just the file. At 90 cols auto
	// resolves to unified and says nothing about it: that is the mode working, not failing.
	if strings.Contains(out, "unified") || strings.Contains(out, "side by side") {
		t.Errorf("auto should not name a layout in the title bar:\n%s", out)
	}
	if n := columnRules(t, out, "second"); n != 1 {
		t.Errorf("auto should resolve to unified at 90 cols — the content row has %d column "+
			"rule(s), want 1:\n%s", n, out)
	}

	// s cycles auto → unified → side by side. The first press is the explicit unified,
	// which at this width renders the same thing auto just did but now says so.
	tm = pump(tm, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if out = tm.View(); !strings.Contains(out, "unified") {
		t.Errorf("the first press of s should select unified and label it:\n%s", out)
	}
	if strings.Contains(out, "needs") {
		t.Errorf("unified fits at any width — nothing to warn about:\n%s", out)
	}

	// The second press asks for side by side. The terminal is 90 cols — under
	// minSplitWidth — so it must fall back and explain itself rather than look like a
	// dead key. This is the one case that still warns; auto never does.
	tm = pump(tm, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if out = tm.View(); !strings.Contains(out, "needs") {
		t.Errorf("at 90 cols, an explicit side by side should fall back and say why:\n%s", out)
	}

	// Widen past minSplitWidth: the same request now takes effect.
	tm, _ = tm.Update(tea.WindowSizeMsg{Width: 140, Height: 30})
	if out = tm.View(); !strings.Contains(out, "side by side") || strings.Contains(out, "needs") {
		t.Errorf("at 140 cols, side by side should engage and drop the warning:\n%s", out)
	}

	// w must wrap the DIFF, not the log pane, while the pane is unfocused.
	tm = pump(tm, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	if d := tm.(core.Router).Top().(*repoui.DiffScreen); !d.Wrapped() {
		t.Error("w should wrap the diff while the output pane is unfocused")
	}
	if pane.Wrapped() {
		t.Error("w must not wrap the log pane while the diff is on top and the pane unfocused")
	}
	if out = tm.View(); !strings.Contains(out, "wrap") {
		t.Errorf("the title bar should advertise wrap mode:\n%s", out)
	}

	// esc leaves the diff and returns to the picker, rather than quitting the flow.
	tm = pump(tm, tea.KeyMsg{Type: tea.KeyEsc})
	if _, ok := tm.(core.Router).Top().(*components.PickerScreen); !ok {
		t.Fatalf("esc should return to the file picker, got %T", tm.(core.Router).Top())
	}

	// An untracked file has no HEAD blob, so it takes the --no-index path (which exits 1
	// by design). Drive it: wip.txt is the last row, below the repo row and f.
	tm = pump(tm, tea.KeyMsg{Type: tea.KeyDown})
	tm = pump(tm, tea.KeyMsg{Type: tea.KeyDown})
	tm = pump(tm, tea.KeyMsg{Type: tea.KeyEnter})
	if _, ok := tm.(core.Router).Top().(*repoui.DiffScreen); !ok {
		t.Fatalf("an untracked file should open a DiffScreen, got %T", tm.(core.Router).Top())
	}
	if out = tm.View(); !strings.Contains(out, "uncommitted") {
		t.Errorf("an untracked file should diff against /dev/null and show its contents as additions:\n%s", out)
	}
	if strings.Contains(out, "could not read the diff") {
		t.Errorf("the --no-index exit status 1 must not be treated as a failure:\n%s", out)
	}

	// This screen was opened fresh at 140 cols, so auto resolves to side by side — the
	// point of the default: a wide terminal gives the better layout with no keypress. The
	// separator is the tell, since auto doesn't announce itself in the title.
	// This screen was opened fresh at 140 cols, so auto resolves to side by side — the
	// point of the default: a wide terminal gives the better layout with no keypress.
	// Auto doesn't announce itself, so the tell is the row's shape, not the title.
	if n := columnRules(t, out, "uncommitted"); n != 2 {
		t.Errorf("auto should open side by side at 140 cols without a keypress — the content row "+
			"has %d column rule(s), want 2 (unified would have 1):\n%s", n, out)
	}
}

// columnRules counts the column rules ("│") on the rendered row carrying needle. It is how
// these tests tell the layouts apart: a side-by-side row has two (one per column's line-number
// gutter), a unified row has one. Counting them across the whole frame would prove nothing —
// the header box is drawn with the same character.
func columnRules(t *testing.T, view, needle string) int {
	t.Helper()
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, needle) {
			return strings.Count(line, "│")
		}
	}
	t.Fatalf("no rendered row contains %q:\n%s", needle, view)
	return 0
}
