package commands

import (
	"fmt"
	"os"

	"github.com/redjax/syst/internal/commands/zipBakCommand/archive"
	"github.com/redjax/syst/internal/commands/zipBakCommand/config"
	"github.com/redjax/syst/internal/utils/spinner"

	"github.com/knadh/koanf/v2"
	"github.com/spf13/cobra"
)

func BackupCmd(cfg *config.BackupConfig, k *koanf.Koanf) *cobra.Command {
	var (
		outputDir   string
		backupName  string
		dryRun      bool
		doCleanup   bool
		keepBackups int
		ignorePaths []string
		backupSrc   string
	)

	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup a directory to a zip archive",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Override config with CLI flags if set
			if backupSrc != "" {
				cfg.BackupSrc = backupSrc
			}
			if outputDir != "" {
				cfg.OutputDir = outputDir
			}
			if backupName != "" {
				cfg.BackupName = backupName
			}
			if cmd.Flags().Changed("dry-run") {
				cfg.DryRun = dryRun
			}
			if cmd.Flags().Changed("cleanup") {
				cfg.DoCleanup = doCleanup
			}
			if len(ignorePaths) > 0 {
				cfg.IgnorePaths = ignorePaths
			}
			if cfg.BackupSrc == "" {
				return fmt.Errorf("backup source directory is required")
			}

			cfg.KeepBackups = keepBackups

			if cmd.Flags().Changed("keep-backups") && !cfg.DoCleanup {
				fmt.Fprintln(os.Stderr, "Warning: --keep-backups was specified, but --cleanup was not set. No cleanup will be performed.")
			}

			stop := spinner.StartSpinner("Backing up...")
			err := archive.StartBackup(cfg)
			stop()
			if err != nil {
				return err
			}
			fmt.Println("Backup completed successfully.")
			return nil
		},
	}

	cmd.Flags().StringVarP(&backupSrc, "src", "s", "", "Source directory to backup")
	cmd.Flags().StringVarP(&outputDir, "output-dir", "o", "", "Output directory for the backup")
	cmd.Flags().StringVarP(&backupName, "filename", "f", "", "Base filename for the backup")
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Perform a dry run")
	cmd.Flags().BoolVarP(&doCleanup, "cleanup", "c", false, "Clean up old backups after backup")
	cmd.Flags().IntVarP(&keepBackups, "keep-backups", "k", 3, "Number of backups to keep")
	cmd.Flags().StringSliceVarP(&ignorePaths, "ignore-paths", "i", nil, "Paths to ignore during backup")

	return cmd
}
