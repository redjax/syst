package diffService

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/redjax/syst/internal/utils/terminal"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type ViewMode int

const (
	OverviewView ViewMode = iota
	FilesView
	DiffView
	StatsView
)

type DiffAnalysis struct {
	FromRef      string
	ToRef        string
	FromCommit   string
	ToCommit     string
	FilesChanged []FileDiff
	Stats        DiffStats
	Summary      string
}

type FileDiff struct {
	Path      string
	Status    string // "modified", "added", "deleted", "renamed", "copied"
	OldPath   string // For renames/copies
	Additions int
	Deletions int
	Changes   []DiffLine
	IsBinary  bool
}

type DiffLine struct {
	Type    string // "added", "deleted", "context", "header"
	OldLine int
	NewLine int
	Content string
}

type DiffStats struct {
	FilesChanged int
	Additions    int
	Deletions    int
	TotalChanges int
}

type model struct {
	// Current state
	currentView     ViewMode
	analysis        DiffAnalysis
	selectedFile    FileDiff
	selectedFileIdx int

	// UI components
	overviewList list.Model
	filesList    list.Model
	searchInput  textinput.Model

	// UI state
	loading    bool
	err        error
	tuiHelper *terminal.ResponsiveTUIHelper
	showSearch bool
}

// Messages
type diffAnalysisMsg struct {
	analysis DiffAnalysis
}

type errMsg struct {
	err error
}

// RunDiffExplorer starts the interactive diff explorer TUI
func RunDiffExplorer(args []string) error {
	// Parse arguments to determine what to compare
	fromRef := "HEAD^"
	toRef := "HEAD"

	if len(args) >= 1 {
		fromRef = args[0]
	}
	if len(args) >= 2 {
		toRef = args[1]
	}

	// Initialize model
	m := model{
		currentView: OverviewView,
		loading:     true,
		tuiHelper: terminal.NewResponsiveTUIHelper(),
	}

	// Initialize UI components
	m.overviewList = list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	m.overviewList.Title = "📊 Diff Overview"
	m.overviewList.SetShowHelp(false)

	m.filesList = list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	m.filesList.Title = "📁 Changed Files"
	m.filesList.SetShowHelp(false)

	m.searchInput = textinput.New()
	m.searchInput.Placeholder = "Search files..."
	m.searchInput.CharLimit = 100

	// Start the TUI
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Load diff analysis
	go func() {
		p.Send(loadDiffAnalysis(fromRef, toRef))
	}()

	_, err := p.Run()
	return err
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.tuiHelper.HandleWindowSizeMsg(msg)

		listHeight := m.tuiHelper.GetHeight() - 10
		listWidth := m.tuiHelper.GetWidth() - 4

		m.overviewList.SetSize(listWidth, listHeight)
		m.filesList.SetSize(listWidth, listHeight)

	case diffAnalysisMsg:
		m.loading = false
		m.analysis = msg.analysis

		// Update overview list
		overviewItems := []list.Item{
			OverviewItem{title: "📊 Summary", desc: m.analysis.Summary},
			OverviewItem{title: "📁 Files Changed", desc: fmt.Sprintf("%d files", m.analysis.Stats.FilesChanged)},
			OverviewItem{title: "➕ Additions", desc: fmt.Sprintf("+%d lines", m.analysis.Stats.Additions)},
			OverviewItem{title: "➖ Deletions", desc: fmt.Sprintf("-%d lines", m.analysis.Stats.Deletions)},
			OverviewItem{title: "🔄 Total Changes", desc: fmt.Sprintf("%d lines", m.analysis.Stats.TotalChanges)},
		}
		m.overviewList.SetItems(overviewItems)

		// Update files list
		fileItems := make([]list.Item, len(m.analysis.FilesChanged))
		for i, file := range m.analysis.FilesChanged {
			fileItems[i] = FileDiffItem{diff: file}
		}
		m.filesList.SetItems(fileItems)

	case errMsg:
		m.loading = false
		m.err = msg.err

	case tea.KeyMsg:
		// Handle global keys first
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))):
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			if m.showSearch {
				m.showSearch = false
				m.searchInput.Blur()
				return m, nil
			}
			if m.currentView != OverviewView {
				m.currentView = OverviewView
				return m, nil
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("/"))):
			if m.currentView == FilesView {
				m.showSearch = !m.showSearch
				if m.showSearch {
					m.searchInput.Focus()
				} else {
					m.searchInput.Blur()
				}
				return m, nil
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("1"))):
			m.currentView = OverviewView
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("2"))):
			m.currentView = FilesView
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("3"))):
			if len(m.analysis.FilesChanged) > 0 {
				m.currentView = DiffView
			}
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("4"))):
			m.currentView = StatsView
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
			m.loading = true
			return m, func() tea.Msg {
				return loadDiffAnalysis(m.analysis.FromRef, m.analysis.ToRef)
			}
		}

		// Handle view-specific keys
		if m.showSearch {
			switch msg.Type {
			case tea.KeyEnter:
				// Perform search
				query := m.searchInput.Value()
				if query != "" {
					// Filter file list
					filteredFiles := make([]FileDiff, 0)
					for _, file := range m.analysis.FilesChanged {
						if strings.Contains(strings.ToLower(file.Path), strings.ToLower(query)) {
							filteredFiles = append(filteredFiles, file)
						}
					}
					items := make([]list.Item, len(filteredFiles))
					for i, file := range filteredFiles {
						items[i] = FileDiffItem{diff: file}
					}
					m.filesList.SetItems(items)
				}
				m.showSearch = false
				m.searchInput.Blur()
				return m, nil
			}
			m.searchInput, cmd = m.searchInput.Update(msg)
			return m, cmd
		}

		switch m.currentView {
		case OverviewView:
			m.overviewList, cmd = m.overviewList.Update(msg)

		case FilesView:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				if item, ok := m.filesList.SelectedItem().(FileDiffItem); ok {
					m.selectedFile = item.diff
					m.selectedFileIdx = m.filesList.Index()
					m.currentView = DiffView
					return m, nil
				}
			}
			m.filesList, cmd = m.filesList.Update(msg)

		case DiffView:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("left", "h"))):
				if m.selectedFileIdx > 0 {
					m.selectedFileIdx--
					m.selectedFile = m.analysis.FilesChanged[m.selectedFileIdx]
				}
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("right", "l"))):
				if m.selectedFileIdx < len(m.analysis.FilesChanged)-1 {
					m.selectedFileIdx++
					m.selectedFile = m.analysis.FilesChanged[m.selectedFileIdx]
				}
				return m, nil
			}

		case StatsView:
			// No specific handling needed for stats view
		}
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.loading {
		return m.renderLoading()
	}

	if m.err != nil {
		return m.renderError()
	}

	switch m.currentView {
	case OverviewView:
		return m.renderOverview()
	case FilesView:
		return m.renderFilesView()
	case DiffView:
		return m.renderDiffView()
	case StatsView:
		return m.renderStatsView()
	default:
		return m.renderOverview()
	}
}

