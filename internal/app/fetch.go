package app

import (
	"github.com/brohd11/bubblestack/core"
	"github.com/brohd11/gitstack/repo"
	"github.com/brohd11/gitstack/repoui"

	tea "charm.land/bubbletea/v2"
)

// fetchAllCmd fetches every scanned repo (plus the root when it is a checkout) via
// repoui.FetchAllCmd. The repo set is captured on the UI thread.
func fetchAllCmd(sh *core.Shared) tea.Cmd {
	c := Of(sh)
	repos := append([]repo.Repo{}, c.Repos...)
	if c.RootRepo != nil {
		repos = append(repos, *c.RootRepo)
	}
	return repoui.FetchAllCmd(func() []repo.Repo { return repos })
}
