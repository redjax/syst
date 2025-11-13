package gitcommand

import (
	"fmt"
	"path/filepath"
	"strings"

	worktreeservice "github.com/redjax/syst/internal/services/gitService/worktreeService"
	"github.com/spf13/cobra"
)

// NewGitWorktreeCommand returns the git worktree command with all subcommands.
func NewGitWorktreeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "worktree",
		Aliases: []string{"wt"},
		Short:   "Manage Git worktrees",
		Long:    "Manage Git worktrees with both CLI and TUI interfaces. Run without subcommands to open the interactive TUI.",
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no args provided, run the TUI
			// (subcommands will handle themselves)
			repoPath, _ := cmd.Flags().GetString("repo")
			return worktreeservice.RunWorktreeTUI(repoPath)
		},
	}

	// Add flags
	cmd.PersistentFlags().StringP("repo", "r", "", "Path to git repository (defaults to current directory)")

	// Add subcommands
	cmd.AddCommand(NewWorktreeListCommand())
	cmd.AddCommand(NewWorktreeAddCommand())
	cmd.AddCommand(NewWorktreeRemoveCommand())
	cmd.AddCommand(NewWorktreeMoveCommand())
	cmd.AddCommand(NewWorktreePruneCommand())

	return cmd
}

// NewWorktreeListCommand returns the worktree list command.
func NewWorktreeListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all worktrees",
		Long:    "List all Git worktrees in the repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath, _ := cmd.Flags().GetString("repo")

			manager, err := worktreeservice.NewWorktreeManager(repoPath)
			if err != nil {
				return err
			}

			worktrees, err := manager.ListWorktrees()
			if err != nil {
				return err
			}

			if len(worktrees) == 0 {
				fmt.Println("No worktrees found")
				return nil
			}

			fmt.Println("Worktrees:")
			for _, wt := range worktrees {
				branch := wt.Branch
				if branch == "" {
					branch = "(detached)"
				}
				fmt.Printf("  %s [%s]\n", wt.Path, branch)
			}

			return nil
		},
	}

	return cmd
}

// NewWorktreeAddCommand returns the worktree add command.
func NewWorktreeAddCommand() *cobra.Command {
	var (
		outputDir  string
		name       string
		branch     string
		force      bool
		detach     bool
		noCheckout bool
	)

	cmd := &cobra.Command{
		Use:     "add <name> [<branch>]",
		Aliases: []string{"a", "create", "new"},
		Short:   "Add a new worktree",
		Long:    "Create a new Git worktree. Branches are automatically created if they don't exist.",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath, _ := cmd.Flags().GetString("repo")

			manager, err := worktreeservice.NewWorktreeManager(repoPath)
			if err != nil {
				return err
			}

			// First arg is the name
			name = args[0]

			// If branch not specified via flag, check args
			if branch == "" && len(args) > 1 {
				branch = args[1]
			}

			// If name looks like a branch name and no explicit branch, use it as branch
			if branch == "" && !strings.Contains(name, "/") && len(args) == 1 {
				branch = name
			}

			// Auto-detect if we need to create a new branch
			createNewBranch := false
			if branch != "" {
				// Check if branch exists
				exists, err := manager.BranchExists(branch)
				if err != nil {
					return fmt.Errorf("failed to check branch: %w", err)
				}
				// If branch doesn't exist, automatically create it
				if !exists {
					createNewBranch = true
				}
			}

			opts := worktreeservice.AddWorktreeOptions{
				OutputDir: outputDir,
				Name:      name,
				Branch:    branch,
				NewBranch: createNewBranch,
				Force:     force,
				Detach:    detach,
				Checkout:  !noCheckout,
			}

			if err := manager.AddWorktree(opts); err != nil {
				return err
			}

			// Construct display path
			displayOutputDir := outputDir
			if displayOutputDir == "" {
				displayOutputDir = "../"
			}
			fullPath := displayOutputDir + name

			fmt.Printf("Created worktree at %s", fullPath)
			if branch != "" {
				fmt.Printf(" on branch %s", branch)
			}
			fmt.Println()

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output-dir", "o", "", "Parent directory for the worktree (default: ../)")
	cmd.Flags().StringVarP(&branch, "branch", "b", "", "Branch to checkout or create (auto-created if doesn't exist)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force creation even if worktree path exists")
	cmd.Flags().BoolVar(&detach, "detach", false, "Detach HEAD in the new worktree")
	cmd.Flags().BoolVar(&noCheckout, "no-checkout", false, "Don't checkout files in the new worktree")

	return cmd
}

// NewWorktreeRemoveCommand returns the worktree remove command.
func NewWorktreeRemoveCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:     "remove <worktree>",
		Aliases: []string{"rm", "delete", "del"},
		Short:   "Remove a worktree",
		Long:    "Remove a Git worktree",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath, _ := cmd.Flags().GetString("repo")

			manager, err := worktreeservice.NewWorktreeManager(repoPath)
			if err != nil {
				return err
			}

			worktreePath := args[0]

			// Confirm before deletion unless force is set
			if !force {
				fmt.Printf("Remove worktree %s? [y/N]: ", worktreePath)
				var response string
				fmt.Scanln(&response)
				response = strings.ToLower(strings.TrimSpace(response))
				if response != "y" && response != "yes" {
					fmt.Println("Cancelled")
					return nil
				}
			}

			if err := manager.RemoveWorktree(worktreePath, force); err != nil {
				return err
			}

			fmt.Printf("Removed worktree %s\n", worktreePath)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force removal even if worktree is dirty")

	return cmd
}

