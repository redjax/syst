package gitcommand

import (
	"fmt"

	gitservice "github.com/redjax/syst/internal/services/gitService"
	"github.com/spf13/cobra"
)

func NewGitInfoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show information about the current Git repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			info, err := gitservice.GetRepoInfo()
			if err != nil {
				return err
			}

			fmt.Printf("Repository Path:   %s\n", info.Path)
			fmt.Printf("Is Git Repo:       %v\n", info.IsRepo)

			if info.IsRepo {
				fmt.Printf("Current Branch:    %s\n", info.CurrentBranch)
				fmt.Printf("Repo Size (bytes): %d\n", info.SizeBytes)
			}

			if len(info.Remotes) == 0 {
				fmt.Println("No remotes found")
			} else {
				fmt.Println("Remotes:")
				for _, remote := range info.Remotes {
					fmt.Printf("  +%s:\n", remote.Name)
					fmt.Printf("      - Fetch: %s\n", remote.FetchURL)
					fmt.Printf("      - Push:  %s\n", remote.PushURL)
				}
			}

			return nil
		},
	}

	return cmd
}
