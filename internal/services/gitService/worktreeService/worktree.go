package worktreeservice

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-git/go-git/v5"
	pathutil "github.com/redjax/syst/internal/utils/path"
)

// Worktree represents a Git worktree
type Worktree struct {
	Path   string
	Branch string
	Commit string
	IsBare bool
}

// WorktreeManager handles Git worktree operations
type WorktreeManager struct {
	repo     *git.Repository
	repoPath string
}

// NewWorktreeManager creates a new worktree manager
func NewWorktreeManager(repoPath string) (*WorktreeManager, error) {
	if repoPath == "" {
		var err error
		repoPath, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	return &WorktreeManager{
		repo:     repo,
		repoPath: repoPath,
	}, nil
}

// ListWorktrees returns a list of all worktrees
func (wm *WorktreeManager) ListWorktrees() ([]Worktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = wm.repoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	return parseWorktreeList(string(output)), nil
}

// parseWorktreeList parses the output of git worktree list --porcelain
func parseWorktreeList(output string) []Worktree {
	var worktrees []Worktree
	var current *Worktree

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			if current != nil {
				worktrees = append(worktrees, *current)
				current = nil
			}
			continue
		}

		if strings.HasPrefix(line, "worktree ") {
			current = &Worktree{
				Path: strings.TrimPrefix(line, "worktree "),
			}
		} else if strings.HasPrefix(line, "HEAD ") && current != nil {
			current.Commit = strings.TrimPrefix(line, "HEAD ")
		} else if strings.HasPrefix(line, "branch ") && current != nil {
			current.Branch = strings.TrimPrefix(line, "branch ")
		} else if line == "bare" && current != nil {
			current.IsBare = true
		}
	}

	// Add the last worktree if exists
	if current != nil {
		worktrees = append(worktrees, *current)
	}

	return worktrees
}

// AddWorktreeOptions contains options for adding a worktree
type AddWorktreeOptions struct {
	OutputDir  string // Parent directory for the worktree (default: ../)
	Name       string // Name of the worktree directory
	Branch     string
	NewBranch  bool
	Force      bool
	Detach     bool
	Checkout   bool
	LockReason string
}

// AddWorktree creates a new worktree
func (wm *WorktreeManager) AddWorktree(opts AddWorktreeOptions) error {
	// Build the full path from output dir and name
	var fullPath string
	if opts.OutputDir == "" {
		opts.OutputDir = "../"
	}

	// Expand the output directory
	expandedOutputDir, err := pathutil.ExpandPath(opts.OutputDir)
	if err != nil {
		return fmt.Errorf("failed to expand output directory: %w", err)
	}

	// Build full path
	fullPath = filepath.Join(expandedOutputDir, opts.Name)

	// Expand the full path in case there are any remaining ~ or relative paths
	expandedPath, err := pathutil.ExpandPath(fullPath)
	if err != nil {
		return fmt.Errorf("failed to expand path: %w", err)
	}

	args := []string{"worktree", "add"}

	if opts.Force {
		args = append(args, "--force")
	}

	if opts.Detach {
		args = append(args, "--detach")
	}

	if !opts.Checkout {
		args = append(args, "--no-checkout")
	}

	if opts.LockReason != "" {
		args = append(args, "--lock", "--reason", opts.LockReason)
	}

	if opts.NewBranch && opts.Branch != "" {
		args = append(args, "-b", opts.Branch)
	}

	args = append(args, expandedPath)

	if !opts.NewBranch && opts.Branch != "" {
		args = append(args, opts.Branch)
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = wm.repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add worktree: %w", err)
	}

	return nil
}

// RemoveWorktree removes a worktree
func (wm *WorktreeManager) RemoveWorktree(path string, force bool) error {
	// Expand the path
	expandedPath, err := pathutil.ExpandPath(path)
	if err != nil {
		return fmt.Errorf("failed to expand path: %w", err)
	}

	args := []string{"worktree", "remove"}

	if force {
		args = append(args, "--force")
	}

	args = append(args, expandedPath)

	cmd := exec.Command("git", args...)
	cmd.Dir = wm.repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
	}

	return nil
}

// PruneWorktrees removes worktree information for deleted working trees
func (wm *WorktreeManager) PruneWorktrees(dryRun bool) error {
	args := []string{"worktree", "prune"}

	if dryRun {
		args = append(args, "--dry-run")
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = wm.repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to prune worktrees: %w", err)
	}

	return nil
}

// BranchExists checks if a branch exists in the repository
func (wm *WorktreeManager) BranchExists(branchName string) (bool, error) {
	// Use git command to check if branch exists
	cmd := exec.Command("git", "rev-parse", "--verify", "refs/heads/"+branchName)
	cmd.Dir = wm.repoPath

	if err := cmd.Run(); err != nil {
		// Branch doesn't exist
		return false, nil
	}

	return true, nil
}

// GetRepoName returns the repository name from the path
func (wm *WorktreeManager) GetRepoName() string {
	return filepath.Base(wm.repoPath)
}

// GenerateWorktreePath generates a default worktree path
func (wm *WorktreeManager) GenerateWorktreePath(branchName string) string {
	repoName := wm.GetRepoName()
	parentDir := filepath.Dir(wm.repoPath)
	return filepath.Join(parentDir, fmt.Sprintf("%s-%s", repoName, branchName))
}

// OpenInEditor opens a worktree in the system's default editor
func OpenInEditor(path string) error {
	var cmd *exec.Cmd
	var cmdName string

	switch runtime.GOOS {
	case "darwin":
		// macOS - try common editors
		if _, err := exec.LookPath("code"); err == nil {
			cmdName = "code"
			cmd = exec.Command("code", path)
		} else if _, err := exec.LookPath("subl"); err == nil {
			cmdName = "subl"
			cmd = exec.Command("subl", path)
		} else {
			cmdName = "open"
			cmd = exec.Command("open", path)
		}
	case "windows":
		// Windows - try VSCode or default
		if _, err := exec.LookPath("code"); err == nil {
			cmdName = "code"
			cmd = exec.Command("code", path)
		} else {
			cmdName = "cmd /c start"
			cmd = exec.Command("cmd", "/c", "start", path)
		}
	default:
		// Linux/Unix - try common editors
		if _, err := exec.LookPath("code"); err == nil {
			cmdName = "code"
			cmd = exec.Command("code", path)
		} else if editor := os.Getenv("EDITOR"); editor != "" {
			cmdName = editor
			cmd = exec.Command(editor, path)
		} else if _, err := exec.LookPath("xdg-open"); err == nil {
			cmdName = "xdg-open"
			cmd = exec.Command("xdg-open", path)
		} else {
			return fmt.Errorf("no suitable editor found (tried: code, $EDITOR, xdg-open)")
		}
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open with %s: %w", cmdName, err)
	}

	// Don't wait for the command to finish since editors run in background
	return nil
}
