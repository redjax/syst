package gitcommand

import (
	gitservice "github.com/redjax/syst/internal/services/gitService"
	"github.com/spf13/cobra"
)

func NewGitSparseCloneCommand() *cobra.Command {
	var opts gitservice.SparseCloneOptions

	cmd := &cobra.Command{
		Use:   "sparse-clone",
		Short: "Clone a git repo with sparse checkout in one step",
		RunE: func(cmd *cobra.Command, args []string) error {
			return gitservice.SparseClone(opts)
		},
	}

	cmd.Flags().StringVar(&opts.Provider, "provider", "github", "Git provider (github, gitlab, codeberg)")
	cmd.Flags().StringVarP(&opts.User, "username", "u", "", "Git username or org")
	cmd.Flags().StringVarP(&opts.Repository, "repository", "r", "", "Repository name")
	cmd.Flags().StringVarP(&opts.Output, "output-dir", "o", "", "Output directory (defaults to repo name)")
	cmd.Flags().StringVarP(&opts.Branch, "checkout-branch", "b", "main", "Branch name to checkout")
	cmd.Flags().StringSliceVarP(&opts.Paths, "checkout-path", "p", []string{}, "Paths to sparse-checkout (repeatable)")
	cmd.Flags().StringVar(&opts.Protocol, "protocol", "ssh", "Clone protocol: ssh or https")

	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("repository")
	cmd.MarkFlagRequired("checkout-path")

	return cmd
}
