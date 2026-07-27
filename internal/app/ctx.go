package app

import (
	"github.com/brohd11/bubblestack/core"
	"github.com/brohd11/gitstack/repo"
)

// Ctx is repoview's app context, stored on core.Shared.App and recovered with Of. It holds the
// scan root/depth and the repos found by the last scan. There is no manifest — the list is
// simply whatever a fresh scan of Root turns up, which is the whole point of the tool.
type Ctx struct {
	Root  string
	Depth int
	Repos []repo.Repo
	// RootRepo is the scanned base itself, when it is a git checkout (nil otherwise). It never
	// rides Repos — the list is nested checkouts only. The header reads it to show the root's
	// own status marker, fetch-all appends it to the fetch set, and the all-repos menu offers
	// it via its include-root toggle.
	RootRepo *repo.Repo
}

// New builds the context and performs the initial scan, so the first screen has rows to show.
func New(root string, depth int) *Ctx {
	c := &Ctx{Root: root, Depth: depth}
	c.Scan()
	return c
}

// Of recovers the repoview context from a Shared. Screens call c := app.Of(sh).
func Of(sh *core.Shared) *Ctx { return core.App[Ctx](sh) }

// Scan re-reads every git checkout under Root — branch, upstream divergence, and dirty state
// per repo (all local reads). A scan error leaves the previous list intact rather than blanking
// the screen.
func (c *Ctx) Scan() {
	repos, err := repo.Scan(c.Root, c.Depth)
	if err != nil {
		return // leave the previous list intact
	}
	c.Repos = repos
	// Scan omits the base itself; describe it separately so the header can show its status and
	// fetch-all / the batch menu can opt it in, without it ever appearing as a list row.
	c.RootRepo = nil
	if root, ok := repo.DescribeRoot(c.Root); ok {
		c.RootRepo = &root
	}
}

// RescanMsg is repoview's "reload yourself" broadcast: the repo list re-scans from disk on it.
// The Refresh action and the global Refresh key ('r') raise it; the shared git flows raise
// repoui.RefreshMsg after an op, which the screen treats the same way.
type RescanMsg struct{}

// Receive handles app-level broadcasts. On a theme change it rebuilds the cached root so it
// re-bakes its list/delegate styles from the new palette (core.OnThemeChange; the
// router-drawn chrome repaints on its own). Everything else is handled by the screen.
func (c *Ctx) Receive(sh *core.Shared, payload any) core.Action {
	return core.OnThemeChange(payload)
}
