package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
// otherwise $REPOVIEW_DEPTH, otherwise the flag's own default. Everything typed still
// outranks the environment, and a positional integer outranks the flag in runRoot, so
// the ladder reads argument, flag, environment, default.
//
// A malformed or negative value is refused rather than ignored: the variable lives in a
// shell profile, where a silently misread depth would never be noticed. An unset or
// blank one is not malformed, which is what makes `REPOVIEW_DEPTH= repoview` the way to
// drop it for a single run.
func resolveDepth(flagDepth int, flagChanged bool) (int, error) {
	if flagChanged {
		return flagDepth, nil
	}
	raw, ok := os.LookupEnv(depthEnv)
	if !ok {
		return flagDepth, nil
	}
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return flagDepth, nil
	}
	n, err := strconv.Atoi(trimmed)
	if err != nil {
		return flagDepth, fmt.Errorf("%s %q is not a number", depthEnv, raw)
	}
	if n < 0 {
		return flagDepth, fmt.Errorf("%s %d is negative", depthEnv, n)
	}
	return n, nil
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
