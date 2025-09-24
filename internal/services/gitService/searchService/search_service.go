package searchService

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/redjax/syst/internal/utils/terminal"
)

type SearchOptions struct {
	Query         []string
	SearchCommits bool
	SearchFiles   bool
	SearchContent bool
	SearchAuthors bool
	SearchCurrent bool
	CaseSensitive bool
	MaxResults    int
	SinceDate     string
	UntilDate     string
	AuthorFilter  string
	FileFilter    string
}

type SearchResult struct {
	Type       string // "commit", "file", "content"
	ItemTitle  string
	ItemDesc   string
	Hash       string
	Author     string
	Date       time.Time
	FilePath   string
	LineNumber int
	Content    string
	Commit     *object.Commit
}

func (s SearchResult) Title() string       { return s.ItemTitle }
func (s SearchResult) Description() string { return s.ItemDesc }
func (s SearchResult) FilterValue() string {
	// Return all searchable content in lowercase for case-insensitive filtering
	return strings.ToLower(s.ItemTitle + " " + s.ItemDesc + " " + s.Content + " " + s.Author + " " + s.FilePath)
}

type SearchMode int

const (
	InputMode SearchMode = iota
	ResultsMode
	DetailMode
)

type model struct {
	searchInput    textinput.Model
	resultsList    list.Model
	spinner        spinner.Model
	currentMode    SearchMode
	searchQuery    string
	results        []SearchResult
	selectedResult *SearchResult
	loading        bool
	searchProgress string
	err            error
	tuiHelper      *terminal.ResponsiveTUIHelper
	searchOptions  SearchOptions
}

type searchCompletedMsg struct {
	results []SearchResult
}

type searchProgressMsg struct {
	message string
}

type initialSearchMsg struct {
	query string
}

type errMsg struct {
	err error
}

func (e errMsg) Error() string { return e.err.Error() }

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true).
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	detailStyle = lipgloss.NewStyle().
			Padding(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("39"))

	searchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("211"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	matchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Background(lipgloss.Color("235")).
			Bold(true)
)

func initialModelWithOptions(opts SearchOptions) model {
	searchInput := textinput.New()
	searchInput.Placeholder = "Enter search query (commits, files, content, authors)..."
	searchInput.CharLimit = 256
	searchInput.Focus()

	// If query provided, start with that query
	if len(opts.Query) > 0 {
		query := strings.Join(opts.Query, " ")
		searchInput.SetValue(query)
	}

	resultsList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	resultsList.Title = "Search Results"
	resultsList.SetShowStatusBar(false)
	resultsList.SetFilteringEnabled(true) // Enable built-in filtering

	// Initialize spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	m := model{
		searchInput:   searchInput,
		resultsList:   resultsList,
		spinner:       s,
		currentMode:   InputMode,
		tuiHelper:     terminal.NewResponsiveTUIHelper(),
		searchOptions: opts,
	}

	return m
}

func (m model) Init() tea.Cmd {
	// If we have an initial query, search immediately
	if m.searchInput.Value() != "" {
		query := m.searchInput.Value()
		// Set the search query for display purposes
		// Note: we can't modify m here, so we'll handle this in Update
		return tea.Batch(
			textinput.Blink,
			m.spinner.Tick,
			func() tea.Msg {
				// Send a special message to set the query and start search
				return initialSearchMsg{query: query}
			},
		)
	}
	return textinput.Blink
}

func performAdvancedSearch(query string, options SearchOptions) tea.Msg {
	// This function performs a comprehensive search based on specified options:
	// - Git history (commits, messages, authors)
	// - Historical file names across all commits
	// - File content (both current and historical)
	// - Current filesystem

	var allResults []SearchResult

	repo, err := git.PlainOpen(".")
	if err != nil {
		return errMsg{err}
	}

	// Search based on enabled options
	if options.SearchCommits {
		if commitResults, err := searchCommits(repo, query); err == nil {
			allResults = append(allResults, commitResults...)
		}
	}

	if options.SearchFiles {
		if fileResults, err := searchHistoricalFiles(repo, query); err == nil {
			allResults = append(allResults, fileResults...)
		}
	}

	if options.SearchContent {
		if contentResults, err := searchHistoricalContent(repo, query); err == nil {
			allResults = append(allResults, contentResults...)
		}
	}

	if options.SearchCurrent {
		if currentResults, err := searchCurrentFiles(query); err == nil {
			allResults = append(allResults, currentResults...)
		}
	}

	if options.SearchAuthors {
		if authorResults, err := searchAuthors(repo, query); err == nil {
			allResults = append(allResults, authorResults...)
		}
	}

	return searchCompletedMsg{results: allResults}
}

