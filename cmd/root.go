// The root command for the CLI.
// This root 'composes' your subcommands and provides global config flags like --debug.
package cmd

import (
	"fmt"
	"os"
	"strings"

	// Import your CLI subcommands
	// "github.com/redjax/syst/internal/commands"
	scanPath "github.com/redjax/syst/internal/commands/scanPathCommand"
	"github.com/redjax/syst/internal/commands/showCommand"
	zipBak "github.com/redjax/syst/internal/commands/zipBakCommand"

	// Import your CLI config
	// "github.com/redjax/syst/internal/config"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/cobra"
)

var (
	// A path to a file to load configuration from
	cfgFile string
	// For enabling debug logging with --debug/-D
	debug bool
	// Initialize Koanf config instance
	k = koanf.New(".")
)

// Cobra root command
var rootCmd = &cobra.Command{
	// The command you run to call the compiled binary
	Use: "syst",
	// A short description of what the command does
	Short: "System controls, script launcher, etc.",
	// A longer description for the command
	Long: `My personal system swiss utility knife. Cross platform utility functions & scripts.`,
	// Adds a help menu you can display with --help/-h
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Execute the root Cobra command
func Execute() {
	// Import this into a main.go and call with cmd.Execute()
	cobra.CheckErr(rootCmd.Execute())
}

// Initialize the root command
func init() {
	// Add flags to the CLI's root command, making them 'global'
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (JSON)")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "D", false, "Enable debug logging")

	// Add other CLI subcommands
	// rootCmd.AddCommand(commands.HelloCmd())
	rootCmd.AddCommand(showCommand.NewShowCmd())
	rootCmd.AddCommand(zipBak.NewZipbakCommand())
	rootCmd.AddCommand(scanPath.NewScanPathCommand())

	// Call the initConfig function when the root command is initialized
	cobra.OnInitialize(initConfig)
}

// Load configuration for CLI app
func initConfig() {
	// Load from config file passed as arg
	if cfgFile != "" {
		if err := k.Load(file.Provider(cfgFile), json.Parser()); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
	}

	// Load from environment variables (prefix "SYST_")
	k.Load(env.Provider("SYST_", ".", func(s string) string {
		return strings.Replace(strings.ToLower(
			strings.TrimPrefix(s, "SYST_")), "_", ".", -1)
	}), nil)

	// Optionally: Unmarshal into a struct here if you have one
	// k.Unmarshal("", &yourConfigStruct)
}