func loadDiffAnalysis(fromRef, toRef string) tea.Msg {
	analysis, err := analyzeDiff(fromRef, toRef)
	if err != nil {
		return errMsg{err}
	}
	return diffAnalysisMsg{analysis}
}

func analyzeDiff(fromRef, toRef string) (DiffAnalysis, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return DiffAnalysis{}, err
	}

	// Resolve references to commits
	fromCommit, err := resolveRef(repo, fromRef)
	if err != nil {
		return DiffAnalysis{}, fmt.Errorf("failed to resolve '%s': %w", fromRef, err)
	}

	toCommit, err := resolveRef(repo, toRef)
	if err != nil {
		return DiffAnalysis{}, fmt.Errorf("failed to resolve '%s': %w", toRef, err)
	}

	// Get commit objects
	fromCommitObj, err := repo.CommitObject(fromCommit)
	if err != nil {
		return DiffAnalysis{}, err
	}

	toCommitObj, err := repo.CommitObject(toCommit)
	if err != nil {
		return DiffAnalysis{}, err
	}

	// Get trees
	fromTree, err := fromCommitObj.Tree()
	if err != nil {
		return DiffAnalysis{}, err
	}

	toTree, err := toCommitObj.Tree()
	if err != nil {
		return DiffAnalysis{}, err
	}

	// Calculate diff
	changes, err := fromTree.Diff(toTree)
	if err != nil {
		return DiffAnalysis{}, err
	}

	// Process changes
	var filesChanged []FileDiff
	totalAdditions := 0
	totalDeletions := 0

	for _, change := range changes {
		fileDiff := processFileDiff(change)
		filesChanged = append(filesChanged, fileDiff)
		totalAdditions += fileDiff.Additions
		totalDeletions += fileDiff.Deletions
	}

	// Sort files by path
	sort.Slice(filesChanged, func(i, j int) bool {
		return filesChanged[i].Path < filesChanged[j].Path
	})

	stats := DiffStats{
		FilesChanged: len(filesChanged),
		Additions:    totalAdditions,
		Deletions:    totalDeletions,
		TotalChanges: totalAdditions + totalDeletions,
	}

	summary := fmt.Sprintf("Comparing %s → %s", fromRef, toRef)

	return DiffAnalysis{
		FromRef:      fromRef,
		ToRef:        toRef,
		FromCommit:   fromCommit.String(),
		ToCommit:     toCommit.String(),
		FilesChanged: filesChanged,
		Stats:        stats,
		Summary:      summary,
	}, nil
}

