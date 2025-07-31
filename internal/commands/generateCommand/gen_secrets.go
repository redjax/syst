package generatecommand

import (
	"fmt"

	secretsservice "github.com/redjax/syst/internal/services/secretsService"
	"github.com/spf13/cobra"
)

func NewSecretsCommand() *cobra.Command {
	var length int
	var showMethods bool

	// List of supported methods (this should match internal/services/secretsService's methods)
	availableMethods := []string{
		"sha256",
		"md5",
		"openssl",
		"openssl-hex",
		"openssl-safe",
	}

	cmd := &cobra.Command{
		Use:   "secret [method]",
		Short: "Generate secrets with different encoding or hashing methods",
		Long: `Generate secrets using various methods. Run with --methods to see all available generators.

Examples:
  appname generate secrets base64 --length 32
  appname generate secrets sha256 --length 64
`,
		Args: func(cmd *cobra.Command, args []string) error {
			showMethods, _ := cmd.Flags().GetBool("methods")
			if showMethods {
				// If --methods is specified, no positional args required
				return nil
			}
			// else require exactly one positional argument (method)
			if len(args) != 1 {
				return fmt.Errorf("accepts 1 arg(s), received %d", len(args))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if showMethods {
				// Print the list and return quietly
				fmt.Printf("Generate secret methods:\n%v\n\n", availableMethods)
				fmt.Printf("Usage: \n  $> syst gen secret [method]\n")

				return nil
			}

			svc := secretsservice.NewSecretsService()
			result, err := svc.GenerateSecret(args[0], length)

			if err != nil {
				return err
			}

			fmt.Println(result)

			return nil
		},
	}

	cmd.Flags().BoolVar(&showMethods, "methods", false, "Show available secret generation methods")
	cmd.Flags().IntVarP(&length, "length", "l", 32, "length of the generated secret in bytes")

	return cmd
}
