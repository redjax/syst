package gitcommand

import (
	gitservice "github.com/redjax/syst/internal/services/gitService"
	sparsecloneservice "github.com/redjax/syst/internal/services/gitService/sparseCloneService"
	"github.com/spf13/cobra"
)

func NewGitSparseCloneCommand() *cobra.Command {
	var opts gitservice.SparseCloneOptions

	cmd := &cobra.Command{
		Use:   "sparse-clone",
		Short: "Clone a git repo with sparse checkout in one step",
		Long: `Clone a git repository with sparse checkout in one step.

If no flags are provided, an interactive TUI will guide you through the configuration.
Otherwise, use the flags to specify the clone options directly.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if required flags are provided
			userFlag := cmd.Flag("username")
			repoFlag := cmd.Flag("repository")
			pathsFlag := cmd.Flag("checkout-path")

			userProvided := userFlag != nil && userFlag.Changed
			repoProvided := repoFlag != nil && repoFlag.Changed
			pathsProvided := pathsFlag != nil && pathsFlag.Changed

			// If no required flags are provided, launch TUI
			if !userProvided && !repoProvided && !pathsProvided {
				tuiOpts, err := sparsecloneservice.RunSparseCloneTUI()
				if err != nil {
					return err
				}
				return gitservice.SparseClone(*tuiOpts)
			}

			// Validate that all required flags are provided when using CLI mode
			if !userProvided {
				return cmd.Help()
			}
			if !repoProvided {
				return cmd.Help()
			}
			if !pathsProvided {
				return cmd.Help()
			}

			// Use the provided flags
			return gitservice.SparseClone(opts)
		},
	}

	cmd.Flags().StringVar(&opts.Provider, "provider", "github", "Git provider (github, gitlab, codeberg)")
	cmd.Flags().StringVarP(&opts.User, "username", "u", "", "Git username or org (required)")
	cmd.Flags().StringVarP(&opts.Repository, "repository", "r", "", "Repository name (required)")
	cmd.Flags().StringVarP(&opts.Output, "output-dir", "o", "", "Output directory (defaults to repo name)")
	cmd.Flags().StringVarP(&opts.Branch, "checkout-branch", "b", "main", "Branch name to checkout")
	cmd.Flags().StringSliceVarP(&opts.Paths, "checkout-path", "p", []string{}, "Paths to sparse-checkout (required, repeatable)")
	cmd.Flags().StringVar(&opts.Protocol, "protocol", "ssh", "Clone protocol: ssh or https")

	return cmd
}
