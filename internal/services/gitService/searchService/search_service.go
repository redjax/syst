package searchService

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

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
func (s SearchResult) FilterValue() string { return s.ItemTitle + " " + s.ItemDesc + " " + s.Content }

type SearchMode int

const (
	InputMode SearchMode = iota
	ResultsMode
	DetailMode
)

type SearchType int

const (
	CommitSearch SearchType = iota
	FileSearch
	ContentSearch
	AuthorSearch
)

type model struct {
	searchInput    textinput.Model
	resultsList    list.Model
	currentMode    SearchMode
	searchType     SearchType
	searchQuery    string
	results        []SearchResult
	selectedResult *SearchResult
	loading        bool
	err            error
	width          int
	height         int
}

type searchCompletedMsg struct {
	results []SearchResult
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

	typeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("99")).
			Bold(true)

	matchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Background(lipgloss.Color("235")).
			Bold(true)
)

func initialModel(args []string) model {
	searchInput := textinput.New()
	searchInput.Placeholder = "Enter search query (commits, files, content, authors)..."
	searchInput.CharLimit = 256
	searchInput.Focus()

	// If args provided, start with that query
	if len(args) > 0 {
		query := strings.Join(args, " ")
		searchInput.SetValue(query)
	}

	resultsList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	resultsList.Title = "Search Results"
	resultsList.SetShowStatusBar(false)
	resultsList.SetShowHelp(false)

	return model{
		searchInput: searchInput,
		resultsList: resultsList,
		currentMode: InputMode,
		searchType:  CommitSearch,
	}
}

func (m model) Init() tea.Cmd {
	// If we have an initial query, search immediately
	if m.searchInput.Value() != "" {
		return tea.Batch(
			textinput.Blink,
			func() tea.Msg {
				return performSearch(m.searchInput.Value(), CommitSearch)
			},
		)
	}
	return textinput.Blink
}

func performSearch(query string, searchType SearchType) tea.Msg {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return errMsg{err}
	}

	var results []SearchResult

	switch searchType {
	case CommitSearch:
		results, err = searchCommits(repo, query)
	case FileSearch:
		results, err = searchFiles(repo, query)
	case ContentSearch:
		results, err = searchContent(repo, query)
	case AuthorSearch:
		results, err = searchAuthors(repo, query)
	default:
		// Search all types
		commitResults, _ := searchCommits(repo, query)
		fileResults, _ := searchFiles(repo, query)
		contentResults, _ := searchContent(repo, query)
		authorResults, _ := searchAuthors(repo, query)

		results = append(results, commitResults...)
		results = append(results, fileResults...)
		results = append(results, contentResults...)
		results = append(results, authorResults...)
	}

	if err != nil {
		return errMsg{err}
	}

	return searchCompletedMsg{results: results}
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

func searchFiles(repo *git.Repository, query string) ([]SearchResult, error) {
	var results []SearchResult
	queryLower := strings.ToLower(query)

	ref, err := repo.Head()
	if err != nil {
		return results, err
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return results, err
	}

	tree, err := commit.Tree()
	if err != nil {
		return results, err
	}

	err = tree.Files().ForEach(func(f *object.File) error {
		filenameLower := strings.ToLower(f.Name)
		if strings.Contains(filenameLower, queryLower) {
			results = append(results, SearchResult{
				Type:      "file",
				ItemTitle: fmt.Sprintf("📄 %s", f.Name),
				ItemDesc:  fmt.Sprintf("File match • Size: %d bytes", f.Size),
				FilePath:  f.Name,
				Content:   f.Name,
			})
		}
		return nil
	})

	return results, err
}

