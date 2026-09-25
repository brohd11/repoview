package app

import (
	"sort"
	"strings"

	"github.com/brohd11/bubblestack/components"
	"github.com/brohd11/bubblestack/core"
	"github.com/brohd11/bubblestack/sysopen"
	"github.com/brohd11/gitstack/repo"
	"github.com/brohd11/gitstack/repoui"

	"charm.land/bubbles/v2/list"
)

// repoSortModes is the repo list's sort cycle: A→Z, Z→A, then status (attention-worthy
// repos first). Passed to components.CycleSort by the root's "s" callback.
var repoSortModes = []components.SortMode{components.SortAlpha, components.SortReverse, components.SortStatus}

// repoListItems builds the list contents from the last scan, ordered per mode: one row
// per repo, or a single inert placeholder when the directory holds no git checkouts.
func repoListItems(sh *core.Shared, mode components.SortMode) []list.Item {
	repos := Of(sh).Repos
	if len(repos) == 0 {
		return []list.Item{components.Item{
			Name: "No git repositories found",
			Desc: "nothing under this directory has a .git — try a different path or -depth",
		}}
	}
	sorted := make([]repo.Repo, len(repos))
	copy(sorted, repos)
	sortRepos(sorted, mode)
	items := make([]list.Item, len(sorted))
	for i, r := range sorted {
		items[i] = repoRow(r)
	}
	return items
}

// sortRepos sorts repos in place by name or by repoRank, sorting repo values rather than the
// marker-suffixed rows.
func sortRepos(repos []repo.Repo, mode components.SortMode) {
	name := func(i int) string { return strings.ToLower(repos[i].Name) }
	switch mode {
	case components.SortReverse:
		sort.SliceStable(repos, func(i, j int) bool { return name(i) > name(j) })
	case components.SortStatus:
		sort.SliceStable(repos, func(i, j int) bool {
			ri, rj := repoRank(repos[i]), repoRank(repos[j])
			if ri != rj {
				return ri < rj
			}
			return name(i) < name(j)
		})
	default: // SortAlpha
		sort.SliceStable(repos, func(i, j int) bool { return name(i) < name(j) })
	}
}

// Attention tiers for SortStatus, most urgent first.
const (
	rankBehind = iota // behind its upstream
	rankDirty         // uncommitted changes
	rankAhead         // unpushed local commits
	rankClean         // nothing to report
)

// repoRank scores a repo for the status sort. Any flag can only raise urgency (take the
// minimum), so a behind+dirty repo ranks "behind".
func repoRank(r repo.Repo) int {
	rank := rankClean
	if r.Sync.Ahead > 0 && rankAhead < rank {
		rank = rankAhead
	}
	if r.Dirty && rankDirty < rank {
		rank = rankDirty
	}
	if r.Sync.Behind > 0 && rankBehind < rank {
		rank = rankBehind
	}
	return rank
}

// repoRow builds one row: relative path (with markers) and branch. Enter and v open the git
// menu, d the diff list, t/T a terminal.
func repoRow(r repo.Repo) components.Item {
	return components.Item{
		Name: r.Name + repo.StatusMarker(r),
		Desc: repoDesc(r),
		Pick: func(sh *core.Shared) core.Action { return core.Push(repoui.RepoMenu(sh, r, r.Name)) },
		Keys: func(sh *core.Shared, k string) (core.Action, bool) {
			switch {
			case core.MatchKey(k, keys.Git):
				return core.Push(repoui.RepoMenu(sh, r, r.Name)), true
			case core.MatchKey(k, keys.Diff):
				return repoui.DiffAction(sh, r, r.Name), true
			case core.MatchKey(k, keys.Terminal):
				return sysopen.TerminalInlineFor("repoview", r.Dir), true
			case core.MatchKey(k, keys.TerminalWindow):
				return sysopen.Terminal(r.Dir), true
			case core.MatchKey(k, keys.OpenDir):
				return sysopen.Path(r.Dir, false), true
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