func searchCommits(repo *git.Repository, query string) ([]SearchResult, error) {
	var results []SearchResult
	queryLower := strings.ToLower(query)

	ref, err := repo.Head()
	if err != nil {
		return results, err
	}

	cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return results, err
	}

	err = cIter.ForEach(func(c *object.Commit) error {
		messageLower := strings.ToLower(c.Message)
		if strings.Contains(messageLower, queryLower) {
			firstLine := strings.Split(c.Message, "\n")[0]
			results = append(results, SearchResult{
				Type:      "commit",
				ItemTitle: fmt.Sprintf("📝 %s", firstLine),
				ItemDesc:  fmt.Sprintf("%s • %s • %s", c.Hash.String()[:8], c.Author.Name, c.Author.When.Format("2006-01-02")),
				Hash:      c.Hash.String(),
				Author:    c.Author.Name,
				Date:      c.Author.When,
				Content:   c.Message,
				Commit:    c,
			})
		}
		return nil
	})

	return results, err
}

func searchAuthors(repo *git.Repository, query string) ([]SearchResult, error) {
	var results []SearchResult
	queryLower := strings.ToLower(query)
	authorCommits := make(map[string][]*object.Commit)

	ref, err := repo.Head()
	if err != nil {
		return results, err
	}

	cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return results, err
	}

	err = cIter.ForEach(func(c *object.Commit) error {
		authorLower := strings.ToLower(c.Author.Name)
		emailLower := strings.ToLower(c.Author.Email)

		if strings.Contains(authorLower, queryLower) || strings.Contains(emailLower, queryLower) {
			key := c.Author.Name + " <" + c.Author.Email + ">"
			authorCommits[key] = append(authorCommits[key], c)
		}
		return nil
	})

	// Create results for matching authors
	for author, commits := range authorCommits {
		results = append(results, SearchResult{
			Type:      "author",
			ItemTitle: fmt.Sprintf("👤 %s", author),
			ItemDesc:  fmt.Sprintf("Author match • %d commits", len(commits)),
			Author:    author,
			Content:   fmt.Sprintf("%d commits", len(commits)),
		})
	}

	return results, err
}

// searchHistoricalFiles searches through file names across all commits in git history
func searchHistoricalFiles(repo *git.Repository, query string) ([]SearchResult, error) {
	var results []SearchResult
	queryLower := strings.ToLower(query)
	seenFiles := make(map[string]bool)

	ref, err := repo.Head()
	if err != nil {
		return results, err
	}

	cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return results, err
	}

	err = cIter.ForEach(func(c *object.Commit) error {
		tree, err := c.Tree()
		if err != nil {
			return nil // Continue with other commits
		}

		_ = tree.Files().ForEach(func(f *object.File) error {
			filenameLower := strings.ToLower(f.Name)
			if strings.Contains(filenameLower, queryLower) && !seenFiles[f.Name] {
				seenFiles[f.Name] = true
				results = append(results, SearchResult{
					Type:      "historical-file",
					ItemTitle: fmt.Sprintf("📁 %s", f.Name),
					ItemDesc:  fmt.Sprintf("Historical file • Found in commit %s", c.Hash.String()[:8]),
					FilePath:  f.Name,
					Hash:      c.Hash.String(),
					Date:      c.Author.When,
					Content:   f.Name,
				})
			}
			return nil
		})
		return nil
	})

	return results, err
}

