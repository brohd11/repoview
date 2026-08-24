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

// version is the binary version; defaults to "dev" for a plain `go build`. The makefile stamps
// it via -X ldflags (git describe --tags --always --dirty), so release and `make` binaries report
// their real version and the self-update check can compare it against the latest tag.
var version = "dev"

var rootDepth int

// depthEnv names the environment variable that supplies a scan depth when the command
// line gives none, so a depth you always want need not be typed every run. It backs both
// the TUI and `repoview repos`, whose --depth means the same thing.
const depthEnv = "REPOVIEW_DEPTH"

// resolveDepth picks the depth to scan with: the flag when it was actually typed,
// otherwise REPOVIEW_DEPTH, otherwise the flag's own default. The ladder itself is
// goutil/envopt.Int -- gote had written the identical function for GOTE_DEPTH, down to
// the doc comment and the test table. repoview does not need envopt's `set` return.
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
	// The real default is the ladder resolveDepth walks, not the 1 pflag would print on
	// its own. DefValue is only ever the string cobra renders in "(default %s)", so
	// rewriting it states that ladder in the one place a reader looks for it.
	rootCmd.Flags().Lookup("depth").DefValue = "$REPOVIEW_DEPTH, else 1"
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// runRoot parses the flexible positionals (int → depth, else → root dir) and launches the TUI.
// A positional integer overrides the --depth flag, which in turn overrides $REPOVIEW_DEPTH
// (see resolveDepth); the last non-integer arg wins as the root.
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
