package app

import (
	"strings"
	"testing"

	"github.com/brohd11/bubblestack/components"
	"github.com/brohd11/bubblestack/core"
	"github.com/brohd11/gitstack/repoui"

	tea "charm.land/bubbletea/v2"
)

func TestDensityPreservesRepoStateAndFetchGuard(t *testing.T) {
	base := twoRepoTree(t)
	c := New(base, 5, "dev")
	sh := core.NewShared(c)
	sh.Chrome = &core.Chrome{Status: components.NewStatusLine()}
	r := core.NewRouter(sh, []core.TabEntry{{Title: "Repos", New: func(sh *core.Shared) core.Screen { return NewReposScreen(sh) }}})
	// Never execute the returned commands: fetch scheduling is tested without network IO.
	step := func(msg tea.Msg) {
		m, _ := r.Update(msg)
		r = m.(core.Router)
	}
	step(tea.WindowSizeMsg{Width: 80, Height: 24})
	root := r.Top().(*components.RootListScreen)
	root.List().SetFilterText("alpha")
	step(keyMsg("s"))
	step(keyMsg("f"))
	step(keyMsg("D"))
	if r.Top() != root || !root.Compact() || !strings.Contains(root.List().Title, "Z→A") || root.List().FilterValue() != "alpha" {
		t.Fatal("density must retain root identity, sort and filter")
	}
	step(keyMsg("f"))
	if !strings.Contains(sh.Chrome.Status.View(), "already running") {
		t.Fatal("density must not reset the fetch guard")
	}
	step(keyMsg("enter"))
	menu, ok := r.Top().(*components.PickerScreen)
	if !ok || !menu.Compact() {
		t.Fatal("shared Git menus must inherit root density")
	}
	step(keyMsg("D"))
	step(keyMsg("esc"))
	if r.Top() != root || root.Compact() {
		t.Fatal("picker density must follow back to the same root")
	}
	step(core.PropagateAll(RescanMsg{}))
	if root.List().FilterValue() != "alpha" || len(root.List().VisibleItems()) != 1 {
		t.Fatal("rescan must preserve a populated filter")
	}
	step(core.PropagateAll(repoui.FetchDoneMsg{}))
	step(keyMsg("f"))
	if !strings.Contains(sh.Chrome.Status.View(), "fetching repos") {
		t.Fatal("fetch completion must clear the guard and allow another fetch")
	}
}
