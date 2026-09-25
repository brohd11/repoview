package app

import (
	"github.com/brohd11/bubblestack/core"
	"github.com/brohd11/gitstack/repo"
	"github.com/brohd11/gitstack/repoui"
)

// Ctx is repoview's app context: the scan root/depth and the repos found by the last scan.
type Ctx struct {
	// ListCompact is the session density shared by standard roots and pickers.
	ListCompact bool

	Root    string
	Depth   int
	Version string
	Repos   []repo.Repo
	// RootRepo is the scanned base when it is itself a checkout (nil otherwise). It is not in
	// Repos.
	RootRepo *repo.Repo
}

// New builds the context and performs the initial scan, so the first screen has rows to show.
// version is the binary's version string, used by the self-update check.
func New(root string, depth int, version string) *Ctx {
	c := &Ctx{Root: root, Depth: depth, Version: version}
	c.Scan()
	return c
}

// Of recovers the repoview context from a Shared. Screens call c := app.Of(sh).
func Of(sh *core.Shared) *Ctx { return core.App[Ctx](sh) }

// Scan re-reads every checkout under Root (local reads only). On error the previous list is
// kept and the error returned.
func (c *Ctx) Scan() error {
	repos, err := repo.Scan(c.Root, c.Depth)
	if err != nil {
		return err // leave the previous list intact
	}
	c.Repos = repos
	// Scan omits the base itself; describe it separately so the header can show its status and
	// fetch-all / the batch menu can opt it in, without it ever appearing as a list row.
	c.RootRepo = nil
	if root, ok := repo.DescribeRoot(c.Root); ok {
		c.RootRepo = &root
	}
	return nil
}

// HasAny reports whether the last scan found anything a git op can act on — a nested repo
// or the root checkout itself.
func (c *Ctx) HasAny() bool { return len(c.Repos) > 0 || c.RootRepo != nil }

// RefreshRepo updates a known checkout and its enclosing repos without scanning
// the tree. An operation on the root itself can change the tree, so it rescans.
func (c *Ctx) RefreshRepo(msg repoui.RepoRefreshMsg) error {
	if msg.Targets(c.Root) {
		return c.Scan()
	}
	known := false
	for _, r := range c.Repos {
		known = known || msg.Targets(r.Dir)
	}
	if !known {
		return nil
	}
	for i, r := range c.Repos {
		if msg.Affects(r.Dir) {
			c.Repos[i] = repo.Describe(r.Name, r.Dir)
		}
	}
	if c.RootRepo != nil && msg.Affects(c.RootRepo.Dir) {
		root, ok := repo.DescribeRoot(c.Root)
		c.RootRepo = nil
		if ok {
			c.RootRepo = &root
		}
	}
	return nil
}

// RescanMsg makes the repo list re-scan from disk.
type RescanMsg struct{}

// Receive rebuilds the cached root on a theme change.
func (c *Ctx) Receive(sh *core.Shared, payload any) core.Action {
	return core.OnThemeChange(payload)
}

// ListDensity opts standard lists into the app-wide session preference.
func (c *Ctx) ListDensity() *bool { return &c.ListCompact }
