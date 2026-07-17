package app

import (
	"fmt"
	"strings"

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

// diffViaGitMenu decides what "d" leaves beneath the diff picker.
//
// When true, "d" seeds the git submenu under the picker — the same stack pressing "v" and
// choosing Diff would build. The trail reads "Repos › Git › Diff", and since esc is a
// one-level pop, backing out of a diff lands on the hub that has Commit. That's the point:
// reading a diff is the step before committing, and the alternative makes you esc to the list
// and re-enter via "v" to reach the action the diff just argued for.
//
// When false, "d" pushes the picker alone ("Repos › Diff"): esc returns straight to the list,
// and the git state RepoMenu reads on build is never computed.
const diffViaGitMenu = true

// diffAction is what "d" does for one repo row — see diffViaGitMenu.
func diffAction(sh *core.Shared, r repo.Repo) core.Action {
	if !diffViaGitMenu {
		return core.Push(repoui.DiffMenu(sh, r))
	}
	return core.Seq(
		core.Push(repoui.RepoMenu(sh, r)),
		core.Push(repoui.DiffMenu(sh, r)),
	)
}

// repoRow builds one list row: the repo's base-relative path (plus any warning markers) as the
// name, its branch as the description, enter → the shared per-repo git submenu, and the row's own
// shortcuts (dispatched for the highlighted row by RootUpdate) — "v" the git submenu (an alias of
// enter), "d" that repo's diff list (see diffViaGitMenu for what it opens beneath it), and "t" a
// terminal at the repo's directory.
func repoRow(r repo.Repo) components.Item {
	return components.Item{
		Name: r.Name + rowMarker(r),
		Desc: repoDesc(r),
		Pick: func(sh *core.Shared) core.Action { return core.Push(repoui.RepoMenu(sh, r)) },
		Keys: func(sh *core.Shared, k string) (core.Action, bool) {
			switch {
			case core.MatchKey(k, keys.Git):
				return core.Push(repoui.RepoMenu(sh, r)), true
			case core.MatchKey(k, keys.Diff):
				return diffAction(sh, r), true
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

// rowMarker appends the warning suffix from a repo's git state — behind / ahead / dirty —
// mirroring gdaddon's phrasing. Empty when the repo is in sync with its upstream and clean.
// The ahead/behind counts are as fresh as the last fetch (f fetches, then the markers settle).
func rowMarker(r repo.Repo) string {
	var parts []string
	if r.Sync.Behind > 0 {
		parts = append(parts, fmt.Sprintf("behind origin %d", r.Sync.Behind))
	}
	if r.Sync.Ahead > 0 {
		parts = append(parts, fmt.Sprintf("ahead %d", r.Sync.Ahead))
	}
	if r.Dirty {
		parts = append(parts, "uncommitted changes")
	}
	if len(parts) == 0 {
		return ""
	}
	return "  ⚠ [" + strings.Join(parts, " / ") + "]"
}
