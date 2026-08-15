package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/brohd11/gitstack/repo"

	"github.com/spf13/cobra"
)

var (
	reposDir         string
	reposRaw         bool
	reposDirty       bool
	reposDepth       int
	reposIncludeRoot bool
)

var reposCmd = &cobra.Command{
	Use:   "repos [flags] -- <command...>",
	Short: "Run a shell command in every git repo nested under a directory",
	Long: `Walk a directory tree, find every nested git repo (the top-level repo is
excluded unless --include-root), and run a shell command inside each one.

The command is joined and run via "sh -c", so pipes, &&, and redirects work — quote
them as a single argument so your own shell doesn't consume them first:

  repoview repos -- git status -s
  repoview repos --dirty -- git fetch
  repoview repos -- "git log --oneline | head -1"

By default output is captured and a header is printed only for repos that produced
output; use --raw to live-stream output instead.`,
	Args:          cobra.ArbitraryArgs,
	SilenceUsage:  true,
	SilenceErrors: false,
	RunE:          runRepos,
}

func init() {
	// Stop flag parsing at the first non-flag token so the command's own -flags are
	// collected as args; "--" remains supported but optional.
	reposCmd.Flags().SetInterspersed(false)
	reposCmd.Flags().StringVarP(&reposDir, "dir", "C", "", "directory to scan (default: current directory)")
	reposCmd.Flags().BoolVar(&reposRaw, "raw", false, "live-stream each repo's output instead of capturing it")
	reposCmd.Flags().BoolVar(&reposDirty, "dirty", false, "only repos with uncommitted changes")
	reposCmd.Flags().IntVar(&reposDepth, "depth", 1, "max directory depth to search")
	reposCmd.Flags().BoolVar(&reposIncludeRoot, "include-root", false, "also run in the top-level repo (the scanned dir itself)")
	rootCmd.AddCommand(reposCmd)
}

func runRepos(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	errOutW := cmd.ErrOrStderr()

	base := reposDir
	if base == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("could not get current working directory: %w", err)
		}
		base = cwd
	}
	base, err := filepath.Abs(base)
	if err != nil {
		return fmt.Errorf("could not resolve absolute path for %s: %w", base, err)
	}

	repos, err := repo.FindGitRepos(base, reposDepth)
	if err != nil {
		return err
	}
	// FindGitRepos excludes the base itself; --include-root opts it back in as the "." entry, which
	// flows through the dirty filter and both output modes unchanged (base-relative "." resolves to
	// base). Only when base is actually a checkout.
	if reposIncludeRoot {
		if _, ok := repo.DescribeRoot(base); ok {
			repos = append([]string{"."}, repos...)
		}
	}
	if reposDirty {
		dirty := repos[:0:0]
		for _, rel := range repos {
			if repo.HasUncommittedChanges(filepath.Join(base, rel)) {
				dirty = append(dirty, rel)
			}
		}
		repos = dirty
	}

	// No command: list mode — print the matching repo paths, one per line.
	if len(args) == 0 {
		for _, rel := range repos {
			fmt.Fprintln(out, rel)
		}
		return nil
	}

	cmdStr := strings.Join(args, " ")
	prefix := filepath.Base(base)

	for _, rel := range repos {
		full := filepath.Join(base, rel)
		display := filepath.Join(prefix, rel)
		c := exec.CommandContext(cmd.Context(), "sh", "-c", cmdStr)
		c.Dir = full

		if reposRaw {
			reposHeader(out, display)
			c.Stdout = out
			c.Stderr = errOutW
			if err := c.Run(); err != nil {
				fmt.Fprintf(errOutW, "error in %s: %v\n", display, err)
			}
			continue
		}

		var stdout, stderr bytes.Buffer
		c.Stdout = &stdout
		c.Stderr = &stderr
		runErr := c.Run()

		outStr := strings.TrimSpace(stdout.String())
		errStr := strings.TrimSpace(stderr.String())
		if outStr != "" || errStr != "" {
			reposHeader(out, display)
			if outStr != "" {
				fmt.Fprintln(out, outStr)
			}
			if errStr != "" {
				fmt.Fprintln(errOutW, errStr)
			}
		}
		if runErr != nil {
			fmt.Fprintf(errOutW, "error in %s: %v\n", display, runErr)
		}
	}
	return nil
}

func reposHeader(w io.Writer, text string) {
	line := strings.Repeat("-", 50)
	fmt.Fprintf(w, "\n%s\n📁 %s\n%s\n", line, text, line)
}
