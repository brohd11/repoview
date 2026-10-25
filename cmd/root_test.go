package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestNestedDefaults(t *testing.T) {
	repos, _, err := rootCmd.Find([]string{"repos"})
	if err != nil {
		t.Fatal(err)
	}
	for _, cmd := range []*cobra.Command{rootCmd, repos} {
		flag := cmd.Flags().Lookup("nested")
		if flag == nil || flag.DefValue != "false" {
			t.Fatalf("%s must expose --nested with default false", cmd.CommandPath())
		}
	}
}
