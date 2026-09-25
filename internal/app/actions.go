package app

import (
	"github.com/brohd11/repoview/internal/app/docs"

	"github.com/brohd11/bubblestack/components"
	"github.com/brohd11/bubblestack/core"
)

// actionsMenu is the shared Actions picker opened with "a".
func actionsMenu(sh *core.Shared) *components.PickerScreen {
	return components.NewActionsMenu(selfUpdateHooks(Of(sh).Version),
		"rescan the directory and refresh git state", refreshAction, docs.Pages())
}

// refreshAction rescans the directory; shared by Actions ▸ Refresh and the "r" key.
func refreshAction(sh *core.Shared) core.Action {
	return core.Seq(
		core.PropagateAll(RescanMsg{}),
		core.SetStatus("Refreshed"),
	)
}