func searchContent(repo *git.Repository, query string) ([]SearchResult, error) {
	var results []SearchResult
	queryLower := strings.ToLower(query)

	// Create regex for better matching
	regex, err := regexp.Compile("(?i)" + regexp.QuoteMeta(query))
	if err != nil {
		regex = nil
	}

	ref, err := repo.Head()
	if err != nil {
		return results, err
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return results, err
	}

	tree, err := commit.Tree()
	if err != nil {
		return results, err
	}

	err = tree.Files().ForEach(func(f *object.File) error {
		// Skip binary files and large files
		if f.Size > 1024*1024 { // 1MB limit
			return nil
		}

		content, err := f.Contents()
		if err != nil {
			return nil
		}

		// Check if file is likely binary
		if strings.Contains(content, "\x00") {
			return nil
		}

		contentLower := strings.ToLower(content)
		if strings.Contains(contentLower, queryLower) {
			lines := strings.Split(content, "\n")
			for i, line := range lines {
				lineLower := strings.ToLower(line)
				if strings.Contains(lineLower, queryLower) {
					// Highlight the match
					highlightedLine := line
					if regex != nil {
						highlightedLine = regex.ReplaceAllStringFunc(line, func(match string) string {
							return matchStyle.Render(match)
						})
					}

					results = append(results, SearchResult{
						Type:       "content",
						ItemTitle:  fmt.Sprintf("🔍 %s:%d", f.Name, i+1),
						ItemDesc:   fmt.Sprintf("Content match • Line %d", i+1),
						FilePath:   f.Name,
						LineNumber: i + 1,
						Content:    strings.TrimSpace(highlightedLine),
					})

					// Limit results per file to avoid spam
					break
				}
			}
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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.searchInput.Width = msg.Width - 4
		m.resultsList.SetWidth(msg.Width)
		m.resultsList.SetHeight(msg.Height - 8)
		return m, nil

	case searchCompletedMsg:
		m.loading = false
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
		m.err = msg.err
		return m, nil

	case tea.KeyMsg:
		switch m.currentMode {
		case InputMode:
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "enter":
				if m.searchInput.Value() != "" {
					m.loading = true
					m.searchQuery = m.searchInput.Value()
					return m, func() tea.Msg {
						return performSearch(m.searchQuery, m.searchType)
					}
				}
			case "tab":
				// Cycle through search types
				m.searchType = (m.searchType + 1) % 4
				return m, nil
			default:
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				return m, cmd
			}

		case ResultsMode:
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "esc":
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
		return statusStyle.Render("🔍 Searching...")
	}

	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	switch m.currentMode {
	case InputMode:
		searchTypeText := []string{"All", "Commits", "Files", "Content", "Authors"}[m.searchType]

		return fmt.Sprintf(
			"%s\n\n%s\n%s\n\n%s",
			titleStyle.Render("🔍 Advanced Repository Search"),
			searchStyle.Render("Search: "+m.searchInput.View()),
			typeStyle.Render(fmt.Sprintf("Type: %s (tab to change)", searchTypeText)),
			helpStyle.Render("enter: search • tab: change type • q: quit"),
		)

	case DetailMode:
		if m.selectedResult == nil {
			return "No result selected"
		}
		return m.renderResultDetail(*m.selectedResult)

	default: // ResultsMode
		help := fmt.Sprintf("Found %d results for '%s' • enter: details • n: new search • esc: back • q: quit",
			len(m.results), m.searchQuery)

		return fmt.Sprintf(
			"%s\n%s",
			m.resultsList.View(),
			helpStyle.Render(help),
		)
	}
}

func (m model) renderResultDetail(result SearchResult) string {
	var details strings.Builder

	details.WriteString(titleStyle.Render(fmt.Sprintf("🔍 %s Details", strings.Title(result.Type))))
	details.WriteString("\n\n")

	var content string
	switch result.Type {
	case "commit":
		content = fmt.Sprintf(
			"Hash: %s\nAuthor: %s\nDate: %s\nMessage:\n%s",
			result.Hash,
			result.Author,
			result.Date.Format("2006-01-02 15:04:05"),
			result.Content,
		)
	case "file":
		content = fmt.Sprintf(
			"File: %s\nPath: %s\nMatched: %s",
			result.FilePath,
			result.FilePath,
			result.Content,
		)
	case "content":
		content = fmt.Sprintf(
			"File: %s\nLine: %d\nContent:\n%s",
			result.FilePath,
			result.LineNumber,
			result.Content,
		)
	case "author":
		content = fmt.Sprintf(
			"Author: %s\nMatched: %s",
			result.Author,
			result.Content,
		)
	}

	details.WriteString(detailStyle.Render(content))
	details.WriteString("\n\n")
	details.WriteString(helpStyle.Render("esc: back to results • q: quit"))

	return details.String()
}

func RunAdvancedSearch(args []string) error {
	p := tea.NewProgram(initialModel(args), tea.WithAltScreen())
	_, err := p.Run()
	if err != nil {
		fmt.Printf("Error running search: %v\n", err)
		os.Exit(1)
	}
	return nil
}
