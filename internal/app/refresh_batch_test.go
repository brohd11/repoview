package app

import (
	"fmt"
	"testing"

	"github.com/brohd11/bubblestack/components"
	"github.com/brohd11/bubblestack/core"
	"github.com/brohd11/gitstack/repo"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

func batchFixture(n int) (*Ctx, *core.Shared, *reposState, int64) {
	c := &Ctx{Pending: make(map[string]bool), dirtyRows: make(map[string]bool), rankSnap: make(map[string]int)}
	for i := range n {
		name := fmt.Sprintf("repo%d", i)
		c.Repos = append(c.Repos, repo.Repo{Name: name, Dir: "/" + name, Branch: "main"})
		c.Pending["/"+name] = true
		c.rankSnap["/"+name] = rankClean
	}
	c.indexRepos()
	gen, _ := c.Refresh.Start(n)
	sh := core.NewShared(c)
	s := &reposState{sort: components.SortStatus}
	s.screen = components.NewRootList(s.rebuildItems(sh), components.RootListOpts{})
	s.screen.SetSize(sh, 100, 30)
	return c, sh, s, gen
}

// Let the timer generate its opaque broadcast, then hold delivery explicitly.
// Assertions depend on message order, never on sleeping until an expected deadline.
func queuedPublication(t *testing.T, act core.Action) any {
	t.Helper()
	var tick any
	core.DrainBroadcasts(act, func(payload any) core.Action {
		if tick != nil {
			t.Fatal("a burst scheduled multiple publications")
		}
		tick = payload
		return core.Action{}
	})
	if tick == nil {
		t.Fatal("no publication scheduled")
	}
	return tick
}

func TestRefreshBatchPublishesDirtyRowsAndFinalSort(t *testing.T) {
	c, sh, s, gen := batchFixture(3)
	s.screen.List().SetFilterText("repo")
	s.screen.List().SetFilterState(list.Filtering)
	s.screen.List().Select(1)
	old := s.screen.List().Items()
	first := c.Repos[0]
	first.Dirty = true
	tick := queuedPublication(t, s.receive(sh, repoDescribedMsg{gen: gen, repo: first}))
	second := c.Repos[1]
	second.Dirty = true
	s.receive(sh, repoDescribedMsg{gen: gen, repo: second})
	if !old[1].(components.Item).Pending || !s.screen.List().Items()[1].(components.Item).Pending || !c.Repos[1].Dirty {
		t.Fatal("cache must advance immediately while published rows remain unchanged")
	}
	untouched := s.rows[c.Repos[2].Dir].(components.Item)
	untouched.Desc = "cached sentinel"
	s.rows[c.Repos[2].Dir] = untouched
	filters := 0
	s.screen.List().Filter = func(term string, targets []string) []list.Rank {
		filters++
		return list.DefaultFilter(term, targets)
	}
	s.receive(sh, tick)
	if filters != 1 || s.rows[c.Repos[1].Dir].(components.Item).Pending || s.rows[c.Repos[2].Dir].(components.Item).Desc != "cached sentinel" {
		t.Fatal("flush must filter once, update dirty rows, and reuse unchanged rows")
	}
	third := c.Repos[2]
	third.Sync.Behind = 1
	s.receive(sh, repoDescribedMsg{gen: gen, repo: third})
	items := s.screen.List().Items()
	if c.Refresh.Running() || items[0].(components.Item).Key != third.Dir || c.rankSnap != nil {
		t.Fatal("terminal result must immediately apply final status ordering")
	}
	if s.screen.List().SelectedItem().(components.Item).Key != second.Dir || s.screen.List().FilterState() != list.Filtering {
		t.Fatal("marker changes and final sorting lost selection or typing state")
	}
	for _, row := range items {
		if row.(components.Item).Pending {
			t.Fatal("terminal publication left a row pending")
		}
	}
	s.receive(sh, tick)
	if &s.screen.List().Items()[0] != &items[0] {
		t.Fatal("retired tick republished a completed list")
	}
}

func TestRefreshBatchRejectsInvalidResultsAndRootOnlyUpdate(t *testing.T) {
	c, sh, s, _ := batchFixture(2)
	c.RootRepo = &repo.Repo{Name: "root", Dir: "/root", Root: true}
	c.Pending["/root"] = true
	gen, _ := c.Refresh.Start(3)
	items := s.screen.List().Items()
	root := *c.RootRepo
	root.Branch = "updated"
	act := s.receive(sh, repoDescribedMsg{gen: gen, repo: root})
	if c.RootRepo.Branch != "updated" || act.Cmd != nil || &s.screen.List().Items()[0] != &items[0] {
		t.Fatal("root-only result must update the header without publishing rows")
	}
	first := repoDescribedMsg{gen: gen, repo: c.Repos[0]}
	s.receive(sh, first)
	for _, msg := range []repoDescribedMsg{
		first, {gen: gen - 1, repo: c.Repos[1]}, {gen: gen, repo: repo.Repo{Dir: "/unknown"}},
	} {
		if ok, _ := c.applyDescribed(msg); ok {
			t.Fatal("invalid result advanced progress")
		}
	}
	if !c.Refresh.Running() || !c.Pending[c.Repos[1].Dir] {
		t.Fatal("invalid result consumed a pending target")
	}
	s.receive(sh, repoDescribedMsg{gen: gen, repo: c.Repos[1]})
	if c.Refresh.Running() {
		t.Fatal("valid terminal result did not complete")
	}
}

func TestQueuedPublicationSurvivesNavigationAndRootRebuild(t *testing.T) {
	for _, inactive := range []bool{false, true} {
		t.Run(fmt.Sprint(inactive), func(t *testing.T) {
			c, sh, _, gen := batchFixture(3)
			var root *components.RootListScreen
			r := core.NewRouter(sh, []core.TabEntry{
				{Title: "Repos", New: func(sh *core.Shared) core.Screen { root = NewReposScreen(sh); return root }},
				{Title: "Other", New: func(*core.Shared) core.Screen { return components.NewPicker(nil, components.PickerOpts{}) }},
			})
			step := func(msg tea.Msg) tea.Cmd { m, cmd := r.Update(msg); r = m.(core.Router); return cmd }
			first := c.Repos[0]
			first.Dirty = true
			tick := queuedPublication(t, core.Async(step(core.PropagateAll(repoDescribedMsg{gen: gen, repo: first}))))
			if inactive {
				step(core.ShowTab("Other"))
			} else {
				step(core.Push(components.NewPicker(nil, components.PickerOpts{})))
			}
			before := root
			step(core.PropagateAll(core.MsgThemeChanged{}))
			if root == before || root.List().Items()[0].(components.Item).Pending {
				t.Fatal("rebuilt root did not publish the latest cached result")
			}
			items := root.List().Items()
			step(core.PropagateAll(tick))
			if &root.List().Items()[0] != &items[0] {
				t.Fatal("old root's timer republished the new root")
			}
			second := c.Repos[1]
			second.Dirty = true
			next := queuedPublication(t, core.Async(step(core.PropagateAll(repoDescribedMsg{gen: gen, repo: second}))))
			step(core.PropagateAll(tick))
			step(core.PropagateAll(next))
			if root.List().Items()[1].(components.Item).Pending {
				t.Fatal("new batch failed to publish to a covered or inactive root")
			}
		})
	}
}

func TestImmediateSortRetiresPendingBatch(t *testing.T) {
	c, sh, s, gen := batchFixture(3)
	s.screen.List().Select(1)
	first := c.Repos[0]
	first.Dirty = true
	tick := queuedPublication(t, s.receive(sh, repoDescribedMsg{gen: gen, repo: first}))
	s.onKey(sh, keys.Sort.Keys()[0], nil)
	items := s.screen.List().Items()
	if items[0].(components.Item).Pending || len(c.dirtyRows) != 0 || s.screen.List().SelectedItem().(components.Item).Key != c.Repos[1].Dir {
		t.Fatal("explicit sorting must publish queued changes and keep selection")
	}
	s.receive(sh, tick)
	if &s.screen.List().Items()[0] != &items[0] {
		t.Fatal("retired timer republished after an immediate sort")
	}
}
