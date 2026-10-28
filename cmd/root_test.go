package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestBootstrapHooksForSubcommands(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	if rootCmd.PersistentPreRunE == nil {
		t.Fatal("root command needs hook config bootstrap")
	}
	for _, cmd := range rootCmd.Commands() {
		if cmd.PersistentPreRunE != nil || cmd.PersistentPreRun != nil {
			t.Fatalf("%s shadows startup bootstrap", cmd.Name())
		}
	}
	var stdout, stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	if err := rootCmd.PersistentPreRunE(cmd, nil); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, ".repoview", "hooks", "config.yml")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("startup did not create hooks: %v", err)
	}
	if stdout.Len() != 0 || !strings.Contains(stderr.String(), path) {
		t.Fatalf("creation notice must go to stderr: stdout=%q, stderr=%q", stdout.String(), stderr.String())
	}
	stderr.Reset()
	if err := rootCmd.PersistentPreRunE(cmd, nil); err != nil || stderr.Len() != 0 {
		t.Fatalf("existing config should be quiet: err=%v, stderr=%q", err, stderr.String())
	}
	// A blocked config directory reports the problem without stopping startup.
	if err := os.RemoveAll(filepath.Join(home, ".repoview")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".repoview"), []byte("blocked"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := rootCmd.PersistentPreRunE(cmd, nil); err != nil || !strings.Contains(stderr.String(), "could not prepare hook config") {
		t.Fatalf("creation failure should be nonfatal and reported: err=%v, stderr=%q", err, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("bootstrap polluted command output: %q", stdout.String())
	}
}

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
