package app

import (
	"context"
	"time"

	"github.com/brohd11/bubblestack/core"
	"github.com/brohd11/gitstack/repo"

	tea "github.com/charmbracelet/bubbletea"
)

// fetchDone carries a fetch-all pass's results back to the root via a broadcast, where they're
// logged and the list is rebuilt from the freshly updated refs.
type fetchDone struct {
	results []repo.FetchResult
}

// fetchTimeout caps the whole fan-out so an unreachable remote can't leave the fetch pending
// (and the root stuck marked as fetching) forever.
const fetchTimeout = 90 * time.Second

// fetchAllCmd runs `git fetch` concurrently in every scanned repo, off the UI thread. The repo
// set is captured up front and nothing touches Shared inside the goroutine; the results ride
// back on the broadcast. Every scanned repo is a git checkout (repo.Scan only returns those),
// so the whole set is passed straight to repo.FetchAll. The base itself always rides along when
// it's a checkout — the standalone fetch is the one op the root joins unconditionally.
func fetchAllCmd(sh *core.Shared) tea.Cmd {
	repos := Of(sh).Repos
	if root := Of(sh).RootRepo; root != nil {
		repos = append(repos, *root)
	}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
		defer cancel()
		return core.PropagateAll(fetchDone{results: repo.FetchAll(ctx, repos)})
	}
}
