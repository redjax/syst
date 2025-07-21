package version

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
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

// Returns the user/repo portion of a repository URL
func getRepoUrlPath() (string, error) {
	u, err := url.Parse(RepoUrl)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Trim leading and trailing slashes for safe splitting
	trimmedPath := strings.Trim(u.Path, "/")

	// Split the path
	segments := strings.Split(trimmedPath, "/")
	if len(segments) < 2 {
		return "", fmt.Errorf("URL doesn't contain enough path segments for user/repo")
	}

	return fmt.Sprintf("%s/%s", segments[0], segments[1]), nil
}
