package cmd

import (
	"github.com/brohd11/gitstack/repocmd"
)

// The command lives in gitstack/repocmd; repoview supplies the shallow default depth.
func init() {
	rootCmd.AddCommand(repocmd.New(repocmd.Options{
		AppName:      "repoview",
		DefaultDepth: 1,
		// The real default is the ladder resolveDepth walks, not the 1 pflag would print.
		DepthDefText: "$" + depthEnv + ", else 1",
		ResolveDepth: resolveDepth,
	}))
}
