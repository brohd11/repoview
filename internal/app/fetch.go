package app

import (
	"github.com/brohd11/bubblestack/core"
	"github.com/brohd11/gitstack/repo"

	tea "charm.land/bubbletea/v2"
)

// fetchAllCmd fetches every scanned repo (plus the root when it is a checkout) via
// the app-owned FetchState. The repo set is captured on the UI thread.
func fetchAllCmd(sh *core.Shared) (tea.Cmd, bool) {
	c := Of(sh)
	repos := append([]repo.Repo{}, c.Repos...)
	if c.RootRepo != nil {
		repos = append(repos, *c.RootRepo)
	}
	return c.Fetch.Start(func() []repo.Repo { return repos }, c.GitHookApp())
}
