package app

import (
	"github.com/brohd11/bubblestack/core"
	"github.com/brohd11/gitstack/repo"
	"github.com/brohd11/gitstack/repoui"

	tea "charm.land/bubbletea/v2"
)

// fetchAllCmd runs `git fetch` concurrently in every scanned repo, off the UI thread, via the
// shared repoui fan-out (repoui.FetchAllCmd owns the timeout, concurrency, and FetchDoneMsg
// broadcast). Every scanned repo is a git checkout (repo.Scan only returns those), so the
// whole set is fetched. The base itself always rides along when it's a checkout — the
// standalone fetch is the one op the root joins unconditionally. The repo set is captured
// here on the UI thread and handed back from the gather closure; nothing touches Shared
// inside the goroutine.
func fetchAllCmd(sh *core.Shared) tea.Cmd {
	c := Of(sh)
	repos := append([]repo.Repo{}, c.Repos...)
	if c.RootRepo != nil {
		repos = append(repos, *c.RootRepo)
	}
	return repoui.FetchAllCmd(func() []repo.Repo { return repos })
}
