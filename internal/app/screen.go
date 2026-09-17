package app

import (
	"github.com/brohd11/bubblestack/components"
	"github.com/brohd11/bubblestack/core"
	"github.com/brohd11/gitstack/repo"
	"github.com/brohd11/gitstack/repoui"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
)

const listTitle = "Repos"

// reposState owns sorting and in-flight work; RootListScreen owns list interaction.
type reposState struct {
	screen   *components.RootListScreen
	sort     components.SortMode
	fetching bool
}

func NewReposScreen(sh *core.Shared) *components.RootListScreen {
	s := &reposState{}
	s.screen = components.NewRootList(repoListItems(sh, s.sort), components.RootListOpts{
		PickerOpts: components.PickerOpts{
			Title: components.SortTitle(listTitle, s.sort),
			Crumb: "Repos",
			Help:  []key.Binding{keys.Git, keys.Diff, keys.Terminal, keys.TerminalWindow, keys.OpenDir, keys.Fetch, keys.GitAll, keys.RootGit, keys.Actions, keys.Sort},
			OnKey: s.onKey,
		},
		Receive: s.receive,
	})
	return s.screen
}

func (s *reposState) onKey(sh *core.Shared, k string, _ list.Item) (core.Action, bool) {
	switch {
	case core.MatchKey(k, keys.GitAll):
		if c := Of(sh); !c.HasAny() {
			return core.SetStatus("no repos to act on"), true
		}
		return core.Push(repoui.AllReposMenu(sh, allScope(),
			repoui.RootOptionFor(func(sh *core.Shared) *repo.Repo { return Of(sh).RootRepo }))), true
	case core.MatchKey(k, keys.RootGit):
		return rootGitAction(sh), true
	case core.MatchKey(k, keys.Fetch):
		if s.fetching {
			return core.SetStatus("fetch already running"), true
		}
		if c := Of(sh); !c.HasAny() {
			return core.SetStatus("no repos to fetch"), true
		}
		s.fetching = true
		return core.Seq(core.SetStatus("fetching repos…"), core.Async(fetchAllCmd(sh))), true
	case core.MatchKey(k, keys.Actions):
		return core.Push(actionsMenu(sh)), true
	case core.MatchKey(k, keys.Sort):
		components.CycleSort(s.screen.List(), &s.sort, repoSortModes, listTitle,
			func(m components.SortMode) []list.Item { return repoListItems(sh, m) })
		return core.Action{}, true
	}
	return core.Action{}, false
}

// rootGitAction opens the scanned root's own git menu — the ctrl+v key and a header
// click both resolve to it. A non-checkout base only reports on the status line.
func rootGitAction(sh *core.Shared) core.Action {
	c := Of(sh)
	if c.RootRepo == nil {
		return core.SetStatus("base directory is not a git checkout")
	}
	return core.Push(repoui.RepoMenu(sh, *c.RootRepo, c.RootRepo.Name))
}

// receive rebuilds the list from a fresh scan on repoview's own RescanMsg (the Refresh
// action / "r") or the shared git flows' repoui.RefreshMsg (raised after a pull/push/commit/
// single fetch) — both mean "the tree changed, re-read it". repoui.FetchDoneMsg additionally
// logs each repo's outcome and summarizes (repoui.LogFetchResults).
func (s *reposState) receive(sh *core.Shared, payload any) core.Action {
	switch p := payload.(type) {
	case repoui.RefreshMsg, RescanMsg:
		return s.rescan(sh)
	case repoui.FetchDoneMsg:
		s.fetching = false
		return core.Seq(
			s.rescan(sh),
			repoui.LogFetchResults(sh, p.Results, "repo(s)", "no repos fetched"),
		)
	}
	return core.Action{}
}

// rescan re-reads the tree and rebuilds the list from it. A scan failure keeps the old list
// (Ctx.Scan leaves it intact) and says so on the status line, rather than letting the stale
// list look current.
func (s *reposState) rescan(sh *core.Shared) core.Action {
	if err := Of(sh).Scan(); err != nil {
		return core.StatusErr(err)
	}
	s.screen.SetItems(repoListItems(sh, s.sort))
	return core.Action{}
}

// allScope is the single "repos" scope handed to the all-repos menu: every scanned repo, read
// fresh. With one scope the menu shows no scope-cycle row (gated in repoui by len(scopes) > 1).
func allScope() []repoui.Scope {
	return []repoui.Scope{{
		Label: "repos",
		Repos: func(sh *core.Shared) []repo.Repo { return Of(sh).Repos },
	}}
}