func resolveRef(repo *git.Repository, ref string) (plumbing.Hash, error) {
	// Try to resolve as a hash first
	if hash := plumbing.NewHash(ref); hash.IsZero() == false {
		// Check if it's a valid commit
		if _, err := repo.CommitObject(hash); err == nil {
			return hash, nil
		}
	}

	// Try to resolve as a reference
	resolved, err := repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return plumbing.ZeroHash, err
	}

	return *resolved, nil
}

func processFileDiff(change *object.Change) FileDiff {
	// Determine status and paths
	var status, path, oldPath string
	var additions, deletions int

	switch {
	case change.From.Name == "" && change.To.Name != "":
		// Added file
		status = "added"
		path = change.To.Name
	case change.From.Name != "" && change.To.Name == "":
		// Deleted file
		status = "deleted"
		path = change.From.Name
	case change.From.Name != change.To.Name:
		// Renamed/moved file
		status = "renamed"
		path = change.To.Name
		oldPath = change.From.Name
	default:
		// Modified file
		status = "modified"
		path = change.To.Name
	}

	// Get patch for line counts (simplified)
	patch, err := change.Patch()
	if err == nil {
		stats := patch.Stats()
		if len(stats) > 0 {
			additions = stats[0].Addition
			deletions = stats[0].Deletion
		}
	}

	// Check if binary file
	isBinary := false
	if patch != nil {
		isBinary = strings.Contains(patch.String(), "Binary files")
	}

	// Generate diff lines for display (simplified)
	var diffLines []DiffLine
	if !isBinary && patch != nil {
		diffLines = generateDiffLines(patch.String())
	}

	return FileDiff{
		Path:      path,
		Status:    status,
		OldPath:   oldPath,
		Additions: additions,
		Deletions: deletions,
		Changes:   diffLines,
		IsBinary:  isBinary,
	}
}

func generateDiffLines(patchStr string) []DiffLine {
	var lines []DiffLine
	patchLines := strings.Split(patchStr, "\n")

	oldLine := 0
	newLine := 0

	for _, line := range patchLines {
		if len(line) == 0 {
			continue
		}

		var diffLine DiffLine

		switch line[0] {
		case '+':
			if strings.HasPrefix(line, "+++") {
				// File header
				diffLine = DiffLine{
					Type:    "header",
					Content: line,
				}
			} else {
				// Added line
				newLine++
				diffLine = DiffLine{
					Type:    "added",
					NewLine: newLine,
					Content: line,
				}
			}
		case '-':
			if strings.HasPrefix(line, "---") {
				// File header
				diffLine = DiffLine{
					Type:    "header",
					Content: line,
				}
			} else {
				// Deleted line
				oldLine++
				diffLine = DiffLine{
					Type:    "deleted",
					OldLine: oldLine,
					Content: line,
				}
			}
		case '@':
			// Hunk header
			diffLine = DiffLine{
				Type:    "header",
				Content: line,
			}
		default:
			// Context line
			oldLine++
			newLine++
			diffLine = DiffLine{
				Type:    "context",
				OldLine: oldLine,
				NewLine: newLine,
				Content: " " + line, // Add space prefix for context
			}
		}

		lines = append(lines, diffLine)

		// Limit lines to avoid overwhelming the UI
		if len(lines) > 200 {
			lines = append(lines, DiffLine{
				Type:    "header",
				Content: "... (truncated, showing first 200 lines)",
			})
			break
		}
	}

	return lines
}

