package sshcommand

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	sshservice "github.com/redjax/syst/internal/services/sshService"
	"github.com/redjax/syst/internal/utils/path"
	"github.com/spf13/cobra"
)

// NewSSHKeygenCommand creates the `ssh keygen` subcommand
func NewSSHKeygenCommand() *cobra.Command {
	var (
		filePath  string
		algorithm string
		bits      int
		password  string
		comment   string
	)

	cmd := &cobra.Command{
		Use:   "keygen",
		Short: "Generate an SSH key pair",
		Long: `Generate SSH key pairs using RSA or Ed25519 algorithms.
You can specify bit size (for RSA), optional password encryption, 
an optional comment, and the output file path.`,
		Example: `  syst ssh keygen -a rsa -b 4096 -f ~/.ssh/id_rsa -c "my key"
  syst ssh keygen -a ed25519 -f ~/.ssh/id_ed25519 -P "secret"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate file path
			if filePath == "" {
				return errors.New("file path is required (use -f or --file)")
			}

			expandedPath, err := path.ExpandPath(filePath)
			if err != nil {
				return fmt.Errorf("failed to resolve path: %w", err)
			}
			filePath = expandedPath

			dir := filepath.Dir(filePath)
			if err := os.MkdirAll(dir, 0700); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}

			// Validate algorithm
			switch algorithm {
			case "rsa", "ed25519":
				// valid
			default:
				return fmt.Errorf("unsupported algorithm: %s (must be rsa or ed25519)", algorithm)
			}

			// Set default bit size for RSA
			if algorithm == "rsa" && bits < 2048 {
				return fmt.Errorf("insecure RSA key size: %d (must be >= 2048)", bits)
			}
			if algorithm == "ed25519" && bits != 0 {
				fmt.Printf("[WARN] bit size ignored for Ed25519 keys\n")
			}

			opts := sshservice.KeyGenOptions{
				FilePath:  filePath,
				Algorithm: algorithm,
				Bits:      bits,
				Password:  password,
				Comment:   comment,
			}

			// Generate keys
			privKey, pubKey, err := sshservice.GenerateKey(opts)
			if err != nil {
				return fmt.Errorf("failed to generate key: %w", err)
			}

			// Write private key
			if err := os.WriteFile(filePath, []byte(privKey), 0600); err != nil {
				return fmt.Errorf("failed to save private key: %w", err)
			}
			fmt.Printf("Private key saved to %s\n", filePath)

			// Write public key
			pubPath := filePath + ".pub"
			if err := os.WriteFile(pubPath, []byte(pubKey), 0644); err != nil {
				return fmt.Errorf("failed to save public key: %w", err)
			}
			fmt.Printf("Public key saved to %s\n", pubPath)

			return nil
		},
	}

	// Get user's home dir
	usrHome, _ := os.UserHomeDir()
	// Set default SSH key file path to ~/.ssh/id_rsa
	defaultSshKeyPath := filepath.Join(usrHome, ".ssh", "id_rsa")

	cmd.Flags().StringVarP(
		&filePath,
		"output-file",
		"o",
		defaultSshKeyPath,
		fmt.Sprintf("Full path to save private key (default: %s)", defaultSshKeyPath),
	)
	cmd.Flags().StringVarP(&algorithm, "algorithm", "a", "rsa", "Key algorithm: rsa or ed25519")
	cmd.Flags().IntVarP(&bits, "bits", "b", 4096, "Bit size for RSA key (ignored for Ed25519)")
	cmd.Flags().StringVarP(&password, "password", "P", "", "Optional password to encrypt private key")
	cmd.Flags().StringVarP(&comment, "comment", "c", "generated-by-syst", "Optional comment added to public key")

	return cmd
}