// searchHistoricalContent searches through file content across git history
func searchHistoricalContent(repo *git.Repository, query string) ([]SearchResult, error) {
	var results []SearchResult
	queryLower := strings.ToLower(query)
	regex, _ := regexp.Compile("(?i)" + regexp.QuoteMeta(query))

	ref, err := repo.Head()
	if err != nil {
		return results, err
	}

	cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return results, err
	}

	// Limit to recent commits to avoid too much processing
	commitCount := 0
	maxCommits := 100

	err = cIter.ForEach(func(c *object.Commit) error {
		if commitCount >= maxCommits {
			return fmt.Errorf("reached commit limit") // Stop iteration
		}
		commitCount++

		tree, err := c.Tree()
		if err != nil {
			return nil
		}

		_ = tree.Files().ForEach(func(f *object.File) error {
			// Skip large files and binary files
			if f.Size > 512*1024 { // 512KB limit
				return nil
			}

			content, err := f.Contents()
			if err != nil || strings.Contains(content, "\x00") {
				return nil // Skip binary files
			}

			contentLower := strings.ToLower(content)
			if strings.Contains(contentLower, queryLower) {
				lines := strings.Split(content, "\n")
				for i, line := range lines {
					lineLower := strings.ToLower(line)
					if strings.Contains(lineLower, queryLower) {
						highlightedLine := line
						if regex != nil {
							highlightedLine = regex.ReplaceAllStringFunc(line, func(match string) string {
								return matchStyle.Render(match)
							})
						}

						results = append(results, SearchResult{
							Type:       "historical-content",
							ItemTitle:  fmt.Sprintf("🔍 %s:%d (commit %s)", f.Name, i+1, c.Hash.String()[:8]),
							ItemDesc:   fmt.Sprintf("Historical content • Line %d • %s", i+1, c.Author.When.Format("2006-01-02")),
							FilePath:   f.Name,
							LineNumber: i + 1,
							Hash:       c.Hash.String(),
							Date:       c.Author.When,
							Content:    strings.TrimSpace(highlightedLine),
						})

						// Limit results per file
						return nil
					}
				}
			}
			return nil
		})
		return nil
	})

	return results, err
}

// searchCurrentFiles searches through current filesystem files
func searchCurrentFiles(query string) ([]SearchResult, error) {
	var results []SearchResult
	queryLower := strings.ToLower(query)
	regex, _ := regexp.Compile("(?i)" + regexp.QuoteMeta(query))

	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Continue walking
		}

		// Skip hidden directories and files, and common ignore patterns
		// But don't skip the current directory "."
		if (strings.HasPrefix(d.Name(), ".") && d.Name() != ".") ||
			strings.Contains(path, "node_modules") ||
			strings.Contains(path, "vendor") ||
			strings.Contains(path, "dist") ||
			strings.Contains(path, "build") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			return nil
		}

		// Check filename match
		filenameLower := strings.ToLower(d.Name())
		if strings.Contains(filenameLower, queryLower) {
			results = append(results, SearchResult{
				Type:      "current-file",
				ItemTitle: fmt.Sprintf("📄 %s", path),
				ItemDesc:  fmt.Sprintf("Current file match • %s", d.Name()),
				FilePath:  path,
				Content:   path,
			})
		}

		// Check file content for text files
		if isTextFile(path) {
			content, err := os.ReadFile(path)
			if err != nil || len(content) > 1024*1024 { // 1MB limit
				return nil
			}

			contentStr := string(content)
			if strings.Contains(contentStr, "\x00") {
				return nil // Skip binary files
			}

			contentLower := strings.ToLower(contentStr)
			if strings.Contains(contentLower, queryLower) {
				lines := strings.Split(contentStr, "\n")
				for i, line := range lines {
					lineLower := strings.ToLower(line)
					if strings.Contains(lineLower, queryLower) {
						highlightedLine := line
						if regex != nil {
							highlightedLine = regex.ReplaceAllStringFunc(line, func(match string) string {
								return matchStyle.Render(match)
							})
						}

						results = append(results, SearchResult{
							Type:       "current-content",
							ItemTitle:  fmt.Sprintf("🔍 %s:%d", path, i+1),
							ItemDesc:   fmt.Sprintf("Current file content • Line %d", i+1),
							FilePath:   path,
							LineNumber: i + 1,
							Content:    strings.TrimSpace(highlightedLine),
						})

						// Limit results per file
						break
					}
				}
			}
		}

		return nil
	})

	return results, err
}

