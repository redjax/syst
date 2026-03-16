package cmd

import (
	"testing"
)

func TestRootCommandExists(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("rootCmd is nil")
	}
	if rootCmd.Use != "syst" {
		t.Errorf("rootCmd.Use = %q, want %q", rootCmd.Use, "syst")
	}
}

func TestSubcommandsRegistered(t *testing.T) {
	expectedSubcommands := []string{
		"show",
		"ping",
		"strutil",
		"which",
		"git",
		"weather",
		"generate",
		"encode",
		"sqlite",
		"ssh",
		"scanpath",
		"zipbak",
		"self",
	}

	commands := rootCmd.Commands()
	registered := make(map[string]bool)
	for _, cmd := range commands {
		registered[cmd.Name()] = true
	}

	for _, name := range expectedSubcommands {
		if !registered[name] {
			t.Errorf("expected subcommand %q not registered", name)
		}
	}
}

func TestRootCommandHasGlobalFlags(t *testing.T) {
	flags := rootCmd.PersistentFlags()

	if flags.Lookup("config") == nil {
		t.Error("missing --config flag")
	}
	if flags.Lookup("debug") == nil {
		t.Error("missing --debug flag")
	}
	if flags.Lookup("version") == nil {
		t.Error("missing --version flag")
	}
}

func TestRootCommandHelp(t *testing.T) {
	// Ensure the root command doesn't error when showing help
	rootCmd.SetArgs([]string{"--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("root --help returned error: %v", err)
	}
}
