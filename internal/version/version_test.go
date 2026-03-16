package version

import "testing"

func TestCompareVersion(t *testing.T) {
	tests := []struct {
		name   string
		v1, v2 string
		want   int
	}{
		{"equal", "1.0.0", "1.0.0", 0},
		{"v1 greater major", "2.0.0", "1.0.0", 1},
		{"v1 less major", "1.0.0", "2.0.0", -1},
		{"v1 greater minor", "1.2.0", "1.1.0", 1},
		{"v1 less minor", "1.1.0", "1.2.0", -1},
		{"v1 greater patch", "1.0.2", "1.0.1", 1},
		{"v1 less patch", "1.0.1", "1.0.2", -1},
		{"different lengths", "1.0", "1.0.0", 0},
		{"with suffix stripped", "1.2.3-abc123", "1.2.3", 0},
		{"suffix comparison", "1.2.3-abc", "1.2.4", -1},
		{"v prefix stripped TBD", "1.0.0", "1.0.0", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareVersion(tt.v1, tt.v2)
			if got != tt.want {
				t.Errorf("compareVersion(%q, %q) = %d, want %d", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}

func TestGetRepoUrlPath(t *testing.T) {
	// This depends on the package-level RepoUrl variable
	// Default is "https://github.com/redjax/syst"
	path, err := getRepoUrlPath()
	if err != nil {
		t.Fatalf("getRepoUrlPath() error: %v", err)
	}
	if path != "redjax/syst" {
		t.Errorf("getRepoUrlPath() = %q, want %q", path, "redjax/syst")
	}
}

func TestGetPackageInfo(t *testing.T) {
	info := GetPackageInfo()

	if info.RepoUrl != RepoUrl {
		t.Errorf("PackageInfo.RepoUrl = %q, want %q", info.RepoUrl, RepoUrl)
	}
	if info.PackageVersion != Version {
		t.Errorf("PackageInfo.PackageVersion = %q, want %q", info.PackageVersion, Version)
	}
	if info.PackageCommit != Commit {
		t.Errorf("PackageInfo.PackageCommit = %q, want %q", info.PackageCommit, Commit)
	}
	if info.PackageName == "" {
		t.Error("PackageInfo.PackageName is empty")
	}
}
