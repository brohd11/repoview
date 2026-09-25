// Command repoview shows git status for every repository under a directory and drives
// fetch/pull/push/commit from a TUI. `repoview repos` runs a command in each repo.
package main

import "github.com/brohd11/repoview/cmd"

func main() {
	cmd.Execute()
}
