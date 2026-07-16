package cmd

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/brohd11/repoview/internal/app"

	"github.com/spf13/cobra"
)

// version is the binary version; defaults to "dev" for a plain `go build`. (No ldflags injection
// yet — the makefile doesn't set it — but Version is wired so `repoview --version` works and a
// future -X can fill it in, mirroring gdaddon.)
var version = "dev"

var rootDepth int

var rootCmd = &cobra.Command{
	Use:   "repoview [dir] [depth]",
	Short: "Show git status across every repo nested under a directory (TUI)",
	Long: `repoview scans a directory for nested git checkouts and shows each one's status —
branch, uncommitted changes, ahead/behind — in a single TUI list, driving fetch/pull/
push/commit through a shared git menu.

Positional args are order-free: an all-digits argument is the scan depth, anything else
is the directory. dir defaults to the current directory; depth to 1 (that dir only).

  repoview            # current dir, depth 1
  repoview 4          # current dir, depth 4
  repoview /path 3    # /path, depth 3`,
	Version:       version,
	Args:          cobra.ArbitraryArgs,
	SilenceUsage:  true,
	SilenceErrors: false,
	RunE:          runRoot,
}

func init() {
	rootCmd.SetVersionTemplate("repoview {{.Version}}\n")
	rootCmd.Flags().IntVarP(&rootDepth, "depth", "d", 1, "maximum directory depth to scan for git repos")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// runRoot parses the flexible positionals (int → depth, else → root dir) and launches the TUI.
// A positional integer overrides the --depth flag; the last non-integer arg wins as the root.
func runRoot(cmd *cobra.Command, args []string) error {
	depth := rootDepth
	root := "."
	for _, arg := range args {
		if n, err := strconv.Atoi(arg); err == nil {
			depth = n
		} else {
			root = arg
		}
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	return app.Run(abs, depth)
}
