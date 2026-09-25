package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/brohd11/goutil/envopt"
	"github.com/brohd11/repoview/internal/app"

	"github.com/spf13/cobra"
)

// version is stamped by the makefile via -X ldflags; "dev" for a plain go build.
var version = "dev"

var rootDepth int

// depthEnv supplies a default scan depth for both the TUI and `repoview repos`.
const depthEnv = "REPOVIEW_DEPTH"

// resolveDepth picks the scan depth: a typed flag, else $REPOVIEW_DEPTH, else the default.
func resolveDepth(flagDepth int, flagChanged bool) (int, error) {
	depth, _, err := envopt.Int(depthEnv, flagDepth, flagChanged)
	return depth, err
}

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
  repoview /path 3    # /path, depth 3

Set REPOVIEW_DEPTH to the depth you always want and both this and "repoview repos" start
there instead of 1. A depth given as an argument or with --depth still wins;
REPOVIEW_DEPTH= (blank) drops it for one run.`,
	Version:       version,
	Args:          cobra.ArbitraryArgs,
	SilenceUsage:  true,
	SilenceErrors: false,
	RunE:          runRoot,
}

func init() {
	rootCmd.SetVersionTemplate("repoview {{.Version}}\n")
	rootCmd.Flags().IntVarP(&rootDepth, "depth", "d", 1, "maximum directory depth to scan for git repos")
	// Show the fallback ladder resolveDepth walks.
	rootCmd.Flags().Lookup("depth").DefValue = "$REPOVIEW_DEPTH, else 1"
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// runRoot parses the positionals (int → depth, else root dir) and launches the TUI. A
// positional depth beats --depth, which beats $REPOVIEW_DEPTH.
func runRoot(cmd *cobra.Command, args []string) error {
	depth, err := resolveDepth(rootDepth, cmd.Flags().Changed("depth"))
	if err != nil {
		return err
	}
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
		return fmt.Errorf("could not resolve absolute path for %s: %w", root, err)
	}
	return app.Run(abs, depth, version)
}
