package app

import (
	"github.com/brohd11/bubblestack/components"
	"github.com/brohd11/bubblestack/core"
	bsupdate "github.com/brohd11/bubblestack/selfupdate"

	tea "github.com/charmbracelet/bubbletea"
)

// selfUpdateRepo is repoview's own GitHub repo slug, passed to the shared self-update library.
const selfUpdateRepo = "brohd11/repoview"

// selfUpdateHooks builds the shared self-update flow's (bubblestack/components) hook
// set for repoview: the app name, the running version, and goutil's self-update library
// aimed at repoview's own repo and the running binary's directory. The wiring lives in
// bubblestack/selfupdate, which owns the (field-identical by design) conversion between
// goutil's selfupdate.Info and the flow's app-agnostic SelfUpdateInfo.
func selfUpdateHooks(version string) components.SelfUpdateHooks {
	return bsupdate.Hooks("repoview", selfUpdateRepo, version)
}

// SelfUpdateCheckCmd is the app-level startup command (wired onto bubblestack Config.Init):
// it checks repoview's own repo for a newer release off the UI thread and, only when an
// update is available, writes an "update available" line to the shared status line and log.
// Anything else (up to date, dev build, fetch error) is silent. The flow and timeout are
// the shared ones in bubblestack/components; only the hooks are repoview's.
func SelfUpdateCheckCmd(sh *core.Shared) tea.Cmd {
	return components.SelfUpdateCheckCmd(selfUpdateHooks(Of(sh).Version))
}
