package gitcommand

import (
	"github.com/redjax/syst/internal/services/gitService/searchService"
	"github.com/spf13/cobra"
)

// NewGitSearchCommand creates the git search command
func NewGitSearchCommand() *cobra.Command {
	var (
		searchCommits bool
		searchFiles   bool
		searchContent bool
		searchAuthors bool
		searchCurrent bool
		caseSensitive bool
		maxResults    int
		sinceDate     string
		untilDate     string
		authorFilter  string
		fileFilter    string
	)

	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Advanced repository search",
		Long: `Search commits, authors, files, and content across repository history with interactive results.

Examples:
  syst git search "bug fix"                    # Search all types for "bug fix"
  syst git search --commits "refactor"         # Search only commit messages
  syst git search --files "config"             # Search only file names
  syst git search --content "TODO"             # Search only file content
  syst git search --authors "john"             # Search only author names
  syst git search --current "readme"           # Search only current files
  syst git search --since "2024-01-01" "fix"   # Search since specific date
  syst git search --author "john" --files      # Combine filters

The search supports:
- Commit messages and metadata
- Historical file names across all commits  
- File content (both current and historical)
- Author names and emails
- Current filesystem files

Interactive commands in TUI:
- enter: view details
- n: new search
- esc: back to search input
- /: filter results (esc to exit filter)
- q: quit`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no search type flags are specified, enable all search types by default
			noSearchTypeFlags := !searchCommits && !searchFiles && !searchContent && !searchAuthors && !searchCurrent
			if noSearchTypeFlags {
				searchCommits = true
				searchFiles = true
				searchContent = true
				searchAuthors = true
				searchCurrent = true
			}

			opts := searchService.SearchOptions{
				Query:         args,
				SearchCommits: searchCommits,
				SearchFiles:   searchFiles,
				SearchContent: searchContent,
				SearchAuthors: searchAuthors,
				SearchCurrent: searchCurrent,
				CaseSensitive: caseSensitive,
				MaxResults:    maxResults,
				SinceDate:     sinceDate,
				UntilDate:     untilDate,
				AuthorFilter:  authorFilter,
				FileFilter:    fileFilter,
			}
			return searchService.RunAdvancedSearchWithOptions(opts)
		},
	}

	// Search type flags
	cmd.Flags().BoolVar(&searchCommits, "commits", false, "Search commit messages and metadata only")
	cmd.Flags().BoolVar(&searchFiles, "files", false, "Search file names only (historical and current)")
	cmd.Flags().BoolVar(&searchContent, "content", false, "Search file content only (historical and current)")
	cmd.Flags().BoolVar(&searchAuthors, "authors", false, "Search author names and emails only")
	cmd.Flags().BoolVar(&searchCurrent, "current", false, "Search current filesystem files only")

	// Filter flags
	cmd.Flags().BoolVar(&caseSensitive, "case-sensitive", false, "Perform case-sensitive search")
	cmd.Flags().IntVar(&maxResults, "max-results", 100, "Maximum number of results to return per search type")
	cmd.Flags().StringVar(&sinceDate, "since", "", "Search commits since date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&untilDate, "until", "", "Search commits until date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&authorFilter, "author", "", "Filter results by author name/email")
	cmd.Flags().StringVar(&fileFilter, "file-pattern", "", "Filter file results by pattern (supports wildcards)")

	return cmd
}
