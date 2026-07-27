package app

import (
	"context"
	"fmt"
	"time"

	"github.com/brohd11/goutil/selfupdate"

	"github.com/brohd11/bubblestack/components"
	"github.com/brohd11/bubblestack/core"

	tea "github.com/charmbracelet/bubbletea"
)

// selfUpdateCheckTimeout caps the release-check fetch so a slow or unreachable host can hang
// neither the loading screen behind the Actions flow nor the startup check.
const selfUpdateCheckTimeout = 30 * time.Second

// selfUpdateRepo is repoview's own GitHub repo slug, passed to the shared self-update library.
const selfUpdateRepo = "brohd11/repoview"

// selfUpdateInfoMsg carries the self-update check result from the background fetch to
// the loading screen's result handler.
type selfUpdateInfoMsg struct {
	info selfupdate.Info
	err  error
}

// newSelfUpdateLoading is the entry point of the Actions ▸ Update repoview flow:
// loading → confirm → task. It checks repoview's own repo off the UI thread; when an
// update exists it opens the confirm, otherwise it reports "up to date" and pops.
func newSelfUpdateLoading(sh *core.Shared) *components.LoadingScreen {
	version := Of(sh).Version
	cmd := func(parent context.Context) tea.Cmd {
		return func() tea.Msg {
			ctx, cancel := context.WithTimeout(parent, selfUpdateCheckTimeout)
			defer cancel()
			info, err := selfupdate.Check(ctx, selfUpdateRepo, version)
			return selfUpdateInfoMsg{info: info, err: err}
		}
	}
	onResult := func(sh *core.Shared, msg tea.Msg) core.Action {
		m, ok := msg.(selfUpdateInfoMsg)
		if !ok {
			return core.Action{}
		}
		if m.err != nil {
			return core.Seq(core.SetStatusAndLog("update check failed: "+m.err.Error()), core.Pop())
		}
		if !m.info.Available {
			return core.Seq(core.SetStatus("repoview is up to date"), core.Pop())
		}
		return core.Replace(newSelfUpdateConfirm(m.info))
	}
	return components.NewLoadingScreen("Update repoview", "checking for repoview update…", cmd, onResult)
}

// newSelfUpdateConfirm shows the pending update ("current → latest") and, on confirm,
// runs the download+install task.
func newSelfUpdateConfirm(info selfupdate.Info) *components.DialogScreen {
	return components.CreateConfirmScreen(components.ConfirmSimple{
		Crumb: "Update repoview",
		Text:  fmt.Sprintf("Update repoview %s → %s?", info.Current, info.LatestTag),
		OnYes: core.Replace(newSelfUpdateTask(info)),
	})
}

// newSelfUpdateTask downloads and installs the new binary over the running one, then pops
// back to the Actions root. The running process keeps the old code in memory, so it reports
// that a relaunch picks up the new binary.
func newSelfUpdateTask(info selfupdate.Info) *components.TaskScreen {
	run := func(ctx context.Context, sh *core.Shared, report func(string, ...any), done chan<- core.TaskEvent) {
		binDir, err := selfupdate.BinDir()
		if err == nil {
			err = selfupdate.Apply(ctx, selfUpdateRepo, info, binDir, report)
		}
		done <- core.TaskEvent{Done: true, Err: err}
	}
	onDone := func(sh *core.Shared, ev core.TaskEvent) core.Action {
		if ev.Err != nil {
			return core.Seq(
				core.SetStatusAndLog("update failed: "+ev.Err.Error(), true),
				core.Pop(),
			)
		}
		return core.Seq(
			core.SetStatusAndLog(fmt.Sprintf("updated to %s — relaunch repoview to use it", info.LatestTag), true),
			core.Pop(),
		)
	}
	return components.NewTask("updating repoview…", run, onDone)
}

// SelfUpdateCheckCmd is the app-level startup command (wired onto bubblestack Config.Init):
// it checks repoview's own repo for a newer release off the UI thread and, only when an
// update is available, writes an "update available" line to the shared status line and log.
// Anything else (up to date, dev build, fetch error) is silent. The returned Action rides
// back on the cmd's tea.Msg and is applied by the router.
func SelfUpdateCheckCmd(sh *core.Shared) tea.Cmd {
	version := Of(sh).Version
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), selfUpdateCheckTimeout)
		defer cancel()
		info, err := selfupdate.Check(ctx, selfUpdateRepo, version)
		if err != nil || !info.Available {
			return nil
		}
		return core.SetStatusAndLog(
			fmt.Sprintf("repoview update available: %s → %s · Actions ▸ Update repoview", info.Current, info.LatestTag),
			true,
		)
	}
}
