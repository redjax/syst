package strutilcommand

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/redjax/syst/internal/utils/strutils"
	"github.com/spf13/cobra"
)

func NewStrUtilCommand() *cobra.Command {
	var (
		toUpper      bool
		toLower      bool
		toTitle      bool
		toCapital    bool
		ignoreCase   bool
		searchString string
		filePath     string
		removeList   []string
		replaceList  []string
	)

	cmd := &cobra.Command{
		Use:   "strutil [string]",
		Short: "Manipulate an input string.",
		Long: `Perform transformations on an input string, such as changing its case, 
removing specific substrings, or replacing them with others.

Examples:
  strutil "Hello World" --upper
  strutil "Hello World" --remove "ll" --replace "World/Earth"
  strutil "hello world" --title
`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var input string

			switch {
			case filePath != "":
				// Read from file
				fmt.Printf("Results from file '%s':\n\n", filePath)

				content, err := os.ReadFile(filePath)
				if err != nil {
					return fmt.Errorf("failed to read file: %w", err)
				}

				input = string(content)

			case len(args) > 0:
				// Use CLI argument
				input = args[0]

			default:
				// Use stdin
				fmt.Printf("Results from stdin:\n")

				inBytes, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return fmt.Errorf("failed to read from stdin: %w", err)
				}

				input = string(inBytes)
			}

			// Ensure stdin is not empty
			if strings.TrimSpace(input) == "" {
				return errors.New("input is empty")
			}

			// Then assign result for further processing
			result := input

			// Only do search if the --search flag is used
			if searchString != "" {
				matches := strutils.FindMatchingLines(input, searchString, ignoreCase)

				if len(matches) == 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "No results for '%s'\n", searchString)
				} else {
					for _, line := range matches {
						fmt.Fprintln(cmd.OutOrStdout(), line)
					}
				}

				return nil
			}

			// Apply removals using utility function
			if len(removeList) > 0 {
				result = strutils.RemoveSubstrings(result, removeList, ignoreCase)
			}

			// Apply replacements using utility function
			if len(replaceList) > 0 {
				var warnings []string
				result, warnings = strutils.ReplaceSubstrings(result, replaceList, ignoreCase)
				for _, bad := range warnings {
					fmt.Fprintf(cmd.ErrOrStderr(), "[warn] Invalid --replace format: '%s'. Use 'search/replace'\n", bad)
				}
			}

			// Case transformations
			if toUpper {
				result = strutils.ToUpper(result)
			}
			if toLower {
				result = strutils.ToLower(result)
			}
			if toTitle {
				result = strutils.ToTitleCase(result)
			}
			if toCapital {
				result = strutils.Capitalize(result)
			}

			// grep-like substring search is --search is present
			if searchString != "" {
				found := strutils.SearchSubstring(result, searchString, ignoreCase)
				if !found {
					return nil
				}
			}

			fmt.Println(result)

			return nil
		},
	}

	cmd.Flags().BoolVar(&toUpper, "upper", false, "Transform string to UPPERCASE")
	cmd.Flags().BoolVar(&toLower, "lower", false, "Transform string to lowercase")
	cmd.Flags().BoolVar(&toTitle, "title", false, "Transform string to Title Case")
	cmd.Flags().BoolVar(&toCapital, "capitalize", false, "Transform string to Capitalized case")
	cmd.Flags().BoolVarP(&ignoreCase, "ignore-case", "i", false, "Perform remove/replace operations case-insensitively")

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to input file (overrides argument or stdin)")
	cmd.Flags().StringVarP(&searchString, "search", "s", "", "Search for a substring (like grep)")
	cmd.Flags().StringArrayVar(&removeList, "remove", []string{}, "Remove all instances of provided string(s)")
	cmd.Flags().StringArrayVar(&replaceList, "replace", []string{}, "Replace substrings using 'search/replace' syntax (supports multiple)")

	return cmd
}
