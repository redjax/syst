package filesService

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/redjax/syst/internal/utils/terminal"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type ViewMode int

const (
	OverviewView ViewMode = iota
	LargeFilesView
	FrequentFilesView
	ExtensionsView
	ContributorsView
)

type FileAnalysis struct {
	Overview           FileOverview
	LargeFiles         []LargeFileInfo
	FrequentFiles      []FrequentFileInfo
	ExtensionBreakdown []ExtensionInfo
	FileContributors   []FileContributorInfo
}

type FileOverview struct {
	TotalFiles      int
	TotalSize       int64
	AverageSize     int64
	LargestFile     string
	LargestFileSize int64
	ExtensionCount  int
	BinaryFiles     int
	TextFiles       int
}

type LargeFileInfo struct {
	Path      string
	Size      int64
	Extension string
	Type      string // "binary" or "text"
}

type FrequentFileInfo struct {
	Path           string
	ChangeCount    int
	Contributors   int
	LastModified   time.Time
	LastCommitHash string
	LastCommitMsg  string
	TotalAdditions int
	TotalDeletions int
}

type ExtensionInfo struct {
	Extension   string
	FileCount   int
	TotalSize   int64
	AverageSize int64
	Language    string
}

type FileContributorInfo struct {
	Path         string
	Contributors []ContributorStat
	TotalChanges int
	Ownership    string // Most active contributor
}

type ContributorStat struct {
	Name       string
	Changes    int
	Percentage float64
}

type model struct {
	analysis    FileAnalysis
	currentView ViewMode
	fileList    list.Model
	loading     bool
	err         error
	tuiHelper *terminal.ResponsiveTUIHelper
	sections    []string
}

type fileItem struct {
	file interface{}
}

func (i fileItem) FilterValue() string {
	switch f := i.file.(type) {
	case LargeFileInfo:
		return f.Path
	case FrequentFileInfo:
		return f.Path
	case ExtensionInfo:
		return f.Extension
	case FileContributorInfo:
		return f.Path
	default:
		return ""
	}
}

func (i fileItem) Title() string {
	switch f := i.file.(type) {
	case LargeFileInfo:
		return fmt.Sprintf("%s (%s)", f.Path, formatBytes(f.Size))
	case FrequentFileInfo:
		return fmt.Sprintf("%s (%d changes)", f.Path, f.ChangeCount)
	case ExtensionInfo:
		return fmt.Sprintf("%s (%d files)", f.Extension, f.FileCount)
	case FileContributorInfo:
		return fmt.Sprintf("%s (%d contributors)", f.Path, len(f.Contributors))
	default:
		return "Unknown"
	}
}

func (i fileItem) Description() string {
	switch f := i.file.(type) {
	case LargeFileInfo:
		return fmt.Sprintf("Type: %s • Extension: %s", f.Type, f.Extension)
	case FrequentFileInfo:
		return fmt.Sprintf("Contributors: %d • Last: %s", f.Contributors, f.LastModified.Format("2006-01-02"))
	case ExtensionInfo:
		return fmt.Sprintf("Language: %s • Total: %s", f.Language, formatBytes(f.TotalSize))
	case FileContributorInfo:
		return fmt.Sprintf("Main contributor: %s • %d total changes", f.Ownership, f.TotalChanges)
	default:
		return ""
	}
}

type dataLoadedMsg struct {
	analysis FileAnalysis
}

type errMsg struct {
	err error
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#F25D94")).
			Padding(0, 1)

	sectionStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 2).
			MarginTop(1)

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginBottom(1)

	statsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	highlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#01FAC6")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)
)

