package app

import (
	"github.com/brohd11/bubblestack/components"
	"github.com/brohd11/bubblestack/core"
	"github.com/brohd11/gitstack/repo"
	"github.com/brohd11/gitstack/repoui"

	"github.com/charmbracelet/bubbles/list"
)

// repoListItems builds the list contents from the last scan: one row per repo, or a single
// inert placeholder when the directory holds no git checkouts.
func repoListItems(sh *core.Shared) []list.Item {
	repos := Of(sh).Repos
	if len(repos) == 0 {
		return []list.Item{components.Item{
			Name: "No git repositories found",
			Desc: "nothing under this directory has a .git — try a different path or -depth",
		}}
	}
	items := make([]list.Item, len(repos))
	for i, r := range repos {
		items[i] = repoRow(r)
	}
	return items
}

// repoRow builds one list row: the repo's base-relative path (plus any warning markers) as the
// name, its branch as the description, enter → the shared per-repo git submenu, and the row's own
// shortcuts (dispatched for the highlighted row by RootUpdate) — "v" the git submenu (an alias of
// enter), "d" that repo's diff list (repoui.DiffAction, seeded beneath the git submenu), and
// "t" a terminal at the repo's directory.
func repoRow(r repo.Repo) components.Item {
	return components.Item{
		Name: r.Name + repo.StatusMarker(r),
		Desc: repoDesc(r),
		Pick: func(sh *core.Shared) core.Action { return core.Push(repoui.RepoMenu(sh, r)) },
		Keys: func(sh *core.Shared, k string) (core.Action, bool) {
			switch {
			case core.MatchKey(k, keys.Git):
				return core.Push(repoui.RepoMenu(sh, r)), true
			case core.MatchKey(k, keys.Diff):
				return repoui.DiffAction(sh, r), true
			case core.MatchKey(k, keys.Terminal):
				return openTerminal(r.Dir), true
			}
			return core.Action{}, false
		},
	}
}

// repoDesc is the row's status line: the checked-out branch (or a note when detached).
func repoDesc(r repo.Repo) string {
	if r.Branch == "" {
		return "⎇ detached HEAD"
	}
	return "⎇ " + r.Branch
}
