package app

import (
	"context"

	"github.com/brohd11/goutil/selfupdate"

	"github.com/brohd11/bubblestack/components"
	"github.com/brohd11/bubblestack/core"

	tea "github.com/charmbracelet/bubbletea"
)

// selfUpdateRepo is repoview's own GitHub repo slug, passed to the shared self-update library.
const selfUpdateRepo = "brohd11/repoview"

// selfUpdateHooks builds the shared self-update flow's (bubblestack/components) hook
// set for repoview: the app name, the running version, and goutil's self-update
// library aimed at repoview's own repo and the running binary's directory. The
// conversion between goutil's selfupdate.Info and the flow's app-agnostic
// SelfUpdateInfo is a direct one — the structs are field-identical by design.
func selfUpdateHooks(version string) components.SelfUpdateHooks {
	return components.SelfUpdateHooks{
		AppName: "repoview",
		Check: func(ctx context.Context) (components.SelfUpdateInfo, error) {
			info, err := selfupdate.Check(ctx, selfUpdateRepo, version)
			return components.SelfUpdateInfo(info), err
		},
		Apply: func(ctx context.Context, info components.SelfUpdateInfo, report func(string, ...any)) error {
			binDir, err := selfupdate.BinDir()
			if err != nil {
				return err
			}
			return selfupdate.Apply(ctx, selfUpdateRepo, selfupdate.Info(info), binDir, report)
		},
	}
}

// newSelfUpdateLoading is the entry point of the Actions ▸ Update repoview flow:
// loading → confirm → task. The flow itself is shared (bubblestack/components);
// repoview only injects its hooks.
func newSelfUpdateLoading(sh *core.Shared) *components.LoadingScreen {
	return components.NewSelfUpdateLoading(selfUpdateHooks(Of(sh).Version))
}

// SelfUpdateCheckCmd is the app-level startup command (wired onto bubblestack Config.Init):
// it checks repoview's own repo for a newer release off the UI thread and, only when an
// update is available, writes an "update available" line to the shared status line and log.
// Anything else (up to date, dev build, fetch error) is silent. The flow and timeout are
// the shared ones in bubblestack/components; only the hooks are repoview's.
func SelfUpdateCheckCmd(sh *core.Shared) tea.Cmd {
	return components.SelfUpdateCheckCmd(selfUpdateHooks(Of(sh).Version))
}
