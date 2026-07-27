package app

import "github.com/charmbracelet/bubbles/key"

// keys are repoview's screen-level bindings that aren't part of bubblestack's framework keymap
// (core.Keys). Enter — open the highlighted repo's git menu — is the list's own select key, so
// it isn't here; these are the extras the repo list advertises and matches on.
var keys = struct {
	Git      key.Binding // open the highlighted repo's git menu (alias of enter)
	Diff     key.Binding // open the highlighted repo's diff list, skipping the git menu
	Terminal key.Binding // open an OS terminal at the highlighted repo's directory
	OpenDir  key.Binding // open the highlighted repo's directory in the OS file manager
	GitAll   key.Binding // open the all-repos git menu (fetch/pull/push across every repo)
	RootGit  key.Binding // open the scanned root's own git menu (the base directory itself)
	Fetch    key.Binding // concurrent fetch-all, refreshing ahead/behind
	Actions  key.Binding // open the Actions menu (theme, refresh)
	Sort     key.Binding // cycle the repo list's sort order (A→Z / Z→A / status)
}{
	// v/d/t are row-level (dispatched via the highlighted row's Item.Keys). v = version control,
	// mirroring gdaddon — not g/G, which bubbles reserves for jump-to-top/bottom on every list.
	Git:      key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "git")),
	Diff:     key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "diff")),
	Terminal: key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "terminal")),
	OpenDir:  key.NewBinding(key.WithKeys("T"), key.WithHelp("T", "open dir")),
	GitAll:   key.NewBinding(key.WithKeys("V"), key.WithHelp("V", "git all")),
	RootGit:  key.NewBinding(key.WithKeys("ctrl+v"), key.WithHelp("ctrl+v", "root git")),
	Fetch:    key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "fetch all")),
	Actions:  key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "actions")),
	Sort:     key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "sort")),
}
