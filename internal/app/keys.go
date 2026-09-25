package app

import (
	"charm.land/bubbles/v2/key"

	"github.com/brohd11/bubblestack/core"
)

// keys are repoview's screen-level bindings beyond core.Keys.
var keys = struct {
	Git            key.Binding // open the highlighted repo's git menu (alias of enter)
	Diff           key.Binding // open the highlighted repo's diff list, skipping the git menu
	Terminal       key.Binding // open a terminal in this process at the highlighted repo's directory
	TerminalWindow key.Binding // open a detached terminal window at the highlighted repo's directory
	OpenDir        key.Binding // open the highlighted repo's directory in the OS file manager
	GitAll         key.Binding // open the all-repos git menu (fetch/pull/push across every repo)
	RootGit        key.Binding // open the scanned root's own git menu (the base directory itself)
	Fetch          key.Binding // concurrent fetch-all, refreshing ahead/behind
	Actions        key.Binding // open the Actions menu (theme, refresh)
	Sort           key.Binding // cycle the repo list's sort order (A→Z / Z→A / status)
}{
	// Row-level keys. v, not g/G, which lists use for top/bottom.
	Git:            key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "git")),
	Diff:           key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "diff")),
	Terminal:       key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "terminal")),
	TerminalWindow: key.NewBinding(key.WithKeys("T"), key.WithHelp("T", "term window")),
	OpenDir:        key.NewBinding(key.WithKeys("ctrl+t"), key.WithHelp("ctrl+t", "open dir")),
	GitAll:         key.NewBinding(key.WithKeys("V"), key.WithHelp("V", "git all")),
	RootGit:        key.NewBinding(key.WithKeys("ctrl+v"), key.WithHelp("ctrl+v", "root git")),
	Fetch:          key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "fetch all")),
	// The shared binding, so the ctrl+alt+a alias reaches repoview too (core.Keys).
	Actions: core.Keys.Actions,
	Sort:    key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "sort")),
}
