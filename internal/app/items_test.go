package app

import (
	"testing"

	"github.com/brohd11/bubblestack/components"
	"github.com/brohd11/gitstack/repo"
)

// names extracts the ordered repo names after a sort, for compact assertions.
func names(repos []repo.Repo) []string {
	out := make([]string, len(repos))
	for i, r := range repos {
		out[i] = r.Name
	}
	return out
}

func eq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestSortReposNames proves the two name modes order case-insensitively, ignoring git
// state entirely (a dirty/behind repo still sorts purely by name).
func TestSortReposNames(t *testing.T) {
	fixture := func() []repo.Repo {
		return []repo.Repo{
			{Name: "Beta", Dirty: true},
			{Name: "alpha", Sync: repo.GitSync{Behind: 2}},
			{Name: "gamma"},
		}
	}

	alpha := fixture()
	sortRepos(alpha, components.SortAlpha)
	if got := names(alpha); !eq(got, []string{"alpha", "Beta", "gamma"}) {
		t.Errorf("SortAlpha = %v, want [alpha Beta gamma]", got)
	}

	rev := fixture()
	sortRepos(rev, components.SortReverse)
	if got := names(rev); !eq(got, []string{"gamma", "Beta", "alpha"}) {
		t.Errorf("SortReverse = %v, want [gamma Beta alpha]", got)
	}
}

// TestSortReposStatus proves the status mode groups by attention tier (behind → dirty →
// ahead → clean), with a name tie-break within a tier.
func TestSortReposStatus(t *testing.T) {
	repos := []repo.Repo{
		{Name: "clean"},
		{Name: "ahead", Sync: repo.GitSync{Ahead: 1}},
		{Name: "dirty", Dirty: true},
		{Name: "behind", Sync: repo.GitSync{Behind: 3}},
		{Name: "aclean"}, // clean, sorts before "clean" within the clean tier
	}
	sortRepos(repos, components.SortStatus)
	want := []string{"behind", "dirty", "ahead", "aclean", "clean"}
	if got := names(repos); !eq(got, want) {
		t.Errorf("SortStatus = %v, want %v", got, want)
	}
}

// TestRepoRankMin proves a repo hitting several flags ranks at its most-urgent tier: a
// behind+dirty checkout ranks "behind", not "dirty".
func TestRepoRankMin(t *testing.T) {
	behindDirty := repo.Repo{Sync: repo.GitSync{Behind: 1, Ahead: 4}, Dirty: true}
	if got := repoRank(behindDirty); got != rankBehind {
		t.Errorf("repoRank(behind+dirty+ahead) = %d, want rankBehind (%d)", got, rankBehind)
	}
	if got := repoRank(repo.Repo{}); got != rankClean {
		t.Errorf("repoRank(clean) = %d, want rankClean (%d)", got, rankClean)
	}
}