func (m model) Init() tea.Cmd {
	return loadFileAnalysis
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.tuiHelper.HandleWindowSizeMsg(msg)
		m.fileList.SetWidth(m.tuiHelper.GetWidth())
		m.fileList.SetHeight(m.tuiHelper.GetHeight() - 12)
		return m, nil

	case dataLoadedMsg:
		m.analysis = msg.analysis
		m.loading = false
		m.sections = []string{
			"Overview",
			"Large Files",
			"Frequent Changes",
			"Extensions",
			"Contributors",
		}
		m.updateListItems()
		return m, nil

	case errMsg:
		m.err = msg.err
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c", "esc"))):
			return m, tea.Quit
		case key.Matches(msg, key.NewBinding(key.WithKeys("1"))):
			m.currentView = OverviewView
			m.updateListItems()
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("2"))):
			m.currentView = LargeFilesView
			m.updateListItems()
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("3"))):
			m.currentView = FrequentFilesView
			m.updateListItems()
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("4"))):
			m.currentView = ExtensionsView
			m.updateListItems()
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("5"))):
			m.currentView = ContributorsView
			m.updateListItems()
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("left", "h"))):
			if m.currentView > 0 {
				m.currentView--
				m.updateListItems()
			}
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("right", "l"))):
			if int(m.currentView) < len(m.sections)-1 {
				m.currentView++
				m.updateListItems()
			}
			return m, nil
		default:
			var cmd tea.Cmd
			m.fileList, cmd = m.fileList.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m *model) updateListItems() {
	var items []list.Item

	switch m.currentView {
	case LargeFilesView:
		for _, file := range m.analysis.LargeFiles {
			items = append(items, fileItem{file: file})
		}
	case FrequentFilesView:
		for _, file := range m.analysis.FrequentFiles {
			items = append(items, fileItem{file: file})
		}
	case ExtensionsView:
		for _, ext := range m.analysis.ExtensionBreakdown {
			items = append(items, fileItem{file: ext})
		}
	case ContributorsView:
		for _, file := range m.analysis.FileContributors {
			items = append(items, fileItem{file: file})
		}
	}

	m.fileList.SetItems(items)
}

func (m model) View() string {
	if m.loading {
		return "\n  Analyzing repository files...\n"
	}

	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("\n  Error: %v\n", m.err))
	}

	var sections []string

	// Title
	title := titleStyle.Render("📁 File Analysis")
	sections = append(sections, title)

	// Navigation tabs
	tabs := m.renderTabs()
	sections = append(sections, tabs)

	// Content based on current view
	content := m.renderCurrentView()
	sections = append(sections, sectionStyle.Render(content))

	// Instructions
	help := helpStyle.Render("1-5: sections • ←/→: navigate • ↑/↓: scroll • q: quit")
	sections = append(sections, help)

	return strings.Join(sections, "\n")
}

func (m model) renderTabs() string {
	var tabs []string

	for i, section := range m.sections {
		style := lipgloss.NewStyle().Padding(0, 1)
		if ViewMode(i) == m.currentView {
			style = style.Foreground(lipgloss.Color("#01FAC6")).Bold(true).
				Background(lipgloss.Color("#874BFD"))
		}
		tabs = append(tabs, style.Render(fmt.Sprintf("%d. %s", i+1, section)))
	}

	return strings.Join(tabs, " ")
}

func (m model) renderCurrentView() string {
	switch m.currentView {
	case OverviewView:
		return m.renderOverview()
	case LargeFilesView:
		return m.renderWithList("📦 Large Files", "Files larger than 100KB")
	case FrequentFilesView:
		return m.renderWithList("🔄 Frequently Changed Files", "Files with the most commits")
	case ExtensionsView:
		return m.renderWithList("🗂️ File Extensions", "File types and their distribution")
	case ContributorsView:
		return m.renderWithList("👥 File Contributors", "Files with multiple contributors")
	default:
		return "Unknown view"
	}
}

