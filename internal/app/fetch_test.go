package app

import (
	"errors"
	"strings"
	"testing"

	"github.com/brohd11/bubblestack/components"
	"github.com/brohd11/bubblestack/core"
	"github.com/brohd11/gitstack/repo"
	"github.com/brohd11/gitstack/repoui"

	tea "charm.land/bubbletea/v2"
)

// Run an empty gather and hold its owned completion in the UI queue. No Git or
// user hooks run, and the guard stays claimed until the test delivers the payload.
func queuedFetchCompletion(t *testing.T, state *repoui.FetchState) repoui.FetchDoneMsg {
	t.Helper()
	cmd, ok := state.Start(func() []repo.Repo { return nil })
	if !ok {
		t.Fatal("fetch already active")
	}
	var done repoui.FetchDoneMsg
	count := 0
	core.DrainBroadcasts(core.Async(cmd), func(payload any) core.Action {
		done = payload.(repoui.FetchDoneMsg)
		count++
		return core.Action{}
	})
	if count != 1 {
		t.Fatalf("got %d completions", count)
	}
	return done
}

func TestFetchGuardSurvivesThemeRebuild(t *testing.T) {
	c := New(twoRepoTree(t), 5, "dev", false)
	sh := core.NewShared(c)
	sh.Chrome = &core.Chrome{Status: components.NewStatusLine()}
	r := core.NewRouter(sh, []core.TabEntry{{Title: "Repos", New: func(sh *core.Shared) core.Screen {
		return NewReposScreen(sh)
	}}})
	step := func(msg tea.Msg) {
		m, _ := r.Update(msg)
		r = m.(core.Router)
	}
	// Leave commands queued: the fetch is still active when the roots are rebuilt.
	root := r.Top()
	step(keyMsg("f"))
	step(core.PropagateAll(core.MsgThemeChanged{}))
	if r.Top() == root {
		t.Fatal("theme change must rebuild the root")
	}
	step(keyMsg("f"))
	if !strings.Contains(sh.Chrome.Status.View(), "already running") {
		t.Fatal("theme rebuild lost the in-flight fetch guard")
	}
}

func TestFetchShortcutWithoutRepos(t *testing.T) {
	c := New(t.TempDir(), 5, "dev", false)
	sh := core.NewShared(c)
	screen := NewReposScreen(sh)
	_, act := screen.Update(sh, keyMsg("f"))
	statuses := core.DrainBroadcasts(act, func(payload any) core.Action {
		t.Fatalf("empty shortcut scheduled a fetch: %T", payload)
		return core.Action{}
	})
	if c.Fetch.Running() || len(statuses) != 1 || statuses[0] != "no repos to fetch" {
		t.Fatalf("empty shortcut claimed a run or changed its message: %q", statuses)
	}
}

type fetchLog struct {
	*components.LogPane
	lines []string
}

func (l *fetchLog) Log(line string, force bool) { l.lines = append(l.lines, line) }

func TestFetchCompletionOwnershipAcrossNavigation(t *testing.T) {
	for _, inactive := range []bool{false, true} {
		name := "covered"
		if inactive {
			name = "inactive"
		}
		t.Run(name, func(t *testing.T) {
			c := New(twoRepoTree(t), 5, "dev", false)
			sh := core.NewShared(c)
			log := &fetchLog{LogPane: components.NewLogPane()}
			sh.Chrome = &core.Chrome{Status: components.NewStatusLine(), Output: log}
			var root *components.RootListScreen
			r := core.NewRouter(sh, []core.TabEntry{
				{Title: "Repos", New: func(sh *core.Shared) core.Screen { root = NewReposScreen(sh); return root }},
				{Title: "Other", New: func(*core.Shared) core.Screen { return components.NewPicker(nil, components.PickerOpts{}) }},
			})
			step := func(msg tea.Msg) { m, _ := r.Update(msg); r = m.(core.Router) }
			first := queuedFetchCompletion(t, &c.Fetch)
			first.Results = []repo.FetchResult{{Name: "owned", Err: errors.New("fetch failed"), Output: []string{"worker output"}}}
			if inactive {
				step(core.ShowTab("Other"))
			} else {
				step(core.Push(components.NewPicker(nil, components.PickerOpts{})))
			}
			oldRoot := root
			step(core.PropagateAll(core.MsgThemeChanged{}))
			if root == oldRoot || !c.Fetch.Running() {
				t.Fatal("rebuild must replace the root and retain fetch ownership")
			}
			var foreign repoui.FetchState
			other := queuedFetchCompletion(t, &foreign)
			reject := func(done repoui.FetchDoneMsg) {
				t.Helper()
				marker, _ := c.Refresh.Start(1)
				sh.WriteStatus("keep status")
				top, currentRoot, lines := r.Top(), root, len(log.lines)
				step(core.PropagateAll(done))
				if !c.Fetch.Running() || c.Refresh.Stale(marker) || r.Top() != top || root != currentRoot ||
					len(log.lines) != lines || !strings.Contains(sh.Chrome.Status.View(), "keep status") {
					t.Fatal("rejected completion changed guard, refresh, navigation, log or status")
				}
			}
			reject(repoui.FetchDoneMsg{Results: first.Results})
			reject(other)
			step(core.PropagateAll(first))
			if c.Fetch.Running() || len(log.lines) != 3 || !c.Refresh.Running() {
				t.Fatalf("owned failure must release, log and refresh: running=%v, log=%q", c.Fetch.Running(), log.lines)
			}
			second := queuedFetchCompletion(t, &c.Fetch)
			reject(first)
			reject(other)
			step(core.PropagateAll(second))
			if c.Fetch.Running() {
				t.Fatal("owned empty completion must release the next run")
			}
			step(core.ShowTab("Repos"))
			step(keyMsg("f"))
			if !c.Fetch.Running() || !strings.Contains(sh.Chrome.Status.View(), "fetching repos") {
				t.Fatal("completion must allow a new shortcut fetch")
			}
		})
	}
}
