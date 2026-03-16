package gitservice

import "testing"

func TestGetHostByProvider(t *testing.T) {
	tests := []struct {
		provider string
		want     string
	}{
		{"github", "github.com"},
		{"gitlab", "gitlab.com"},
		{"codeberg", "codeberg.org"},
		{"unknown", ""},
		{"", ""},
		{"GitHub", ""}, // case sensitive
		{"GITHUB", ""}, // case sensitive
	}
	for _, tt := range tests {
		got := GetHostByProvider(tt.provider)
		if got != tt.want {
			t.Errorf("GetHostByProvider(%q) = %q, want %q", tt.provider, got, tt.want)
		}
	}
}

func TestValidateGitProvider(t *testing.T) {
	tests := []struct {
		provider string
		want     bool
	}{
		{"github", true},
		{"gitlab", true},
		{"codeberg", true},
		{"unknown", false},
		{"", false},
		{"bitbucket", false},
	}
	for _, tt := range tests {
		got := ValidateGitProvider(tt.provider)
		if got != tt.want {
			t.Errorf("ValidateGitProvider(%q) = %v, want %v", tt.provider, got, tt.want)
		}
	}
}

func TestBuildRepoURL(t *testing.T) {
	tests := []struct {
		name     string
		protocol string
		host     string
		user     string
		repo     string
		want     string
	}{
		{"https github", "https", "github.com", "redjax", "syst", "https://github.com/redjax/syst.git"},
		{"ssh github", "ssh", "github.com", "redjax", "syst", "git@github.com:redjax/syst.git"},
		{"https gitlab", "https", "gitlab.com", "user", "project", "https://gitlab.com/user/project.git"},
		{"ssh codeberg", "ssh", "codeberg.org", "user", "repo", "git@codeberg.org:user/repo.git"},
		{"http fallback to https format", "http", "github.com", "user", "repo", "https://github.com/user/repo.git"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildRepoURL(tt.protocol, tt.host, tt.user, tt.repo)
			if got != tt.want {
				t.Errorf("BuildRepoURL(%q, %q, %q, %q) = %q, want %q",
					tt.protocol, tt.host, tt.user, tt.repo, got, tt.want)
			}
		})
	}
}
