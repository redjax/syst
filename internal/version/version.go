package version

import (
	"os"
	"path/filepath"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"

	// Change this for new packages
	RepoUrl = "https://github.com/redjax/syst"
)

type PackageInfo struct {
	PackageName        string
	RepoUrl            string
	RepoUser           string
	RepoName           string
	PackageVersion     string
	PackageCommit      string
	PackageReleaseDate string
}

// GetPackageInfo returns a struct with information about the current package
func GetPackageInfo() PackageInfo {
	exePath, err := os.Executable()
	binName := "<unknown>"

	if err == nil {
		binName = filepath.Base(exePath)
	}

	return PackageInfo{
		PackageName:        binName,
		RepoUrl:            RepoUrl,
		PackageVersion:     Version,
		PackageCommit:      Commit,
		PackageReleaseDate: Date,
	}
}