// isTextFile determines if a file is likely a text file based on extension
func isTextFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	textExtensions := []string{
		".go", ".py", ".js", ".ts", ".java", ".c", ".cpp", ".h", ".hpp",
		".txt", ".md", ".json", ".yaml", ".yml", ".xml", ".html", ".css",
		".sh", ".bat", ".ps1", ".sql", ".conf", ".ini", ".cfg", ".log",
		".rs", ".rb", ".php", ".pl", ".lua", ".vim", ".r", ".scala",
	}

	for _, textExt := range textExtensions {
		if ext == textExt {
			return true
		}
	}
	return false
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.tuiHelper.HandleWindowSizeMsg(msg)
		m.searchInput.Width = msg.Width - 4
		m.resultsList.SetWidth(m.tuiHelper.GetWidth())
		m.resultsList.SetHeight(m.tuiHelper.GetHeight() - 8)
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case initialSearchMsg:
		m.loading = true
		m.searchQuery = msg.query
		return m, tea.Batch(
			m.spinner.Tick,
			func() tea.Msg {
				return performAdvancedSearch(msg.query, m.searchOptions)
			},
		)

	case searchProgressMsg:
		m.searchProgress = msg.message
		return m, nil

	case searchCompletedMsg:
		m.loading = false
		m.searchProgress = ""
		m.results = msg.results

		// Convert to list items
		items := make([]list.Item, len(msg.results))
		for i, result := range msg.results {
			items[i] = result
		}
		m.resultsList.SetItems(items)

		if len(msg.results) > 0 {
			m.currentMode = ResultsMode
		}
		return m, nil

	case errMsg:
		m.loading = false
		m.searchProgress = ""
		m.err = msg.err
		return m, nil

	case tea.KeyMsg:
		switch m.currentMode {
		case InputMode:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "enter":
				if m.searchInput.Value() != "" {
					m.loading = true
					m.searchQuery = m.searchInput.Value()
					return m, tea.Batch(
						m.spinner.Tick,
						func() tea.Msg {
							return performAdvancedSearch(m.searchQuery, m.searchOptions)
						},
					)
				}
			default:
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				return m, cmd
			}

		case ResultsMode:
			// If we're in filter mode, let the list handle all input except esc
			if m.resultsList.FilterState() == list.Filtering {
				switch msg.String() {
				case "q", "ctrl+c":
					return m, tea.Quit
				case "esc":
					// Exit filter mode but stay in results
					var cmd tea.Cmd
					m.resultsList, cmd = m.resultsList.Update(msg)
					return m, cmd
				default:
					// Let the list handle all other input for filtering
					var cmd tea.Cmd
					m.resultsList, cmd = m.resultsList.Update(msg)
					return m, cmd
				}
			}

			// Normal results mode (not filtering)
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "esc":
				// Go back to input mode
				m.currentMode = InputMode
				m.searchInput.Focus()
				return m, nil
			case "enter":
				if selected := m.resultsList.SelectedItem(); selected != nil {
					if result, ok := selected.(SearchResult); ok {
						m.selectedResult = &result
						m.currentMode = DetailMode
					}
				}
				return m, nil
			case "n":
				// New search
				m.currentMode = InputMode
				m.searchInput.SetValue("")
				m.searchInput.Focus()
				return m, nil
			default:
				var cmd tea.Cmd
				m.resultsList, cmd = m.resultsList.Update(msg)
				return m, cmd
			}

		case DetailMode:
			switch msg.String() {
			case "esc", "q":
				m.currentMode = ResultsMode
				m.selectedResult = nil
				return m, nil
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.loading {
		loadingText := fmt.Sprintf("%s Searching...", m.spinner.View())
		if m.searchProgress != "" {
			loadingText += fmt.Sprintf("\n%s", statusStyle.Render(m.searchProgress))
		}
		return loadingText
	}

	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	switch m.currentMode {
	case InputMode:
		return fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			titleStyle.Render("🔍 Advanced Repository Search"),
			searchStyle.Render("Search: "+m.searchInput.View()),
			helpStyle.Render("enter: search • q: quit"),
		)

	case DetailMode:
		if m.selectedResult == nil {
			return "No result selected"
		}
		return m.renderResultDetail(*m.selectedResult)

	default: // ResultsMode
		// Check if we're in filter mode
		filterHelp := ""
		if m.resultsList.FilterState() == list.Filtering {
			filterHelp = " • filtering: type to filter, esc to exit filter"
		} else {
			filterHelp = " • /: filter results"
		}

		help := fmt.Sprintf("Found %d results for '%s' • enter: details • n: new search • esc: back%s • q: quit",
			len(m.results), m.searchQuery, filterHelp)

		return fmt.Sprintf(
			"%s\n%s",
			m.resultsList.View(),
			helpStyle.Render(help),
		)
	}
}

