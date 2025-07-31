package encodecommand

import (
	"errors"
	"fmt"
	"strings"

	encodeservice "github.com/redjax/syst/internal/services/encodeService"
	"github.com/spf13/cobra"
)

func NewEncodeCommand() *cobra.Command {
	var showMethods bool

	// List of supported encoding methods. Keep in sync with encodeService.
	availableMethods := []string{
		"base32", "b32",
		"base64", "b64",
		"hex",
		"url", "urlencode",
		"utf8", "utf",
	}

	cmd := &cobra.Command{
		Use:   "encode [method] [input]",
		Short: "Encode string with various encoding methods.",
		Long: `Available methods (list may be incomplete, run syst encode --methods to see all):
    base32, base64, hex, url, utf8
	
Examples:
  syst encode base64 "Hello, World!"
  syst encode --methods
`,
		Args: func(cmd *cobra.Command, args []string) error {
			if showMethods {
				// Allow 0 args when --methods flag is used
				return nil
			}

			if len(args) != 2 {
				return errors.New("requires exactly 2 arguments: [method] [input]")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if showMethods {
				fmt.Printf("Available encoders: [%s]\n\n", strings.Join(availableMethods, ", "))
				fmt.Println("Usage:")
				fmt.Println(`    $> syst encode [method] [input]`)

				return nil
			}

			method := strings.ToLower(args[0])
			input := args[1]

			svc := encodeservice.NewEncodeService()
			encoded, err := svc.Encode(method, input)
			if err != nil {
				return err
			}

			fmt.Println(encoded)
			return nil
		},
	}

	cmd.Flags().BoolVar(&showMethods, "methods", false, "Show available encoding methods")

	return cmd
}
