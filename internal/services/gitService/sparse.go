package gitservice

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type SparseCloneOptions struct {
	Provider   string
	User       string
	Repository string
	Output     string
	Branch     string
	Paths      []string
	// ssh or https
	Protocol string
}

func SparseClone(opts SparseCloneOptions) error {
	if !CheckGitInstalled() {
		fmt.Printf("Error: git is not installed")
		return ErrGitNotInstalled
	}

	if !ValidateGitProvider(opts.Provider) {
		return fmt.Errorf("unknown git provider: %s", opts.Provider)
	}

	// Determine output directory
	outputDir := opts.Output
	if outputDir == "" || outputDir == "." {
		outputDir = strings.TrimSuffix(opts.Repository, ".git")
	}

	host := GetHostByProvider(opts.Provider)
	repoURL := BuildRepoURL(opts.Protocol, host, opts.User, opts.Repository)

	// Clone no-checkout
	if err := CloneNoCheckout(repoURL, outputDir); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return fmt.Errorf("could not resolve output path: %w", err)
	}

	if _, err := os.Stat(absOutputDir); os.IsNotExist(err) {
		return fmt.Errorf("output directory does not exist after clone")
	}

	if err := os.Chdir(absOutputDir); err != nil {
		return fmt.Errorf("failed to enter output directory: %w", err)
	}

	if err := SparseCheckoutInit(); err != nil {
		return fmt.Errorf("git sparse-checkout init failed: %w", err)
	}

	if err := SparseCheckoutPaths(opts.Paths); err != nil {
		return fmt.Errorf("git sparse-checkout set failed: %w", err)
	}

	if err := CheckoutBranch(opts.Branch); err != nil {
		return fmt.Errorf("git checkout failed: %w", err)
	}

	fmt.Println("Sparse clone complete!")
	return nil
}

func SparseCheckoutInit() error {
	cmd := execCommand("git", "sparse-checkout", "init", "--cone")
	return cmd.Run()
}

func SparseCheckoutPaths(paths []string) error {
	args := append([]string{"sparse-checkout", "set"}, paths...)
	cmd := execCommand("git", args...)
	return cmd.Run()
}
