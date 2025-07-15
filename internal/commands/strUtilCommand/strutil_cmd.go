package strutilcommand

import (
	"fmt"

	"github.com/redjax/syst/internal/utils/strutils"
	"github.com/spf13/cobra"
)

func NewStrUtilCommand() *cobra.Command {
	var (
		toUpper     bool
		toLower     bool
		toTitle     bool
		toCapital   bool
		ignoreCase  bool
		removeList  []string
		replaceList []string
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
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]
			result := target

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

			fmt.Println(result)

			return nil
		},
	}

	cmd.Flags().BoolVar(&toUpper, "upper", false, "Transform string to UPPERCASE")
	cmd.Flags().BoolVar(&toLower, "lower", false, "Transform string to lowercase")
	cmd.Flags().BoolVar(&toTitle, "title", false, "Transform string to Title Case")
	cmd.Flags().BoolVar(&toCapital, "capitalize", false, "Transform string to Capitalized case")
	cmd.Flags().BoolVarP(&ignoreCase, "ignore-case", "i", false, "Perform remove/replace operations case-insensitively")

	cmd.Flags().StringArrayVar(&removeList, "remove", []string{}, "Remove all instances of provided string(s)")
	cmd.Flags().StringArrayVar(&replaceList, "replace", []string{}, "Replace substrings using 'search/replace' syntax (supports multiple)")

	return cmd
}
