package zipBak

import (
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"

	"github.com/redjax/syst/internal/commands/zipBakCommand/commands"
	"github.com/redjax/syst/internal/commands/zipBakCommand/config"

	"github.com/spf13/cobra"
)

// NewZipbakCommand returns the zipbak parent command, with subcommands attached.
func NewZipbakCommand() *cobra.Command {
	var (
		cfgFile      string
		k            = koanf.New(".")
		backupConfig config.BackupConfig
	)

	zipbakCmd := &cobra.Command{
		Use:   "zipbak",
		Short: "Backup directories to zip archives",
		Long:  "Backup directories to zip archives using various subcommands. Filenames will automatically have a timestamp prepended.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Load config if specified
			if cfgFile != "" {
				if err := k.Load(file.Provider(cfgFile), json.Parser()); err != nil {
					return err
				}
			}
			// Load from env as fallback
			k.Load(env.Provider("ZIPBAK_", ".", func(s string) string { return s }), nil)
			// Unmarshal into struct
			return k.Unmarshal("", &backupConfig)
		},
	}

	// Add persistent flags for the zipbak command
	zipbakCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (JSON)")
	zipbakCmd.PersistentFlags().BoolP("debug", "D", false, "Enable debug logging")

	// Add subcommands, passing config
	zipbakCmd.AddCommand(commands.BackupCmd(&backupConfig, k))
	// If you have more subcommands, add them here

	return zipbakCmd
}