func (m model) renderResultDetail(result SearchResult) string {
	var details strings.Builder

	details.WriteString(titleStyle.Render(fmt.Sprintf("🔍 %s Details", strings.ToUpper(result.Type))))
	details.WriteString("\n\n")

	switch result.Type {
	case "commit":
		details.WriteString(m.renderCommitDetail(result))
	case "file", "historical-file":
		details.WriteString(m.renderFileDetail(result))
	case "content", "historical-content":
		details.WriteString(m.renderContentDetail(result))
	case "current-file":
		details.WriteString(m.renderCurrentFileDetail(result))
	case "current-content":
		details.WriteString(m.renderCurrentContentDetail(result))
	case "author":
		details.WriteString(m.renderAuthorDetail(result))
	default:
		details.WriteString(fmt.Sprintf("Type: %s\nContent: %s", result.Type, result.Content))
	}

	details.WriteString("\n\n")
	details.WriteString(helpStyle.Render("esc: back to results • q: quit"))

	return details.String()
}

func (m model) renderCommitDetail(result SearchResult) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("📝 Hash: %s\n", result.Hash))
	content.WriteString(fmt.Sprintf("👤 Author: %s\n", result.Author))
	content.WriteString(fmt.Sprintf("📅 Date: %s\n\n", result.Date.Format("2006-01-02 15:04:05")))

	content.WriteString("💬 Message:\n")
	content.WriteString(detailStyle.Render(result.Content))

	if result.Commit != nil {
		content.WriteString("\n\n📋 Changes:\n")
		if diff := m.getCommitDiff(result.Commit); diff != "" {
			content.WriteString(detailStyle.Render(diff))
		} else {
			content.WriteString(statusStyle.Render("Unable to retrieve diff"))
		}
	}

	return content.String()
}

func (m model) renderFileDetail(result SearchResult) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("📁 File: %s\n", result.FilePath))
	if result.Hash != "" {
		content.WriteString(fmt.Sprintf("📝 Commit: %s\n", result.Hash))
		content.WriteString(fmt.Sprintf("📅 Date: %s\n", result.Date.Format("2006-01-02 15:04:05")))
	}
	content.WriteString("\n")

	if fileContent := m.getFileContent(result); fileContent != "" {
		content.WriteString("📄 File Preview:\n")
		content.WriteString(detailStyle.Render(fileContent))
	}

	return content.String()
}

func (m model) renderContentDetail(result SearchResult) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("🔍 File: %s\n", result.FilePath))
	content.WriteString(fmt.Sprintf("📍 Line: %d\n", result.LineNumber))
	if result.Hash != "" {
		content.WriteString(fmt.Sprintf("📝 Commit: %s\n", result.Hash))
		content.WriteString(fmt.Sprintf("📅 Date: %s\n", result.Date.Format("2006-01-02 15:04:05")))
	}
	content.WriteString("\n")

	if contextContent := m.getContentWithContext(result); contextContent != "" {
		content.WriteString("📄 Content with Context:\n")
		content.WriteString(detailStyle.Render(contextContent))
	} else {
		content.WriteString("📝 Matched Line:\n")
		content.WriteString(detailStyle.Render(result.Content))
	}

	return content.String()
}

func (m model) renderCurrentFileDetail(result SearchResult) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("📄 Current File: %s\n\n", result.FilePath))

	if info, err := os.Stat(result.FilePath); err == nil {
		content.WriteString(fmt.Sprintf("📏 Size: %d bytes\n", info.Size()))
		content.WriteString(fmt.Sprintf("📅 Modified: %s\n\n", info.ModTime().Format("2006-01-02 15:04:05")))
	}

	if fileContent := m.getCurrentFileContent(result.FilePath); fileContent != "" {
		content.WriteString("📄 File Preview:\n")
		content.WriteString(detailStyle.Render(fileContent))
	}

	return content.String()
}

func (m model) renderCurrentContentDetail(result SearchResult) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("🔍 Current File: %s\n", result.FilePath))
	content.WriteString(fmt.Sprintf("📍 Line: %d\n\n", result.LineNumber))

	if contextContent := m.getCurrentContentWithContext(result); contextContent != "" {
		content.WriteString("📄 Content with Context:\n")
		content.WriteString(detailStyle.Render(contextContent))
	} else {
		content.WriteString("📝 Matched Line:\n")
		content.WriteString(detailStyle.Render(result.Content))
	}

	return content.String()
}

func (m model) renderAuthorDetail(result SearchResult) string {
	return fmt.Sprintf("👤 Author: %s\n📊 %s", result.Author, result.Content)
}

