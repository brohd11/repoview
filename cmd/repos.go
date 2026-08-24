package cmd

import (
	"github.com/brohd11/gitstack/repocmd"
)

// The command itself lives in gitstack/repocmd, beside the engine it drives — gdaddon
// ships the same one. What is repoview's here is the shallow default and the
// $REPOVIEW_DEPTH ladder underneath it (see root.go).
func init() {
	rootCmd.AddCommand(repocmd.New(repocmd.Options{
		AppName:      "repoview",
		DefaultDepth: 1,
		// The real default is the ladder resolveDepth walks, not the 1 pflag would print.
		DepthDefText: "$" + depthEnv + ", else 1",
		ResolveDepth: resolveDepth,
	}))
}
