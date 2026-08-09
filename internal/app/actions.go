package app

import (
	"github.com/brohd11/repoview/internal/app/docs"

	"github.com/brohd11/bubblestack/components"
	"github.com/brohd11/bubblestack/core"
)

// actionsMenu is the small Actions picker opened with "a" — the shared bubblestack
// menu (theme, docs, self-update, refresh); repoview only supplies its pages and its
// Refresh row (the rescan the global Refresh key fires).
func actionsMenu(sh *core.Shared) *components.PickerScreen {
	return components.NewActionsMenu(selfUpdateHooks(Of(sh).Version),
		"rescan the directory and refresh git state", refreshAction, docs.Pages())
}

// refreshAction rescans and rebuilds the list — the action both the Actions ▸ Refresh row and
// the global Refresh key ("r") fire. It broadcasts RescanMsg (the screen re-scans on it) and
// sets the status line.
func refreshAction(sh *core.Shared) core.Action {
	return core.Seq(
		core.PropagateAll(RescanMsg{}),
		core.SetStatus("Refreshed"),
	)
}