// NewWorktreeMoveCommand returns the worktree move command.
func NewWorktreeMoveCommand() *cobra.Command {
	var (
		destDir string
		newName string
	)

	cmd := &cobra.Command{
		Use:     "move <worktree-path>",
		Aliases: []string{"mv"},
		Short:   "Move a worktree to a new location",
		Long: `Move an existing worktree to a new directory, optionally renaming it.

The worktree will be moved to the destination directory with the same name
unless --name is specified to rename it during the move.

Examples:
  # Move worktree to new location (keeps same name)
  syst git worktree move ~/git/worktrees/syst-feat-x --dest ~/git/.worktrees

  # Move and rename worktree
  syst git worktree move ~/git/worktrees/old-name --dest ~/git/.worktrees --name new-name`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath, _ := cmd.Flags().GetString("repo")
			worktreePath := args[0]

			// Validate required flags
			if destDir == "" {
				return fmt.Errorf("destination directory is required (--dest or -d)")
			}

			manager, err := worktreeservice.NewWorktreeManager(repoPath)
			if err != nil {
				return err
			}

			// Execute move
			if err := manager.MoveWorktree(worktreePath, destDir, newName); err != nil {
				return fmt.Errorf("failed to move worktree: %w", err)
			}

			// Display success message
			targetName := newName
			if targetName == "" {
				targetName = filepath.Base(worktreePath)
			}
			finalPath := filepath.Join(destDir, targetName)

			fmt.Printf("✅ Worktree moved successfully:\n")
			fmt.Printf("   From: %s\n", worktreePath)
			fmt.Printf("   To:   %s\n", finalPath)

			return nil
		},
	}

	cmd.Flags().StringVarP(&destDir, "dest", "d", "", "Destination directory (required)")
	cmd.Flags().StringVarP(&newName, "name", "n", "", "New name for worktree (optional, keeps original name if not specified)")
	// #nosec G104 - MarkFlagRequired error ignored, flag existence guaranteed by Cobra
	_ = cmd.MarkFlagRequired("dest")

	return cmd
}

// NewWorktreePruneCommand returns the worktree prune command.
func NewWorktreePruneCommand() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Prune worktree information",
		Long:  "Remove worktree information for deleted working trees",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath, _ := cmd.Flags().GetString("repo")

			manager, err := worktreeservice.NewWorktreeManager(repoPath)
			if err != nil {
				return err
			}

			if err := manager.PruneWorktrees(dryRun); err != nil {
				return err
			}

			if dryRun {
				fmt.Println("Dry run completed")
			} else {
				fmt.Println("Pruned worktree information")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be pruned without actually pruning")

	return cmd
}
