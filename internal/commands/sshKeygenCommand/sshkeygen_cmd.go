package sshcommand

import (
	"errors"
	"fmt"
	"strings"

	sshservice "github.com/redjax/syst/internal/services/sshService"
	"github.com/spf13/cobra"
)

func NewSSHKeygenCommand() *cobra.Command {
	var opts sshservice.KeyGenOptions

	cmd := &cobra.Command{
		Use:   "keygen",
		Short: "Generate SSH key pair",
		RunE: func(cmd *cobra.Command, args []string) error {
			privPath, pubPath, err := sshservice.GenerateKey(opts)
			if err != nil {
				if errors.Is(err, sshservice.ErrUserAborted) {
					fmt.Println("Aborted by user.")
					return nil // <- return nil so Cobra does NOT print usage
				}
				return err
			}

			fmt.Printf("Private key saved to %s\n", privPath)
			fmt.Printf("Public key saved to %s\n", pubPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.FilePath, "output-file", "o", "", "Path to save private key (e.g. ~/.ssh/id_rsa)")
	cmd.Flags().StringVarP(&opts.Algorithm, "key-type", "t", "rsa", "Key algorithm: rsa or ed25519")
	cmd.Flags().IntVarP(&opts.Bits, "bits", "b", 4096, "Bit size for RSA keys (ignored for ed25519)")
	cmd.Flags().StringVarP(&opts.Password, "password", "P", "", "Optional password to encrypt private key")
	cmd.Flags().StringVarP(&opts.Comment, "comment", "c", "generated-by-syst", "Optional comment for the public key")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Overwrite existing key files without prompting")

	// Register completion for --key-type / -t
	cmd.RegisterFlagCompletionFunc("key-type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		algorithms := []string{"rsa", "ed25519"}
		var completions []string
		for _, algo := range algorithms {
			if strings.HasPrefix(algo, toComplete) {
				completions = append(completions, algo)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}
