package gitcommand

import (
	"fmt"
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
		branch     string
		newBranch  bool
		force      bool
		detach     bool
		noCheckout bool
	)

	cmd := &cobra.Command{
		Use:     "add <path> [<branch>]",
		Aliases: []string{"a", "create", "new"},
		Short:   "Add a new worktree",
		Long:    "Create a new Git worktree at the specified path",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath, _ := cmd.Flags().GetString("repo")

			manager, err := worktreeservice.NewWorktreeManager(repoPath)
			if err != nil {
				return err
			}

			path := args[0]

			// If branch not specified via flag, check args
			if branch == "" && len(args) > 1 {
				branch = args[1]
			}

			// If path looks like a branch name and no explicit branch, swap them
			if branch == "" && !strings.Contains(path, "/") && len(args) == 1 {
				branch = path
				path = manager.GenerateWorktreePath(branch)
			}

			opts := worktreeservice.AddWorktreeOptions{
				Path:      path,
				Branch:    branch,
				NewBranch: newBranch,
				Force:     force,
				Detach:    detach,
				Checkout:  !noCheckout,
			}

			if err := manager.AddWorktree(opts); err != nil {
				return err
			}

			fmt.Printf("Created worktree at %s", path)
			if branch != "" {
				fmt.Printf(" on branch %s", branch)
			}
			fmt.Println()

			return nil
		},
	}

	cmd.Flags().StringVarP(&branch, "branch", "b", "", "Branch to checkout or create")
	cmd.Flags().BoolVarP(&newBranch, "new-branch", "B", false, "Create a new branch")
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
