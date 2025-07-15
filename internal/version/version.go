package version

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"

	// Change this for new packages
	RepoUser = "redjax"
	RepoName = "syst"
	RepoUrl  = "https://github.com/redjax/syst"
	Package  = "syst"
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
	return PackageInfo{
		PackageName:        Package,
		RepoUrl:            RepoUrl,
		RepoUser:           RepoUser,
		RepoName:           RepoName,
		PackageVersion:     Version,
		PackageCommit:      Commit,
		PackageReleaseDate: Date,
	}
}
