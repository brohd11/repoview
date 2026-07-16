package app

import (
	"os"
	"os/exec"
	"runtime"

	"github.com/brohd11/bubblestack/core"

	tea "github.com/charmbracelet/bubbletea"
)

// openTerminal opens an OS terminal at dir. It mirrors gdaddon's sysopen.Terminal — repoview
// can't import that package (separate repo), and it's the only bit of it repoview needs, so it's
// a small local copy. darwin/windows shell out to a known terminal; linux probes for a common
// emulator and reports a status when none is found.
func openTerminal(dir string) core.Action {
	if _, err := os.Stat(dir); err != nil {
		return core.SetStatusAndLog("path not found: " + dir)
	}
	cmd := terminalCmd(dir)
	if cmd == nil {
		return core.SetStatusAndLog("no terminal emulator found")
	}
	return core.Seq(
		core.SetStatus("opening terminal at "+dir),
		core.Async(func() tea.Msg {
			_ = cmd.Start()
			return nil
		}),
	)
}

// terminalCmd builds the terminal launch command for the current OS, or returns nil when no
// suitable terminal could be found (linux with no known emulator on PATH).
func terminalCmd(dir string) *exec.Cmd {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", "-a", "Terminal", dir)
	case "windows":
		return exec.Command("cmd", "/c", "start", "cmd", "/k", "cd /d "+dir)
	default:
		for _, t := range []struct {
			bin  string
			args []string
		}{
			{"x-terminal-emulator", []string{"--working-directory=" + dir}},
			{"gnome-terminal", []string{"--working-directory=" + dir}},
			{"konsole", []string{"--workdir", dir}},
			{"xfce4-terminal", []string{"--working-directory=" + dir}},
			{"xterm", []string{"-e", "cd " + dir + " && exec ${SHELL:-/bin/sh}"}},
		} {
			if _, err := exec.LookPath(t.bin); err == nil {
				return exec.Command(t.bin, t.args...)
			}
		}
		return nil
	}
}
