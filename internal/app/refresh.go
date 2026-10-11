package app

import (
	"github.com/brohd11/bubblestack/core"
	"github.com/brohd11/gitstack/repo"

	tea "charm.land/bubbletea/v2"
)

// maxDescribe caps concurrent describes (three git processes each).
const maxDescribe = 8

// repoDescribedMsg is one checkout's fresh state from refresh pass gen, broadcast so it
// reaches the repos root whatever is pushed over it.
type repoDescribedMsg struct {
	gen  int64
	repo repo.Repo
}

// StartRefresh re-discovers the tree (fast, no git) so rows show muted at once, then
// describes every checkout off the UI thread, each result landing on its own. A pass
// already running is superseded; its late results are dropped.
func (c *Ctx) StartRefresh() core.Action {
	if err := c.Discover(); err != nil {
		return core.StatusErr(err)
	}
	targets := append([]repo.Repo{}, c.Repos...)
	if c.RootRepo != nil {
		targets = append(targets, *c.RootRepo)
	}
	c.rankSnap = make(map[string]int, len(c.Repos))
	for _, r := range c.Repos {
		c.rankSnap[r.Dir] = repoRank(r)
	}
	gen, act := c.Refresh.Start(len(targets))
	cmds := make([]tea.Cmd, len(targets))
	for i, r := range targets {
		cmds[i] = c.describeCmd(gen, r)
	}
	return core.Seq(act, core.Async(core.BoundedBatch(maxDescribe, cmds...)))
}

// describeCmd reads one checkout's state (local git reads) off the UI thread, skipping the
// work once a newer pass has superseded gen.
func (c *Ctx) describeCmd(gen int64, r repo.Repo) tea.Cmd {
	name, dir, root := r.Name, r.Dir, r.Root
	return func() tea.Msg {
		if c.Refresh.Stale(gen) {
			return nil
		}
		d := repo.Describe(name, dir)
		d.Root = root
		return core.PropagateAll(repoDescribedMsg{gen: gen, repo: d})
	}
}

// applyDescribed stores one result and clears its pending mark. ok is false for a stale
// pass's result; act is the final status once the pass completes.
func (c *Ctx) applyDescribed(m repoDescribedMsg) (ok bool, act core.Action) {
	ok, act = c.Refresh.Land(m.gen)
	if !ok {
		return false, act
	}
	delete(c.Pending, m.repo.Dir)
	if m.repo.Root {
		if c.RootRepo != nil {
			r := m.repo
			c.RootRepo = &r
		}
	} else {
		for i := range c.Repos {
			if c.Repos[i].Dir == m.repo.Dir {
				c.Repos[i] = m.repo
			}
		}
	}
	if !c.Refresh.Running() {
		c.rankSnap = nil
	}
	return true, act
}
