// The root command for the CLI.
// This root 'composes' your subcommands and provides global config flags like --debug.
package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	// Import your CLI subcommands
	pingo "github.com/redjax/syst/internal/commands/pingCommand"
	scanPath "github.com/redjax/syst/internal/commands/scanPathCommand"
	selfcommand "github.com/redjax/syst/internal/commands/selfCommand"
	"github.com/redjax/syst/internal/commands/showCommand"
	strutilcommand "github.com/redjax/syst/internal/commands/strUtilCommand"
	zipBak "github.com/redjax/syst/internal/commands/zipBakCommand"
	"github.com/redjax/syst/internal/version"

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
	rootCmd.PersistentFlags().BoolP("version", "v", false, "Print version and exit")

	// Add other CLI subcommands
	rootCmd.AddCommand(showCommand.NewShowCmd())
	rootCmd.AddCommand(zipBak.NewZipbakCommand())
	rootCmd.AddCommand(scanPath.NewScanPathCommand())
	rootCmd.AddCommand(pingo.NewPingCommand())
	rootCmd.AddCommand(strutilcommand.NewStrUtilCommand())
	rootCmd.AddCommand(selfcommand.NewSelfCommand())

	// Handle persistent flags like -v/--version and -d/--debug
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		// Handle -v/--version
		v, _ := cmd.Flags().GetBool("version")
		if v {
			fmt.Printf("syst version:%s commit:%s date:%s\n", version.Version, version.Commit, version.Date)
			os.Exit(0)
		}

		// Handle -D/--debug
		if d, _ := cmd.Flags().GetBool("debug"); d {
			log.SetFlags(log.LstdFlags | log.Lshortfile)
			log.Println("DEBUG mode enabled")
		}
	}

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
