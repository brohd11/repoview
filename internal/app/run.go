package app

import (
	"github.com/brohd11/bubblestack"
	"github.com/brohd11/bubblestack/components"
	"github.com/brohd11/bubblestack/core"
	"github.com/brohd11/bubblestack/sysopen"
)

// Run scans root for git repos and launches the repoview TUI.
func Run(root string, depth int, version string) error {
	return bubblestack.Run(bubblestack.Config{
		App:    New(root, depth, version),
		Header: Header,
		// A header click opens the root's own git menu, same as ctrl+v.
		HeaderClick: func(sh *core.Shared, _, _ int) core.Action { return rootGitAction(sh) },
		Output:      components.NewLogPane(),
		Status:      components.NewStatusLine(),
		// Theme left unset — bubblestack.Run applies the shared ~/.bubblestack theme.
		Tabs: []bubblestack.TabEntry{
			{Title: "Repos", New: func(sh *core.Shared) core.Screen { return NewReposScreen(sh) }},
		},
		Init:                 SelfUpdateCheckCmd,
		RefreshAction:        func(sh *core.Shared) core.Action { return refreshAction(sh) },
		TerminalAction:       func(dir string) core.Action { return sysopen.TerminalInlineFor("repoview", dir) },
		TerminalWindowAction: func(dir string) core.Action { return sysopen.Terminal(dir) },
		OpenDirAction:        func(dir string) core.Action { return sysopen.Path(dir, false) },
	})
}
