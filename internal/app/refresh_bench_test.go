package app

import (
	"fmt"
	"testing"

	"github.com/brohd11/bubblestack/components"
	"github.com/brohd11/bubblestack/core"
	"github.com/brohd11/gitstack/repo"
)

// Replay CPU-only results; discovery and Git are deliberately outside this benchmark.
func BenchmarkRefreshList(b *testing.B) {
	for _, n := range []int{100, 1000, 10000} {
		for _, filtered := range []bool{false, true} {
			b.Run(fmt.Sprintf("n=%d/filter=%v", n, filtered), func(b *testing.B) {
				c := &Ctx{Repos: make([]repo.Repo, n)}
				for i := range c.Repos {
					c.Repos[i] = repo.Repo{Name: fmt.Sprintf("repo%05d", i), Dir: fmt.Sprintf("/repo/%d", i)}
				}
				sh := core.NewShared(c)
				s := &reposState{screen: components.NewRootList(nil, components.RootListOpts{})}
				s.screen.SetSize(sh, 100, 30)
				if filtered {
					s.screen.List().SetFilterText("repo")
				}
				b.ReportAllocs()
				publications, built := 0, 0
				b.ResetTimer()
				for range b.N {
					c.indexRepos()
					c.dirtyRows = make(map[string]bool)
					c.Pending = make(map[string]bool, n)
					for _, r := range c.Repos {
						c.Pending[r.Dir] = true
					}
					gen, _ := c.Refresh.Start(n)
					s.rebuild(sh)
					publications++
					built += len(c.Repos)
					for i, r := range c.Repos {
						r.Dirty = true
						s.receive(sh, repoDescribedMsg{gen: gen, repo: r})
						if !c.Refresh.Running() {
							publications++
							built += len(c.Repos)
						}
						// Replay ten delivery windows without waiting for real timers.
						if (i+1)%(n/10) == 0 && i+1 < n {
							publications++
							built += len(c.dirtyRows)
							s.publish(sh)
						}
					}
				}
				b.ReportMetric(float64(publications)/float64(b.N), "publications/op")
				b.ReportMetric(float64(built)/float64(b.N), "rows/op")
			})
		}
	}
}