// List item types
type OverviewItem struct {
	title string
	desc  string
}

func (o OverviewItem) Title() string       { return o.title }
func (o OverviewItem) Description() string { return o.desc }
func (o OverviewItem) FilterValue() string { return o.title + " " + o.desc }

type FileDiffItem struct {
	diff FileDiff
}

func (f FileDiffItem) Title() string {
	statusIcon := "📝"
	switch f.diff.Status {
	case "added":
		statusIcon = "✅"
	case "deleted":
		statusIcon = "❌"
	case "renamed":
		statusIcon = "🔄"
	case "copied":
		statusIcon = "📋"
	}

	title := fmt.Sprintf("%s %s", statusIcon, f.diff.Path)
	if f.diff.Status == "renamed" && f.diff.OldPath != "" {
		title = fmt.Sprintf("%s %s ← %s", statusIcon, f.diff.Path, f.diff.OldPath)
	}
	return title
}

func (f FileDiffItem) Description() string {
	if f.diff.IsBinary {
		return "Binary file"
	}
	return fmt.Sprintf("+%d -%d lines", f.diff.Additions, f.diff.Deletions)
}

func (f FileDiffItem) FilterValue() string {
	return f.diff.Path + " " + f.diff.OldPath
}

// Render functions
func (m model) renderLoading() string {
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginTop(2).
		MarginLeft(2)

	return style.Render("🔍 Analyzing diff...")
}

func (m model) renderError() string {
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("196")).
		MarginTop(2).
		MarginLeft(2)

	return style.Render(fmt.Sprintf("❌ Error: %v", m.err))
}

func (m model) renderOverview() string {
	var content strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	title := fmt.Sprintf("📊 Diff Overview: %s → %s", m.analysis.FromRef, m.analysis.ToRef)
	content.WriteString(headerStyle.Render(title))
	content.WriteString("\n")

	// Overview list
	if m.showSearch {
		searchStyle := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("39")).
			Padding(0, 1).
			MarginBottom(1)

		content.WriteString(searchStyle.Render("🔍 " + m.searchInput.View()))
		content.WriteString("\n")
	}

	content.WriteString(m.overviewList.View())
	content.WriteString("\n")

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	help := "1: overview • 2: files • 3: diff • 4: stats • /: search • r: refresh • q: quit"
	content.WriteString(helpStyle.Render(help))

	return content.String()
}

func (m model) renderFilesView() string {
	var content strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	title := fmt.Sprintf("📁 Changed Files (%d files)", len(m.analysis.FilesChanged))
	content.WriteString(headerStyle.Render(title))
	content.WriteString("\n")

	// Search input
	if m.showSearch {
		searchStyle := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("39")).
			Padding(0, 1).
			MarginBottom(1)

		content.WriteString(searchStyle.Render("🔍 " + m.searchInput.View()))
		content.WriteString("\n")
	}

	// Files list
	content.WriteString(m.filesList.View())
	content.WriteString("\n")

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	help := "1: overview • 2: files • 3: diff • enter: view diff • /: search • r: refresh • q: quit"
	content.WriteString(helpStyle.Render(help))

	return content.String()
}

