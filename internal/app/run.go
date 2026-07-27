package app

import (
	"github.com/brohd11/bubblestack"
	"github.com/brohd11/bubblestack/components"
	"github.com/brohd11/bubblestack/core"
	"github.com/brohd11/bubblestack/sysopen"
)

// Run scans root for git repos and launches the repoview TUI: a single repo-list tab (so
// bubblestack draws no tab strip), the shared git screens reached from it, the persistent
// header, a log/output pane (the git flows stream into it), and a status line. The shared
// ~/.bubblestack theme, if any, is applied by bubblestack.Run; the global Refresh key rescans.
// version is the binary's version string — the Init startup command uses it for a background
// self-update check that notes "update available" on the status line (silent otherwise).
func Run(root string, depth int, version string) error {
	return bubblestack.Run(bubblestack.Config{
		App:    New(root, depth, version),
		Header: Header,
		Output: components.NewLogPane(),
		Status: components.NewStatusLine(),
		// Theme left unset — bubblestack.Run applies the shared ~/.bubblestack theme.
		Tabs: []bubblestack.TabEntry{
			{Title: "Repos", New: func(sh *core.Shared) core.Screen { return NewReposScreen(sh) }},
		},
		Init:           SelfUpdateCheckCmd,
		RefreshAction:  func(sh *core.Shared) core.Action { return refreshAction(sh) },
		TerminalAction: func(dir string) core.Action { return sysopen.Terminal(dir) },
		OpenDirAction:  func(dir string) core.Action { return sysopen.Path(dir, false) },
	})
}