func (m model) renderOverview() string {
	overview := m.analysis.Overview
	var content strings.Builder

	content.WriteString(headerStyle.Render("📊 Repository File Overview"))
	content.WriteString("\n\n")

	content.WriteString(fmt.Sprintf("Total Files: %s\n",
		statsStyle.Render(fmt.Sprintf("%d", overview.TotalFiles))))
	content.WriteString(fmt.Sprintf("Total Size: %s\n",
		statsStyle.Render(formatBytes(overview.TotalSize))))
	content.WriteString(fmt.Sprintf("Average File Size: %s\n",
		statsStyle.Render(formatBytes(overview.AverageSize))))
	content.WriteString(fmt.Sprintf("Largest File: %s (%s)\n",
		highlightStyle.Render(overview.LargestFile),
		statsStyle.Render(formatBytes(overview.LargestFileSize))))
	content.WriteString(fmt.Sprintf("File Types: %s extensions\n",
		statsStyle.Render(fmt.Sprintf("%d", overview.ExtensionCount))))
	content.WriteString(fmt.Sprintf("Binary Files: %s\n",
		statsStyle.Render(fmt.Sprintf("%d", overview.BinaryFiles))))
	content.WriteString(fmt.Sprintf("Text Files: %s\n",
		statsStyle.Render(fmt.Sprintf("%d", overview.TextFiles))))

	// Quick stats from other views
	content.WriteString("\n")
	content.WriteString(headerStyle.Render("📈 Quick Statistics"))
	content.WriteString("\n\n")

	if len(m.analysis.LargeFiles) > 0 {
		content.WriteString(fmt.Sprintf("Large Files (>100KB): %s\n",
			statsStyle.Render(fmt.Sprintf("%d", len(m.analysis.LargeFiles)))))
	}

	if len(m.analysis.FrequentFiles) > 0 {
		mostChanged := m.analysis.FrequentFiles[0]
		content.WriteString(fmt.Sprintf("Most Changed File: %s (%d changes)\n",
			highlightStyle.Render(mostChanged.Path), mostChanged.ChangeCount))
	}

	if len(m.analysis.ExtensionBreakdown) > 0 {
		mostCommon := m.analysis.ExtensionBreakdown[0]
		content.WriteString(fmt.Sprintf("Most Common Extension: %s (%d files)\n",
			highlightStyle.Render(mostCommon.Extension), mostCommon.FileCount))
	}

	return content.String()
}

func (m model) renderWithList(title, subtitle string) string {
	var content strings.Builder

	content.WriteString(headerStyle.Render(title))
	content.WriteString("\n")
	content.WriteString(subtitle)
	content.WriteString("\n\n")

	if len(m.fileList.Items()) == 0 {
		content.WriteString("No items to display")
		return content.String()
	}

	content.WriteString(m.fileList.View())
	return content.String()
}

func loadFileAnalysis() tea.Msg {
	analysis, err := analyzeFiles()
	if err != nil {
		return errMsg{err}
	}
	return dataLoadedMsg{analysis}
}

func analyzeFiles() (FileAnalysis, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return FileAnalysis{}, fmt.Errorf("failed to open repository: %w", err)
	}

	ref, err := repo.Head()
	if err != nil {
		return FileAnalysis{}, fmt.Errorf("failed to get HEAD: %w", err)
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return FileAnalysis{}, fmt.Errorf("failed to get commit: %w", err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return FileAnalysis{}, fmt.Errorf("failed to get tree: %w", err)
	}

	analysis := FileAnalysis{}

	// Analyze current files in git tree
	err = analyzeCurrentFiles(tree, &analysis)
	if err != nil {
		return FileAnalysis{}, fmt.Errorf("failed to analyze current files: %w", err)
	}

	// Analyze file history
	err = analyzeFileHistory(repo, &analysis)
	if err != nil {
		return FileAnalysis{}, fmt.Errorf("failed to analyze file history: %w", err)
	}

	// Process and sort results
	processAnalysisResults(&analysis)

	return analysis, nil
}