func (m model) renderDiffView() string {
	var content strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	statusIcon := "📝"
	switch m.selectedFile.Status {
	case "added":
		statusIcon = "✅"
	case "deleted":
		statusIcon = "❌"
	case "renamed":
		statusIcon = "🔄"
	}

	title := fmt.Sprintf("%s %s", statusIcon, m.selectedFile.Path)
	if m.selectedFile.Status == "renamed" && m.selectedFile.OldPath != "" {
		title = fmt.Sprintf("%s %s ← %s", statusIcon, m.selectedFile.Path, m.selectedFile.OldPath)
	}

	content.WriteString(headerStyle.Render(title))
	content.WriteString("\n")

	// File stats
	statsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("242")).
		MarginBottom(1)

	fileNavigation := fmt.Sprintf("File %d of %d", m.selectedFileIdx+1, len(m.analysis.FilesChanged))
	stats := fmt.Sprintf("%s • +%d -%d lines", fileNavigation, m.selectedFile.Additions, m.selectedFile.Deletions)
	if m.selectedFile.IsBinary {
		stats += " • Binary file"
	}

	content.WriteString(statsStyle.Render(stats))
	content.WriteString("\n")

	// Diff content
	if m.selectedFile.IsBinary {
		binaryStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true).
			MarginBottom(1)
		content.WriteString(binaryStyle.Render("📄 Binary file - no diff preview available"))
		content.WriteString("\n")
	} else if len(m.selectedFile.Changes) > 0 {
		diffStyle := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("238")).
			Padding(1, 2).
			MaxHeight(m.tuiHelper.GetHeight() - 10)

		var diff strings.Builder

		for i, line := range m.selectedFile.Changes {
			if i > 50 { // Limit display to avoid overwhelming
				diff.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("... (showing first 50 lines, use git for full diff)\n"))
				break
			}

			var lineStyle lipgloss.Style
			switch line.Type {
			case "added":
				lineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("34")) // green
			case "deleted":
				lineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("31")) // red
			case "context":
				lineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245")) // gray
			case "header":
				lineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true) // yellow
			default:
				lineStyle = lipgloss.NewStyle()
			}

			diff.WriteString(lineStyle.Render(line.Content))
			diff.WriteString("\n")
		}

		content.WriteString(diffStyle.Render(diff.String()))
	} else {
		// No changes to show
		noChangesStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true).
			MarginBottom(1)

		var message string
		switch m.selectedFile.Status {
		case "added":
			message = "New file created"
		case "deleted":
			message = "File was deleted"
		default:
			message = "No detailed changes available"
		}

		content.WriteString(noChangesStyle.Render(message))
		content.WriteString("\n")
	}

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	help := "1: overview • 2: files • ←/→: prev/next file • esc: back • q: quit"
	content.WriteString(helpStyle.Render(help))

	return content.String()
}

func (m model) renderStatsView() string {
	var content strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	title := "📊 Diff Statistics"
	content.WriteString(headerStyle.Render(title))
	content.WriteString("\n")

	// Statistics
	statsStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("238")).
		Padding(1, 2).
		MarginBottom(1)

	var stats strings.Builder
	stats.WriteString(fmt.Sprintf("📁 Files Changed: %d\n", m.analysis.Stats.FilesChanged))
	stats.WriteString(fmt.Sprintf("➕ Lines Added: %d\n", m.analysis.Stats.Additions))
	stats.WriteString(fmt.Sprintf("➖ Lines Deleted: %d\n", m.analysis.Stats.Deletions))
	stats.WriteString(fmt.Sprintf("🔄 Total Changes: %d\n", m.analysis.Stats.TotalChanges))
	stats.WriteString("\n")
	stats.WriteString(fmt.Sprintf("📝 From: %s (%s)\n", m.analysis.FromRef, m.analysis.FromCommit[:8]))
	stats.WriteString(fmt.Sprintf("📝 To: %s (%s)\n", m.analysis.ToRef, m.analysis.ToCommit[:8]))

	content.WriteString(statsStyle.Render(stats.String()))

	// File type breakdown
	if len(m.analysis.FilesChanged) > 0 {
		typeStats := make(map[string]int)
		for _, file := range m.analysis.FilesChanged {
			ext := filepath.Ext(file.Path)
			if ext == "" {
				ext = "no extension"
			}
			typeStats[ext]++
		}

		breakdownStyle := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("238")).
			Padding(1, 2).
			MarginBottom(1)

		var breakdown strings.Builder
		breakdown.WriteString("📊 File Types:\n\n")

		// Sort by count
		type fileTypeStat struct {
			ext   string
			count int
		}
		var sortedStats []fileTypeStat
		for ext, count := range typeStats {
			sortedStats = append(sortedStats, fileTypeStat{ext, count})
		}
		sort.Slice(sortedStats, func(i, j int) bool {
			return sortedStats[i].count > sortedStats[j].count
		})

		for _, stat := range sortedStats {
			breakdown.WriteString(fmt.Sprintf("  %s: %d files\n", stat.ext, stat.count))
		}

		content.WriteString(breakdownStyle.Render(breakdown.String()))
	}

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	help := "1: overview • 2: files • 3: diff • 4: stats • r: refresh • q: quit"
	content.WriteString(helpStyle.Render(help))

	return content.String()
}
