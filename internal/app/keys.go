package app

import "github.com/charmbracelet/bubbles/key"

// keys are repoview's screen-level bindings that aren't part of bubblestack's framework keymap
// (core.Keys). Enter — open the highlighted repo's git menu — is the list's own select key, so
// it isn't here; these are the extras the repo list advertises and matches on.
var keys = struct {
	Git      key.Binding // open the highlighted repo's git menu (alias of enter)
	Terminal key.Binding // open an OS terminal at the highlighted repo's directory
	GitAll   key.Binding // open the all-repos git menu (fetch/pull/push across every repo)
	Fetch    key.Binding // concurrent fetch-all, refreshing ahead/behind
	Actions  key.Binding // open the Actions menu (theme, refresh)
}{
	// v/t are row-level (dispatched via the highlighted row's Item.Keys). v = version control,
	// mirroring gdaddon — not g/G, which bubbles reserves for jump-to-top/bottom on every list.
	Git:      key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "git")),
	Terminal: key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "terminal")),
	GitAll:   key.NewBinding(key.WithKeys("V"), key.WithHelp("V", "git all")),
	Fetch:    key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "fetch all")),
	Actions:  key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "actions")),
}
