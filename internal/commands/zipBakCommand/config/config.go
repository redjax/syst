package config

type BackupConfig struct {
	// The src directory to backup
	BackupSrc string `koanf:"backup_src"`
	// The target directory where backup will be created
	OutputDir string `koanf:"output_dir"`
	// Name of the backup file
	BackupName string `koanf:"backup_name"`
	// Enable dry run, don't take any actions, just print what would happen
	DryRun bool `koanf:"dry_run"`
	// Cleanup backup dest at end
	DoCleanup bool `koanf:"do_cleanup"`
	// Number of backups to keep (keeps n newest backups)
	KeepBackups int `koanf:"keep_backups"`
	// Ignore paths during backup
	IgnorePaths []string `koanf:"ignore_paths"`
}
