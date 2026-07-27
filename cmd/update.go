package cmd

import (
	"context"
	"fmt"

	"github.com/brohd11/repoview/internal/selfupdate"

	"github.com/spf13/cobra"
)

var updateCheck bool

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for a newer repoview release and install it over this binary",
	Long: `Update compares this binary's version against the latest repoview release.
With no flags it downloads and installs the update when one is available, in place
(over wherever this binary lives).

  --check   only report whether an update is available; don't install`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: false,
	RunE:          runUpdate,
}

func init() {
	updateCmd.Flags().BoolVar(&updateCheck, "check", false, "only check for an update; don't download or install")
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	info, err := selfupdate.Check(ctx, version)
	if err != nil {
		return err
	}

	if updateCheck {
		fmt.Printf("current:  %s\nlatest:   %s\n", info.Current, info.LatestTag)
		if info.Available {
			fmt.Println("update available")
		} else {
			fmt.Println("up to date")
		}
		return nil
	}

	if !info.Available {
		if version == "dev" {
			fmt.Println("dev build, skipping update")
		} else {
			fmt.Printf("repoview is up to date (%s)\n", info.Current)
		}
		return nil
	}

	fmt.Printf("updating repoview %s → %s\n", info.Current, info.LatestTag)
	report := func(format string, a ...any) { fmt.Printf(format+"\n", a...) }
	if err := selfupdate.Apply(ctx, info, report); err != nil {
		return err
	}
	fmt.Printf("updated repoview to %s\n", info.LatestTag)
	return nil
}