func (m model) getCommitDiff(commit *object.Commit) string {
	if commit == nil {
		return ""
	}

	parents := commit.Parents()
	parent, err := parents.Next()
	if err != nil {
		return "Initial commit - no parent to diff against"
	}

	parentTree, err := parent.Tree()
	if err != nil {
		return ""
	}

	currentTree, err := commit.Tree()
	if err != nil {
		return ""
	}

	var diff strings.Builder
	changes, err := parentTree.Diff(currentTree)
	if err != nil {
		return ""
	}

	for _, change := range changes {
		action, err := change.Action()
		if err != nil {
			continue
		}
		switch action {
		case 0: // Insert
			diff.WriteString(fmt.Sprintf("+ %s (added)\n", change.To.Name))
		case 1: // Delete
			diff.WriteString(fmt.Sprintf("- %s (deleted)\n", change.From.Name))
		case 2: // Modify
			diff.WriteString(fmt.Sprintf("~ %s (modified)\n", change.To.Name))
		}
	}

	return diff.String()
}

func (m model) getFileContent(result SearchResult) string {
	if result.Hash == "" {
		return ""
	}

	repo, err := git.PlainOpen(".")
	if err != nil {
		return ""
	}

	hash, err := repo.ResolveRevision(plumbing.Revision(result.Hash))
	if err != nil {
		return ""
	}

	commit, err := repo.CommitObject(*hash)
	if err != nil {
		return ""
	}

	file, err := commit.File(result.FilePath)
	if err != nil {
		return ""
	}

	content, err := file.Contents()
	if err != nil {
		return ""
	}

	lines := strings.Split(content, "\n")
	if len(lines) > 50 {
		lines = lines[:50]
		lines = append(lines, "... (truncated)")
	}

	return strings.Join(lines, "\n")
}

func (m model) getCurrentFileContent(filepath string) string {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return ""
	}

	if strings.Contains(string(content), "\x00") {
		return "[Binary file]"
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) > 50 {
		lines = lines[:50]
		lines = append(lines, "... (truncated)")
	}

	return strings.Join(lines, "\n")
}

func (m model) getContentWithContext(result SearchResult) string {
	if result.Hash == "" || result.LineNumber == 0 {
		return ""
	}

	repo, err := git.PlainOpen(".")
	if err != nil {
		return ""
	}

	hash, err := repo.ResolveRevision(plumbing.Revision(result.Hash))
	if err != nil {
		return ""
	}

	commit, err := repo.CommitObject(*hash)
	if err != nil {
		return ""
	}

	file, err := commit.File(result.FilePath)
	if err != nil {
		return ""
	}

	content, err := file.Contents()
	if err != nil {
		return ""
	}

	return m.extractContextLines(content, result.LineNumber, 5)
}

func (m model) getCurrentContentWithContext(result SearchResult) string {
	if result.LineNumber == 0 {
		return ""
	}

	content, err := os.ReadFile(result.FilePath)
	if err != nil {
		return ""
	}

	return m.extractContextLines(string(content), result.LineNumber, 5)
}

func (m model) extractContextLines(content string, lineNumber, contextLines int) string {
	lines := strings.Split(content, "\n")
	if lineNumber > len(lines) {
		return ""
	}

	start := max(0, lineNumber-contextLines-1)
	end := min(len(lines), lineNumber+contextLines)

	var result strings.Builder
	for i := start; i < end; i++ {
		lineNum := i + 1
		line := lines[i]

		if lineNum == lineNumber {
			result.WriteString(fmt.Sprintf(">>> %3d: %s\n", lineNum, line))
		} else {
			result.WriteString(fmt.Sprintf("    %3d: %s\n", lineNum, line))
		}
	}

	return result.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func RunAdvancedSearch(args []string) error {
	// Default options for backward compatibility
	opts := SearchOptions{
		Query:         args,
		SearchCommits: true,
		SearchFiles:   true,
		SearchContent: true,
		SearchAuthors: true,
		SearchCurrent: true,
		MaxResults:    100,
	}
	return RunAdvancedSearchWithOptions(opts)
}

func RunAdvancedSearchWithOptions(opts SearchOptions) error {
	p := tea.NewProgram(initialModelWithOptions(opts), tea.WithAltScreen())
	_, err := p.Run()
	if err != nil {
		fmt.Printf("Error running search: %v\n", err)
		os.Exit(1)
	}
	return nil
}
