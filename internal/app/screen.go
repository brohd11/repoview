package app

import (
	"github.com/brohd11/bubblestack/components"
	"github.com/brohd11/bubblestack/core"
	"github.com/brohd11/gitstack/repo"
	"github.com/brohd11/gitstack/repoui"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

const listTitle = "Repos"

// ReposScreen is repoview's single root screen — the scanned repo list. Enter opens the
// highlighted repo's git submenu; V the all-repos menu; f a concurrent fetch-all; a the Actions
// menu. It mirrors gdaddon's Project screen minus all the manifest/install machinery: the whole
// refresh story is "re-scan the directory".
type ReposScreen struct {
	list list.Model
	// sort is the active list ordering, cycled by "s"; the builder reads it and Receive
	// preserves it across a rescan.
	sort components.SortMode
	// fetching guards against a second f fanning out a duplicate set of fetches while the
	// first is still running; cleared when its repoui.FetchDoneMsg arrives.
	fetching bool
}

var _ core.Filterer = (*ReposScreen)(nil)
var _ core.Receiver = (*ReposScreen)(nil)
var _ core.Crumber = (*ReposScreen)(nil)

func NewReposScreen(sh *core.Shared) *ReposScreen {
	l := core.NewSelectList(repoListItems(sh, components.SortAlpha), components.SortTitle(listTitle, components.SortAlpha),
		keys.Git, keys.Diff, keys.Terminal, keys.OpenDir, keys.Fetch, keys.GitAll, keys.RootGit, keys.Actions, keys.Sort)
	return &ReposScreen{list: l}
}

func (s *ReposScreen) Init(*core.Shared) tea.Cmd        { return nil }
func (s *ReposScreen) Filtering() bool                  { return s.list.FilterState() == list.Filtering }
func (s *ReposScreen) View(*core.Shared) string         { return s.list.View() }
func (s *ReposScreen) HelpView(*core.Shared) string     { return core.ShortHelp(s.list, core.HelpTabbed) }
func (s *ReposScreen) SetSize(_ *core.Shared, w, h int) { s.list.SetSize(w, h) }
func (s *ReposScreen) CrumbLabel(bool) string           { return "Repos" }

func (s *ReposScreen) Update(sh *core.Shared, msg tea.Msg) (core.Screen, core.Action) {
	// The tab's own keys, gated behind the filter guard so they don't hijack filter typing.
	if k, ok := msg.(tea.KeyMsg); ok && !s.Filtering() {
		switch {
		// "V" opens the all-repos git page (fetch/pull/push across every scanned repo). The
		// per-repo page is enter on a row (the row's own Pick). It fires when there's anything
		// to act on — nested repos or the base itself — and hands the menu a RootOption so the
		// base can be toggled into the batch's targets.
		case core.MatchKey(k.String(), keys.GitAll):
			if c := Of(sh); !c.HasAny() {
				return s, core.SetStatus("no repos to act on")
			}
			return s, core.Push(repoui.AllReposMenu(sh, allScope(),
				repoui.RootOptionFor(func(sh *core.Shared) *repo.Repo { return Of(sh).RootRepo })))
		// "ctrl+v" opens the root's own git menu — the same RepoMenu a nested repo's row opens,
		// handed the base itself. "V" puts the root in the batch; ctrl+v works it on its own.
		case core.MatchKey(k.String(), keys.RootGit):
			return s, rootGitAction(sh)
		// "f" fetches every repo concurrently so the ahead/behind markers can see new upstream
		// commits. Network-bound, hence explicit.
		case core.MatchKey(k.String(), keys.Fetch):
			if s.fetching {
				return s, core.SetStatus("fetch already running")
			}
			if c := Of(sh); !c.HasAny() {
				return s, core.SetStatus("no repos to fetch")
			}
			s.fetching = true
			return s, core.Seq(
				core.SetStatus("fetching repos…"),
				core.Async(fetchAllCmd(sh)),
			)
		// "a" opens the small Actions menu (theme, refresh).
		case core.MatchKey(k.String(), keys.Actions):
			return s, core.Push(actionsMenu(sh))
		// "s" cycles the sort order (A→Z / Z→A / status), rebuilding the list in place
		// and keeping the cursor on the same repo.
		case core.MatchKey(k.String(), keys.Sort):
			components.CycleSort(&s.list, &s.sort, repoSortModes, listTitle,
				func(m components.SortMode) []list.Item { return repoListItems(sh, m) })
			return s, core.Action{}
		}
	}
	return s, components.RootUpdate(sh, &s.list, msg)
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

// Receive rebuilds the list from a fresh scan on repoview's own RescanMsg (the Refresh
// action / "r") or the shared git flows' repoui.RefreshMsg (raised after a pull/push/commit/
// single fetch) — both mean "the tree changed, re-read it". repoui.FetchDoneMsg additionally
// logs each repo's outcome and summarizes (repoui.LogFetchResults).
func (s *ReposScreen) Receive(sh *core.Shared, payload any) core.Action {
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
func (s *ReposScreen) rescan(sh *core.Shared) core.Action {
	if err := Of(sh).Scan(); err != nil {
		return core.StatusErr(err)
	}
	s.list.SetItems(repoListItems(sh, s.sort))
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
