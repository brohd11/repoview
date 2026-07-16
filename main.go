// Command repoview shows git status across every repository nested under a directory — branch,
// uncommitted changes, ahead/behind — on a fresh scan each run, and drives fetch/pull/push/commit
// through a shared TUI. It's the manifest-free sibling of gdaddon, built on the same bubblestack
// framework and gitstack git engine/screens. The bare invocation launches the TUI; the `repos`
// subcommand runs a shell command across every nested repo (see cmd/).
package main

import "github.com/brohd11/repoview/cmd"

func main() {
	cmd.Execute()
}
