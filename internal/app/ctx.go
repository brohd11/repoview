package app

import (
	"os"
	"path/filepath"

	"github.com/brohd11/bubblestack/components"
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

	// Pending marks checkouts (by Dir, the root included) whose state the running Refresh is
	// re-reading; their rows show the cached state muted until their result lands.
	Pending map[string]bool
	// Refresh tracks the async describe pass. It lives here, not on the screen, because the
	// root is rebuilt on theme changes while results are in flight.
	Refresh components.Refresh
	// rankSnap freezes each repo's status-sort rank for the running refresh, so rows do not
	// jump under the cursor as results land (see repoListItems).
	rankSnap map[string]int
}

// New builds the context and discovers the checkouts (a directory walk, no git), so the
// first screen has rows to show. Their state is read by the refresh Run's Init starts.
// version is the binary's version string, used by the self-update check.
func New(root string, depth int, version string) *Ctx {
	c := &Ctx{Root: root, Depth: depth, Version: version}
	c.Discover()
	return c
}

// Of recovers the repoview context from a Shared. Screens call c := app.Of(sh).
func Of(sh *core.Shared) *Ctx { return core.App[Ctx](sh) }

// Discover re-walks Root for checkouts without running git: known repos keep their cached
// state, new ones start empty, vanished ones drop, and every one (the root included) is
// marked Pending for StartRefresh to describe. On error the previous list is kept.
func (c *Ctx) Discover() error {
	rels, err := repo.FindGitRepos(c.Root, c.Depth)
	if err != nil {
		return err
	}
	known := make(map[string]repo.Repo, len(c.Repos))
	for _, r := range c.Repos {
		known[r.Dir] = r
	}
	c.Pending = make(map[string]bool, len(rels)+1)
	repos := make([]repo.Repo, 0, len(rels))
	for _, rel := range rels {
		dir := filepath.Join(c.Root, rel)
		r, ok := known[dir]
		if !ok {
			r = repo.Repo{Name: rel, Dir: dir}
		}
		repos = append(repos, r)
		c.Pending[dir] = true
	}
	c.Repos = repos
	if _, err := os.Stat(filepath.Join(c.Root, ".git")); err != nil {
		c.RootRepo = nil
	} else {
		if c.RootRepo == nil {
			c.RootRepo = &repo.Repo{Name: filepath.Base(c.Root), Dir: c.Root, Root: true}
		}
		c.Pending[c.Root] = true
	}
	return nil
}

// Scan re-reads every checkout under Root synchronously (local reads only), for callers
// that need the state now; the UI refreshes through StartRefresh. On error the previous
// list is kept and the error returned.
func (c *Ctx) Scan() error {
	repos, err := repo.Scan(c.Root, c.Depth)
	if err != nil {
		return err // leave the previous list intact
	}
	c.Repos = repos
	c.Pending = nil
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

// RefreshRepo updates a known checkout and its enclosing repos without scanning the tree.
// An operation on the root itself can change the tree, so it reports full instead and the
// caller starts a refresh.
func (c *Ctx) RefreshRepo(msg repoui.RepoRefreshMsg) (full bool) {
	if msg.Targets(c.Root) {
		return true
	}
	known := false
	for _, r := range c.Repos {
		known = known || msg.Targets(r.Dir)
	}
	if !known {
		return false
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
	return false
}

// RescanMsg makes the repo list re-scan from disk.
type RescanMsg struct{}

// Receive rebuilds the cached root on a theme change.
func (c *Ctx) Receive(sh *core.Shared, payload any) core.Action {
	return core.OnThemeChange(payload)
}

// ListDensity opts standard lists into the app-wide session preference.
func (c *Ctx) ListDensity() *bool { return &c.ListCompact }
