package gitservice

import (
	"os"
	"os/exec"
	"testing"
)

func isInGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Stderr = nil
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(out) > 0
}

func TestIsGitRepo(t *testing.T) {
	isRepo, err := IsGitRepo()
	if err != nil {
		t.Fatalf("IsGitRepo() error: %v", err)
	}
	// We're running tests inside the syst repo, so this should be true
	if !isRepo {
		t.Skip("not inside a git repo, skipping")
	}
}

func TestCheckGitInstalled(t *testing.T) {
	got := CheckGitInstalled()
	// git should be available in CI and dev environments
	if !got {
		t.Skip("git not installed, skipping")
	}
}

func TestGetCurrentBranch(t *testing.T) {
	if !isInGitRepo() {
		t.Skip("not in a git repo")
	}
	branch, err := GetCurrentBranch()
	if err != nil {
		t.Fatalf("GetCurrentBranch() error: %v", err)
	}
	if branch == "" {
		t.Error("GetCurrentBranch() returned empty string")
	}
}

func TestGetBranchSyncStatus(t *testing.T) {
	if !isInGitRepo() {
		t.Skip("not in a git repo")
	}
	status := GetBranchSyncStatus()
	if status.CurrentBranch == "" {
		t.Error("BranchSyncStatus.CurrentBranch is empty")
	}
}

func TestDirSize(t *testing.T) {
	// Create a temp directory with a known file
	dir := t.TempDir()
	f, err := os.CreateTemp(dir, "testfile")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	data := make([]byte, 1024)
	f.Write(data)
	f.Close()

	size, err := dirSize(dir)
	if err != nil {
		t.Fatalf("dirSize() error: %v", err)
	}
	if size < 1024 {
		t.Errorf("dirSize() = %d, want >= 1024", size)
	}
}