func analyzeCurrentFiles(tree *object.Tree, analysis *FileAnalysis) error {
	var totalSize int64
	var fileCount int
	var largestFile string
	var largestSize int64
	var binaryCount int

	extensionStats := make(map[string]*ExtensionInfo)
	var largeFiles []LargeFileInfo

	err := tree.Files().ForEach(func(file *object.File) error {
		fileCount++
		totalSize += file.Size

		// Track largest file
		if file.Size > largestSize {
			largestSize = file.Size
			largestFile = file.Name
		}

		// File extension analysis
		ext := strings.ToLower(filepath.Ext(file.Name))
		if ext == "" {
			ext = "no extension"
		}

		if extensionStats[ext] == nil {
			extensionStats[ext] = &ExtensionInfo{
				Extension: ext,
				Language:  getLanguageForExtension(ext),
			}
		}
		extensionStats[ext].FileCount++
		extensionStats[ext].TotalSize += file.Size

		// Check if binary
		isBinary := isBinaryFile(file.Name)
		if isBinary {
			binaryCount++
		}

		// Large files (>100KB)
		if file.Size > 100*1024 {
			fileType := "text"
			if isBinary {
				fileType = "binary"
			}
			largeFiles = append(largeFiles, LargeFileInfo{
				Path:      file.Name,
				Size:      file.Size,
				Extension: ext,
				Type:      fileType,
			})
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Calculate averages for extensions
	var extensions []ExtensionInfo
	for _, ext := range extensionStats {
		if ext.FileCount > 0 {
			ext.AverageSize = ext.TotalSize / int64(ext.FileCount)
		}
		extensions = append(extensions, *ext)
	}

	// Sort large files by size
	sort.Slice(largeFiles, func(i, j int) bool {
		return largeFiles[i].Size > largeFiles[j].Size
	})

	// Sort extensions by file count
	sort.Slice(extensions, func(i, j int) bool {
		return extensions[i].FileCount > extensions[j].FileCount
	})

	analysis.Overview = FileOverview{
		TotalFiles: fileCount,
		TotalSize:  totalSize,
		AverageSize: func() int64 {
			if fileCount > 0 {
				return totalSize / int64(fileCount)
			} else {
				return 0
			}
		}(),
		LargestFile:     largestFile,
		LargestFileSize: largestSize,
		ExtensionCount:  len(extensionStats),
		BinaryFiles:     binaryCount,
		TextFiles:       fileCount - binaryCount,
	}

	analysis.LargeFiles = largeFiles
	analysis.ExtensionBreakdown = extensions

	return nil
}

func analyzeFileHistory(repo *git.Repository, analysis *FileAnalysis) error {
	ref, err := repo.Head()
	if err != nil {
		return err
	}

	cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return err
	}

	fileChangeCount := make(map[string]*FrequentFileInfo)
	fileContributors := make(map[string]map[string]int) // file -> contributor -> count

	err = cIter.ForEach(func(c *object.Commit) error {
		stats, err := c.Stats()
		if err != nil {
			return nil // Skip commits we can't analyze
		}

		for _, stat := range stats {
			fileName := stat.Name

			// Initialize file info if needed
			if fileChangeCount[fileName] == nil {
				fileChangeCount[fileName] = &FrequentFileInfo{
					Path:           fileName,
					LastModified:   c.Author.When,
					LastCommitHash: c.Hash.String()[:8],
					LastCommitMsg:  strings.Split(c.Message, "\n")[0],
				}
			}

			// Update file stats
			fileInfo := fileChangeCount[fileName]
			fileInfo.ChangeCount++
			fileInfo.TotalAdditions += stat.Addition
			fileInfo.TotalDeletions += stat.Deletion

			// Update last modified if this commit is newer
			if c.Author.When.After(fileInfo.LastModified) {
				fileInfo.LastModified = c.Author.When
				fileInfo.LastCommitHash = c.Hash.String()[:8]
				fileInfo.LastCommitMsg = strings.Split(c.Message, "\n")[0]
			}

			// Track contributors
			if fileContributors[fileName] == nil {
				fileContributors[fileName] = make(map[string]int)
			}
			fileContributors[fileName][c.Author.Name]++
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Convert to slices and calculate contributors
	var frequentFiles []FrequentFileInfo
	var fileContribData []FileContributorInfo

	for fileName, fileInfo := range fileChangeCount {
		fileInfo.Contributors = len(fileContributors[fileName])
		frequentFiles = append(frequentFiles, *fileInfo)

		// Build contributor info
		var contributors []ContributorStat
		totalChanges := 0
		maxContributor := ""
		maxChanges := 0

		for contributor, changes := range fileContributors[fileName] {
			totalChanges += changes
			if changes > maxChanges {
				maxChanges = changes
				maxContributor = contributor
			}
			contributors = append(contributors, ContributorStat{
				Name:    contributor,
				Changes: changes,
			})
		}

		// Calculate percentages
		for i := range contributors {
			contributors[i].Percentage = float64(contributors[i].Changes) / float64(totalChanges) * 100
		}

		// Sort contributors by changes
		sort.Slice(contributors, func(i, j int) bool {
			return contributors[i].Changes > contributors[j].Changes
		})

		fileContribData = append(fileContribData, FileContributorInfo{
			Path:         fileName,
			Contributors: contributors,
			TotalChanges: totalChanges,
			Ownership:    maxContributor,
		})
	}

	// Sort frequent files by change count
	sort.Slice(frequentFiles, func(i, j int) bool {
		return frequentFiles[i].ChangeCount > frequentFiles[j].ChangeCount
	})

	// Sort file contributors by total changes
	sort.Slice(fileContribData, func(i, j int) bool {
		return fileContribData[i].TotalChanges > fileContribData[j].TotalChanges
	})

	analysis.FrequentFiles = frequentFiles
	analysis.FileContributors = fileContribData

	return nil
}

func processAnalysisResults(analysis *FileAnalysis) {
	// Limit results to prevent overwhelming display
	if len(analysis.LargeFiles) > 50 {
		analysis.LargeFiles = analysis.LargeFiles[:50]
	}
	if len(analysis.FrequentFiles) > 50 {
		analysis.FrequentFiles = analysis.FrequentFiles[:50]
	}
	if len(analysis.ExtensionBreakdown) > 20 {
		analysis.ExtensionBreakdown = analysis.ExtensionBreakdown[:20]
	}
	if len(analysis.FileContributors) > 50 {
		analysis.FileContributors = analysis.FileContributors[:50]
	}
}

func getLanguageForExtension(ext string) string {
	languages := map[string]string{
		".go":    "Go",
		".js":    "JavaScript",
		".ts":    "TypeScript",
		".py":    "Python",
		".java":  "Java",
		".c":     "C",
		".cpp":   "C++",
		".cs":    "C#",
		".php":   "PHP",
		".rb":    "Ruby",
		".rs":    "Rust",
		".swift": "Swift",
		".kt":    "Kotlin",
		".scala": "Scala",
		".html":  "HTML",
		".css":   "CSS",
		".scss":  "SCSS",
		".less":  "LESS",
		".json":  "JSON",
		".xml":   "XML",
		".yaml":  "YAML",
		".yml":   "YAML",
		".md":    "Markdown",
		".txt":   "Text",
		".sh":    "Shell",
		".ps1":   "PowerShell",
		".bat":   "Batch",
		".sql":   "SQL",
		".r":     "R",
		".m":     "Objective-C",
		".dart":  "Dart",
		".lua":   "Lua",
		".perl":  "Perl",
		".pl":    "Perl",
	}

	if lang, exists := languages[ext]; exists {
		return lang
	}
	return "Unknown"
}

func isBinaryFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	binaryExts := []string{
		".exe", ".dll", ".so", ".dylib", ".bin", ".jpg", ".jpeg", ".png", ".gif",
		".pdf", ".zip", ".tar", ".gz", ".rar", ".7z", ".mp3", ".mp4", ".avi",
		".mov", ".wmv", ".flv", ".webm", ".ogg", ".wav", ".ico", ".ttf", ".woff",
		".woff2", ".eot", ".otf", ".class", ".jar", ".war", ".ear",
	}

	for _, binaryExt := range binaryExts {
		if ext == binaryExt {
			return true
		}
	}
	return false
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// RunFileAnalysis starts the file analysis TUI
func RunFileAnalysis() error {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#01FAC6")).
		BorderLeftForeground(lipgloss.Color("#01FAC6"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#DDDDDD"))

	fileList := list.New([]list.Item{}, delegate, 0, 0)
	fileList.SetShowStatusBar(false)
	fileList.SetShowHelp(false)

	m := model{
		fileList:    fileList,
		currentView: OverviewView,
		loading:     true,
		tuiHelper: terminal.NewResponsiveTUIHelper(),
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
