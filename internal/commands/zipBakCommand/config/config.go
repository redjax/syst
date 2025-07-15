package config

type BackupConfig struct {
	BackupSrc   string   `koanf:"backup_src"`
	OutputDir   string   `koanf:"output_dir"`
	BackupName  string   `koanf:"backup_name"`
	DryRun      bool     `koanf:"dry_run"`
	DoCleanup   bool     `koanf:"do_cleanup"`
	KeepBackups int      `koanf:"keep_backups"`
	IgnorePaths []string `koanf:"ignore_paths"`
}
