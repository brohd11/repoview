package app

import (
	"github.com/brohd11/repoview/internal/app/docs"

	"github.com/brohd11/bubblestack/components"
	"github.com/brohd11/bubblestack/core"
)

// actionsMenu is the small Actions picker opened with "a". The menu itself is the shared
// bubblestack one (theme, self-update, refresh); repoview's only addition is the Docs row.
func actionsMenu(sh *core.Shared) *components.PickerScreen {
	docsRow := components.Item{
		Name: "? Docs",
		Desc: "getting started, controls, git menu",
		Pick: func(sh *core.Shared) core.Action { return core.Push(docs.Index()) },
	}
	return components.NewActionsMenu(selfUpdateHooks(Of(sh).Version),
		"rescan the directory and refresh git state", refreshAction, docsRow)
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
