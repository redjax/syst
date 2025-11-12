package blameService

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/redjax/syst/internal/utils/terminal"
)

type ViewMode int

const (
	FileListView ViewMode = iota
	BlameView
	FileHistoryView
	AuthorStatsView
	CommitDetailsView
	FileDiffView
)

type BlameAnalysis struct {
	FilePath      string
	BlameLines    []BlameLine
	AuthorStats   []AuthorContribution
	FileHistory   []FileCommit
	TotalLines    int
	LastModified  time.Time
	OldestChange  time.Time
	UniqueAuthors int
}

type BlameLine struct {
	LineNumber  int
	Content     string
	Author      string
	AuthorEmail string
	CommitHash  string
	CommitDate  time.Time
	CommitMsg   string
}

type AuthorContribution struct {
	Author      string
	Email       string
	Lines       int
	Percentage  float64
	FirstCommit time.Time
	LastCommit  time.Time
	Commits     int
}

type FileCommit struct {
	Hash      string
	Author    string
	Date      time.Time
	Message   string
	Changes   int
	Additions int
	Deletions int
}

type CommitDetails struct {
	Hash         string
	Author       string
	AuthorEmail  string
	Date         time.Time
	Message      string
	FullMessage  string
	Parents      []string
	FilesChanged []FileChange
	Stats        CommitStats
}

type FileChange struct {
	Path      string
	Status    string // "modified", "added", "deleted", "renamed"
	OldPath   string // For renames
	Additions int
	Deletions int
	Changes   []LineChange
}

type LineChange struct {
	Type    string // "added", "deleted", "context"
	LineNum int
	Content string
}

type CommitStats struct {
	FilesChanged int
	Additions    int
	Deletions    int
	TotalChanges int
}

type FileItem struct {
	path         string
	name         string
	isDirectory  bool
	size         int64
	lastModified time.Time
}

func (f FileItem) Title() string {
	if f.isDirectory {
		return "📁 " + f.name
	}
	return "📄 " + f.name
}

func (f FileItem) Description() string {
	if f.isDirectory {
		return f.path
	}
	sizeStr := formatFileSize(f.size)
	return fmt.Sprintf("%s • %s • %s", f.path, sizeStr, f.lastModified.Format("2006-01-02"))
}

func (f FileItem) FilterValue() string {
	return f.name + " " + f.path
}

type BlameLineItem struct {
	line BlameLine
}

func (b BlameLineItem) Title() string {
	return fmt.Sprintf("%4d │ %s", b.line.LineNumber, b.line.Content)
}

func (b BlameLineItem) Description() string {
	return fmt.Sprintf("%s • %s • %s",
		b.line.Author,
		b.line.CommitHash[:8],
		b.line.CommitDate.Format("2006-01-02"))
}

func (b BlameLineItem) FilterValue() string {
	return fmt.Sprintf("%d %s %s", b.line.LineNumber, b.line.Content, b.line.Author)
}

// FileChangeItem for commit details list
type FileChangeItem struct {
	change FileChange
}

func (f FileChangeItem) Title() string {
	statusIcon := "📝"
	switch f.change.Status {
	case "added":
		statusIcon = "➕"
	case "deleted":
		statusIcon = "❌"
	case "renamed":
		statusIcon = "📝"
	case "modified":
		statusIcon = "📝"
	}
	return fmt.Sprintf("%s %s", statusIcon, f.change.Path)
}

func (f FileChangeItem) Description() string {
	if f.change.Status == "renamed" && f.change.OldPath != "" {
		return fmt.Sprintf("Renamed from %s • +%d -%d", f.change.OldPath, f.change.Additions, f.change.Deletions)
	}
	return fmt.Sprintf("%s • +%d -%d", f.change.Status, f.change.Additions, f.change.Deletions)
}

func (f FileChangeItem) FilterValue() string {
	return f.change.Path + " " + f.change.Status
}

type model struct {
	// Current state
	currentView        ViewMode
	selectedFile       string
	analysis           BlameAnalysis
	commitDetails      CommitDetails
	selectedCommit     string
	selectedFileChange FileChange

	// UI components
	fileList    list.Model
	blameList   list.Model
	historyList list.Model
	commitList  list.Model
	searchInput textinput.Model

	// Data
	files       []FileItem
	currentPath string

	// UI state
	loading    bool
	err        error
	tuiHelper  *terminal.ResponsiveTUIHelper
	showSearch bool
}

type filesLoadedMsg struct {
	files []FileItem
}

type blameAnalysisMsg struct {
	analysis BlameAnalysis
}

type commitDetailsMsg struct {
	details CommitDetails
}

type errMsg struct {
	err error
}

// RunBlameViewer starts the interactive blame viewer TUI
func RunBlameViewer(args []string) error {
	// Open the repository
	repo, err := git.PlainOpen(".")
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Initialize the model
	m := initModel(repo, args)

	// Start the TUI
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func initModel(repo *git.Repository, args []string) model {
	// Initialize file list
	fileList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	fileList.Title = "📁 Repository Files"
	fileList.SetShowStatusBar(false)
	fileList.SetFilteringEnabled(true)
	fileList.SetShowPagination(true)

	// Initialize blame list
	blameList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	blameList.Title = "🔍 File Blame"
	blameList.SetShowStatusBar(false)
	blameList.SetFilteringEnabled(true)
	blameList.SetShowPagination(true)

	// Initialize history list
	historyList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	historyList.Title = "📜 File History"
	historyList.SetShowStatusBar(false)
	historyList.SetFilteringEnabled(false)
	historyList.SetShowPagination(true)

	// Initialize commit details list
	commitList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	commitList.Title = "📝 Commit Details"
	commitList.SetShowStatusBar(false)
	commitList.SetFilteringEnabled(false)
	commitList.SetShowPagination(true)

	// Initialize search input
	searchInput := textinput.New()
	searchInput.Placeholder = "Search files..."
	searchInput.CharLimit = 100

	// Determine starting file/path
	startingPath := "."
	selectedFile := ""
	if len(args) > 0 && args[0] != "" {
		if isFile(args[0]) {
			selectedFile = args[0]
			startingPath = filepath.Dir(args[0])
		} else {
			startingPath = args[0]
		}
	}

	m := model{
		currentView:  FileListView,
		selectedFile: selectedFile,
		fileList:     fileList,
		blameList:    blameList,
		historyList:  historyList,
		commitList:   commitList,
		searchInput:  searchInput,
		currentPath:  startingPath,
		loading:      true,
		tuiHelper:    terminal.NewResponsiveTUIHelper(),
	}

	return m
}

func (m model) Init() tea.Cmd {
	if m.selectedFile != "" {
		// If a specific file was provided, load its blame directly
		return tea.Batch(
			loadFiles(m.currentPath),
			loadBlameAnalysis(m.selectedFile),
		)
	}
	return loadFiles(m.currentPath)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.tuiHelper.HandleWindowSizeMsg(msg)

		listHeight := m.tuiHelper.GetHeight() - 8
		m.fileList.SetSize(m.tuiHelper.GetWidth()-4, listHeight)
		m.blameList.SetSize(m.tuiHelper.GetWidth()-4, listHeight)
		m.historyList.SetSize(m.tuiHelper.GetWidth()-4, listHeight)

	case filesLoadedMsg:
		m.loading = false
		m.files = msg.files

		items := make([]list.Item, len(msg.files))
		for i, file := range msg.files {
			items[i] = file
		}
		m.fileList.SetItems(items)

		// If we have a selected file, switch to blame view
		if m.selectedFile != "" {
			m.currentView = BlameView
		}

	case blameAnalysisMsg:
		m.loading = false
		m.analysis = msg.analysis

		// Update blame list
		blameItems := make([]list.Item, len(msg.analysis.BlameLines))
		for i, line := range msg.analysis.BlameLines {
			blameItems[i] = BlameLineItem{line: line}
		}
		m.blameList.SetItems(blameItems)
		m.blameList.Title = fmt.Sprintf("🔍 Blame: %s", m.analysis.FilePath)

		// Update history list
		historyItems := make([]list.Item, len(msg.analysis.FileHistory))
		for i, commit := range msg.analysis.FileHistory {
			historyItems[i] = FileCommitItem{commit: commit}
		}
		m.historyList.SetItems(historyItems)

	case commitDetailsMsg:
		m.loading = false
		m.commitDetails = msg.details

		// Update commit details list with file changes
		commitItems := make([]list.Item, len(msg.details.FilesChanged))
		for i, fileChange := range msg.details.FilesChanged {
			commitItems[i] = FileChangeItem{change: fileChange}
		}
		m.commitList.SetItems(commitItems)
		m.commitList.Title = fmt.Sprintf("📝 Commit: %s", msg.details.Hash[:8])

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
			if m.currentView != FileListView {
				m.currentView = FileListView
				return m, nil
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("/"))):
			if m.currentView == FileListView {
				m.showSearch = !m.showSearch
				if m.showSearch {
					m.searchInput.Focus()
				} else {
					m.searchInput.Blur()
				}
				return m, nil
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("1"))):
			m.currentView = FileListView
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("2"))):
			if m.selectedFile != "" {
				m.currentView = BlameView
			}
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("3"))):
			if m.selectedFile != "" {
				m.currentView = FileHistoryView
			}
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("4"))):
			if m.selectedFile != "" {
				m.currentView = AuthorStatsView
			}
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("5"))):
			if m.selectedCommit != "" {
				m.currentView = CommitDetailsView
			}
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
			m.loading = true
			if m.selectedFile != "" {
				return m, loadBlameAnalysis(m.selectedFile)
			}
			return m, loadFiles(m.currentPath)
		}

		// Handle view-specific keys
		if m.showSearch {
			switch msg.Type {
			case tea.KeyEnter:
				// Perform search
				query := m.searchInput.Value()
				if query != "" {
					// Filter file list
					filteredFiles := make([]FileItem, 0)
					for _, file := range m.files {
						if strings.Contains(strings.ToLower(file.name), strings.ToLower(query)) ||
							strings.Contains(strings.ToLower(file.path), strings.ToLower(query)) {
							filteredFiles = append(filteredFiles, file)
						}
					}
					items := make([]list.Item, len(filteredFiles))
					for i, file := range filteredFiles {
						items[i] = file
					}
					m.fileList.SetItems(items)
				}
				m.showSearch = false
				m.searchInput.Blur()
				return m, nil
			}
			m.searchInput, cmd = m.searchInput.Update(msg)
			return m, cmd
		}

		switch m.currentView {
		case FileListView:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				if item, ok := m.fileList.SelectedItem().(FileItem); ok {
					if item.isDirectory {
						// Navigate into directory
						m.currentPath = item.path
						m.loading = true
						return m, loadFiles(item.path)
					} else {
						// Load blame for file
						m.selectedFile = item.path
						m.loading = true
						m.currentView = BlameView
						return m, loadBlameAnalysis(item.path)
					}
				}
			}
			m.fileList, cmd = m.fileList.Update(msg)

		case BlameView:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				if item, ok := m.blameList.SelectedItem().(BlameLineItem); ok {
					// Load commit details for the selected blame line
					m.selectedCommit = item.line.CommitHash
					m.loading = true
					m.currentView = CommitDetailsView
					return m, loadCommitDetails(item.line.CommitHash)
				}
			}
			m.blameList, cmd = m.blameList.Update(msg)

		case FileHistoryView:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				if item, ok := m.historyList.SelectedItem().(FileCommitItem); ok {
					// Load commit details for the selected history item
					m.selectedCommit = item.commit.Hash
					m.loading = true
					m.currentView = CommitDetailsView
					return m, loadCommitDetails(item.commit.Hash)
				}
			}
			m.historyList, cmd = m.historyList.Update(msg)

		case CommitDetailsView:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				if item, ok := m.commitList.SelectedItem().(FileChangeItem); ok {
					// Load diff view for the selected file
					m.selectedFileChange = item.change
					m.currentView = FileDiffView
					return m, nil
				}
			}
			m.commitList, cmd = m.commitList.Update(msg)

		case FileDiffView:
			// No specific key handling needed, just allow navigation back

		case AuthorStatsView:
			// No specific handling needed for author stats view
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
	case FileListView:
		return m.renderFileList()
	case BlameView:
		return m.renderBlameView()
	case FileHistoryView:
		return m.renderHistoryView()
	case AuthorStatsView:
		return m.renderAuthorStatsView()
	case CommitDetailsView:
		return m.renderCommitDetailsView()
	case FileDiffView:
		return m.renderFileDiffView()
	default:
		return m.renderFileList()
	}
}

func loadFiles(path string) tea.Cmd {
	return func() tea.Msg {
		files, err := getRepositoryFiles(path)
		if err != nil {
			return errMsg{err}
		}
		return filesLoadedMsg{files}
	}
}

func loadBlameAnalysis(filePath string) tea.Cmd {
	return func() tea.Msg {
		analysis, err := analyzeFileBlame(filePath)
		if err != nil {
			return errMsg{err}
		}
		return blameAnalysisMsg{analysis}
	}
}

func loadCommitDetails(commitHash string) tea.Cmd {
	return func() tea.Msg {
		details, err := analyzeCommitDetails(commitHash)
		if err != nil {
			return errMsg{err}
		}
		return commitDetailsMsg{details}
	}
}

// FileCommitItem for history list
type FileCommitItem struct {
	commit FileCommit
}

func (f FileCommitItem) Title() string {
	return fmt.Sprintf("%s • %s", f.commit.Hash[:8], f.commit.Message)
}

func (f FileCommitItem) Description() string {
	return fmt.Sprintf("%s • %s • +%d -%d",
		f.commit.Author,
		f.commit.Date.Format("2006-01-02 15:04"),
		f.commit.Additions,
		f.commit.Deletions)
}

func (f FileCommitItem) FilterValue() string {
	return f.commit.Message + " " + f.commit.Author
}

// Analysis functions
func getRepositoryFiles(rootPath string) ([]FileItem, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return nil, err
	}

	// Get HEAD commit
	ref, err := repo.Head()
	if err != nil {
		return nil, err
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, err
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	var files []FileItem

	// Add parent directory if not at root
	if rootPath != "." && rootPath != "" {
		parentPath := filepath.Dir(rootPath)
		if parentPath == "." {
			parentPath = ""
		}
		files = append(files, FileItem{
			path:        parentPath,
			name:        "..",
			isDirectory: true,
		})
	}

	// Get files from git tree
	err = tree.Files().ForEach(func(file *object.File) error {
		relPath := file.Name

		// Filter by current path
		if rootPath != "." && rootPath != "" {
			if !strings.HasPrefix(relPath, rootPath+"/") {
				return nil
			}
			relPath = strings.TrimPrefix(relPath, rootPath+"/")
		}

		// Skip nested directories for current view
		if strings.Contains(relPath, "/") {
			// This is in a subdirectory, add the directory if not already added
			dirName := strings.Split(relPath, "/")[0]
			dirPath := rootPath
			if dirPath != "." && dirPath != "" {
				dirPath = dirPath + "/" + dirName
			} else {
				dirPath = dirName
			}

			// Check if we already have this directory
			found := false
			for _, existing := range files {
				if existing.path == dirPath && existing.isDirectory {
					found = true
					break
				}
			}

			if !found {
				files = append(files, FileItem{
					path:         dirPath,
					name:         dirName,
					isDirectory:  true,
					lastModified: time.Now(),
				})
			}
			return nil
		}

		// Add file
		filePath := rootPath
		if filePath != "." && filePath != "" {
			filePath = filePath + "/" + relPath
		} else {
			filePath = relPath
		}

		files = append(files, FileItem{
			path:         filePath,
			name:         relPath,
			isDirectory:  false,
			size:         file.Size,
			lastModified: time.Now(), // We could get this from git but it's expensive
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort files: directories first, then files, both alphabetically
	sort.Slice(files, func(i, j int) bool {
		if files[i].isDirectory != files[j].isDirectory {
			return files[i].isDirectory
		}
		return files[i].name < files[j].name
	})

	return files, nil
}

func analyzeFileBlame(filePath string) (BlameAnalysis, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return BlameAnalysis{}, err
	}

	// Read file content first
	// #nosec G304 - CLI tool reads user-specified files by design
	content, err := os.ReadFile(filePath)
	if err != nil {
		return BlameAnalysis{}, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	lines := strings.Split(string(content), "\n")

	// For now, create a simple blame analysis without git blame
	// This is a simplified version until we can get the git blame API working
	var blameLines []BlameLine
	authorContribs := make(map[string]*AuthorContribution)

	// Get the latest commit info for the file
	ref, err := repo.Head()
	if err != nil {
		return BlameAnalysis{}, err
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return BlameAnalysis{}, err
	}

	// Create simplified blame lines (all attributed to the latest commit for now)
	author := commit.Author.Name
	authorEmail := commit.Author.Email
	commitDate := commit.Author.When
	commitHash := commit.Hash.String()
	commitMsg := strings.Split(commit.Message, "\n")[0]

	for i, line := range lines {
		blameLines = append(blameLines, BlameLine{
			LineNumber:  i + 1,
			Content:     line,
			Author:      author,
			AuthorEmail: authorEmail,
			CommitHash:  commitHash,
			CommitDate:  commitDate,
			CommitMsg:   commitMsg,
		})
	}

	// Track author contributions
	authorContribs[author] = &AuthorContribution{
		Author:      author,
		Email:       authorEmail,
		Lines:       len(lines),
		FirstCommit: commitDate,
		LastCommit:  commitDate,
		Percentage:  100.0,
	}

	// Get file history
	history, err := getFileHistory(repo, filePath)
	if err != nil {
		history = []FileCommit{} // Don't fail if we can't get history
	}

	// Create author stats
	var authorStats []AuthorContribution
	for _, contrib := range authorContribs {
		authorStats = append(authorStats, *contrib)
	}

	return BlameAnalysis{
		FilePath:      filePath,
		BlameLines:    blameLines,
		AuthorStats:   authorStats,
		FileHistory:   history,
		TotalLines:    len(lines),
		LastModified:  commitDate,
		OldestChange:  commitDate,
		UniqueAuthors: len(authorStats),
	}, nil
}

func analyzeCommitDetails(commitHash string) (CommitDetails, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return CommitDetails{}, err
	}

	// Parse the commit hash
	hash := plumbing.NewHash(commitHash)

	// Get the commit object
	commit, err := repo.CommitObject(hash)
	if err != nil {
		return CommitDetails{}, fmt.Errorf("failed to get commit %s: %w", commitHash, err)
	}

	// Get parent commits
	var parents []string
	parentIter := commit.Parents()
	err = parentIter.ForEach(func(parent *object.Commit) error {
		parents = append(parents, parent.Hash.String())
		return nil
	})
	if err != nil {
		return CommitDetails{}, fmt.Errorf("failed to get parent commits: %w", err)
	}

	// Get commit stats
	stats, err := commit.Stats()
	if err != nil {
		return CommitDetails{}, fmt.Errorf("failed to get commit stats: %w", err)
	}

	// Generate diff for each file
	var filesChanged []FileChange
	totalAdditions := 0
	totalDeletions := 0

	// Get parent commit if available for diff
	var parentCommit *object.Commit
	if len(parents) > 0 {
		parentHash := plumbing.NewHash(parents[0])
		parentCommit, _ = repo.CommitObject(parentHash)
	}

	for _, stat := range stats {
		totalAdditions += stat.Addition
		totalDeletions += stat.Deletion

		// Determine file status
		status := "modified"
		if stat.Addition > 0 && stat.Deletion == 0 {
			status = "added"
		} else if stat.Addition == 0 && stat.Deletion > 0 {
			status = "deleted"
		}

		// Generate line changes for this file
		lineChanges := generateFileChanges(repo, commit, parentCommit, stat.Name)

		filesChanged = append(filesChanged, FileChange{
			Path:      stat.Name,
			Status:    status,
			Additions: stat.Addition,
			Deletions: stat.Deletion,
			Changes:   lineChanges,
		})
	}

	commitStats := CommitStats{
		FilesChanged: len(filesChanged),
		Additions:    totalAdditions,
		Deletions:    totalDeletions,
		TotalChanges: totalAdditions + totalDeletions,
	}

	return CommitDetails{
		Hash:         commit.Hash.String(),
		Author:       commit.Author.Name,
		AuthorEmail:  commit.Author.Email,
		Date:         commit.Author.When,
		Message:      strings.Split(commit.Message, "\n")[0], // First line
		FullMessage:  commit.Message,
		Parents:      parents,
		FilesChanged: filesChanged,
		Stats:        commitStats,
	}, nil
}

func generateFileChanges(repo *git.Repository, commit *object.Commit, parentCommit *object.Commit, filePath string) []LineChange {
	var changes []LineChange

	// For simplicity, let's just show a summary instead of full diff
	// Full diff generation can be complex and might be too much for this view

	if parentCommit == nil {
		// New file
		changes = append(changes, LineChange{
			Type:    "info",
			LineNum: 0,
			Content: "📄 New file created in this commit",
		})

		// Show first few lines of the file as preview
		if file, err := commit.File(filePath); err == nil {
			if content, err := file.Contents(); err == nil {
				lines := strings.Split(content, "\n")
				maxLines := 10
				if len(lines) > maxLines {
					changes = append(changes, LineChange{
						Type:    "info",
						LineNum: 0,
						Content: fmt.Sprintf("📋 File preview (first %d lines of %d):", maxLines, len(lines)),
					})
				} else {
					changes = append(changes, LineChange{
						Type:    "info",
						LineNum: 0,
						Content: fmt.Sprintf("📋 File content (%d lines):", len(lines)),
					})
				}

				for i, line := range lines {
					if i >= maxLines {
						changes = append(changes, LineChange{
							Type:    "info",
							LineNum: 0,
							Content: "   ... (truncated)",
						})
						break
					}
					changes = append(changes, LineChange{
						Type:    "added",
						LineNum: i + 1,
						Content: fmt.Sprintf("+ %s", line),
					})
				}
			}
		}
		return changes
	}

	// For modified files, show a simplified summary
	currentFile, err := commit.File(filePath)
	if err != nil {
		// File was deleted
		changes = append(changes, LineChange{
			Type:    "info",
			LineNum: 0,
			Content: "🗑️ File was deleted in this commit",
		})

		// Show last few lines of the deleted file
		if parentFile, parentErr := parentCommit.File(filePath); parentErr == nil {
			if content, contentErr := parentFile.Contents(); contentErr == nil {
				lines := strings.Split(content, "\n")
				maxLines := 10
				startLine := 0
				if len(lines) > maxLines {
					startLine = len(lines) - maxLines
					changes = append(changes, LineChange{
						Type:    "info",
						LineNum: 0,
						Content: fmt.Sprintf("📋 Deleted file preview (last %d lines of %d):", maxLines, len(lines)),
					})
				} else {
					changes = append(changes, LineChange{
						Type:    "info",
						LineNum: 0,
						Content: fmt.Sprintf("📋 Deleted file content (%d lines):", len(lines)),
					})
				}

				for i := startLine; i < len(lines) && i < startLine+maxLines; i++ {
					changes = append(changes, LineChange{
						Type:    "deleted",
						LineNum: i + 1,
						Content: fmt.Sprintf("- %s", lines[i]),
					})
				}
			}
		}
		return changes
	}

	// File was modified - show basic info instead of complex diff
	changes = append(changes, LineChange{
		Type:    "info",
		LineNum: 0,
		Content: "📝 File was modified in this commit",
	})

	// Get basic file info
	if currentContent, err := currentFile.Contents(); err == nil {
		currentLines := strings.Split(currentContent, "\n")

		if parentFile, parentErr := parentCommit.File(filePath); parentErr == nil {
			if parentContent, contentErr := parentFile.Contents(); contentErr == nil {
				parentLines := strings.Split(parentContent, "\n")

				changes = append(changes, LineChange{
					Type:    "info",
					LineNum: 0,
					Content: fmt.Sprintf("📊 Lines: %d → %d (change: %+d)", len(parentLines), len(currentLines), len(currentLines)-len(parentLines)),
				})

				// Show a few context lines around changes (simplified)
				changes = append(changes, LineChange{
					Type:    "info",
					LineNum: 0,
					Content: "💡 Use 'git show <commit>' or 'git diff <commit>^..<commit>' for full diff",
				})
			}
		}
	}

	return changes
}

func getFileHistory(repo *git.Repository, filePath string) ([]FileCommit, error) {
	// Get commit history for the file
	commits, err := repo.Log(&git.LogOptions{
		FileName: &filePath,
	})
	if err != nil {
		return nil, err
	}

	var history []FileCommit
	err = commits.ForEach(func(commit *object.Commit) error {
		// Get file stats for this commit
		stats, err := commit.Stats()
		if err != nil {
			// If we can't get stats, still add the commit with minimal info
			history = append(history, FileCommit{
				Hash:    commit.Hash.String(),
				Author:  commit.Author.Name,
				Date:    commit.Author.When,
				Message: strings.Split(commit.Message, "\n")[0], // First line only
			})
			return nil
		}

		// Find stats for our specific file
		var additions, deletions int
		for _, stat := range stats {
			if stat.Name == filePath {
				additions = stat.Addition
				deletions = stat.Deletion
				break
			}
		}

		history = append(history, FileCommit{
			Hash:      commit.Hash.String(),
			Author:    commit.Author.Name,
			Date:      commit.Author.When,
			Message:   strings.Split(commit.Message, "\n")[0], // First line only
			Changes:   additions + deletions,
			Additions: additions,
			Deletions: deletions,
		})

		// Limit to last 50 commits to avoid overwhelming the UI
		if len(history) >= 50 {
			return fmt.Errorf("limit reached") // Use error to break the loop
		}

		return nil
	})

	// Filter out the "limit reached" error
	if err != nil && !strings.Contains(err.Error(), "limit reached") {
		return nil, err
	}

	return history, nil
}

// Rendering functions
func (m model) renderLoading() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(2, 4).
		Align(lipgloss.Center).
		Width(50)

	content := "🔍 Loading blame analysis...\n\n"
	if m.selectedFile != "" {
		content += "Analyzing: " + m.selectedFile
	} else {
		content += "Loading repository files..."
	}

	return lipgloss.Place(m.tuiHelper.GetWidth(), m.tuiHelper.GetHeight(), lipgloss.Center, lipgloss.Center, style.Render(content))
}

func (m model) renderError() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(2, 4).
		Align(lipgloss.Center).
		Width(60)

	content := "❌ Error\n\n" + m.err.Error()
	return lipgloss.Place(m.tuiHelper.GetWidth(), m.tuiHelper.GetHeight(), lipgloss.Center, lipgloss.Center, style.Render(content))
}

func (m model) renderFileList() string {
	var content strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	content.WriteString(headerStyle.Render("🔍 File Blame Viewer"))
	content.WriteString("\n")

	// Search box if active
	if m.showSearch {
		searchStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1).
			MarginBottom(1)

		content.WriteString(searchStyle.Render("Search: " + m.searchInput.View()))
		content.WriteString("\n")
	}

	// File list
	content.WriteString(m.fileList.View())
	content.WriteString("\n")

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	help := "enter: open • /: search • q: quit"
	if m.selectedFile != "" {
		help = "enter: open • 2: blame • 3: history • 4: authors • /: search • q: quit"
	}

	content.WriteString(helpStyle.Render(help))

	return content.String()
}

func (m model) renderBlameView() string {
	var content strings.Builder

	// Header with file info
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	title := fmt.Sprintf("🔍 Blame: %s", m.analysis.FilePath)
	content.WriteString(headerStyle.Render(title))
	content.WriteString("\n")

	// Stats summary
	statsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("242")).
		MarginBottom(1)

	stats := fmt.Sprintf("Lines: %d • Authors: %d • Last modified: %s",
		m.analysis.TotalLines,
		m.analysis.UniqueAuthors,
		m.analysis.LastModified.Format("2006-01-02 15:04"))

	content.WriteString(statsStyle.Render(stats))
	content.WriteString("\n")

	// Blame list
	content.WriteString(m.blameList.View())
	content.WriteString("\n")

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	help := "1: files • 3: history • 4: authors • enter: commit details • esc: back • q: quit"
	content.WriteString(helpStyle.Render(help))

	return content.String()
}

func (m model) renderHistoryView() string {
	var content strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	title := fmt.Sprintf("📜 History: %s", m.analysis.FilePath)
	content.WriteString(headerStyle.Render(title))
	content.WriteString("\n")

	// Stats
	statsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("242")).
		MarginBottom(1)

	stats := fmt.Sprintf("Commits: %d • First change: %s",
		len(m.analysis.FileHistory),
		m.analysis.OldestChange.Format("2006-01-02"))

	content.WriteString(statsStyle.Render(stats))
	content.WriteString("\n")

	// History list
	content.WriteString(m.historyList.View())
	content.WriteString("\n")

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	help := "1: files • 2: blame • 4: authors • enter: commit details • esc: back • q: quit"
	content.WriteString(helpStyle.Render(help))

	return content.String()
}

func (m model) renderAuthorStatsView() string {
	var content strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	title := fmt.Sprintf("👥 Authors: %s", m.analysis.FilePath)
	content.WriteString(headerStyle.Render(title))
	content.WriteString("\n")

	// Author statistics
	tableStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("238")).
		Padding(1, 2).
		MarginBottom(1)

	var table strings.Builder
	table.WriteString("Author                    Lines    %      First Commit   Last Commit\n")
	table.WriteString("─────────────────────────────────────────────────────────────────────\n")

	for i, author := range m.analysis.AuthorStats {
		if i >= 10 { // Limit to top 10 authors
			break
		}

		table.WriteString(fmt.Sprintf("%-25s %5d %6.1f%%  %s  %s\n",
			truncateString(author.Author, 25),
			author.Lines,
			author.Percentage,
			author.FirstCommit.Format("2006-01-02"),
			author.LastCommit.Format("2006-01-02")))
	}

	content.WriteString(tableStyle.Render(table.String()))

	// Summary
	summaryStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("242")).
		MarginBottom(1)

	if len(m.analysis.AuthorStats) > 0 {
		topAuthor := m.analysis.AuthorStats[0]
		summary := fmt.Sprintf("Top contributor: %s (%.1f%% of lines)",
			topAuthor.Author, topAuthor.Percentage)
		content.WriteString(summaryStyle.Render(summary))
		content.WriteString("\n")
	}

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	help := "1: files • 2: blame • 3: history • esc: back • q: quit"
	content.WriteString(helpStyle.Render(help))

	return content.String()
}

func (m model) renderCommitDetailsView() string {
	var content strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	title := fmt.Sprintf("📝 Commit: %s", m.commitDetails.Hash[:8])
	content.WriteString(headerStyle.Render(title))
	content.WriteString("\n")

	// Commit info
	infoStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("238")).
		Padding(1, 2).
		MarginBottom(1)

	var info strings.Builder
	info.WriteString(fmt.Sprintf("Author:    %s <%s>\n", m.commitDetails.Author, m.commitDetails.AuthorEmail))
	info.WriteString(fmt.Sprintf("Date:      %s\n", m.commitDetails.Date.Format("2006-01-02 15:04:05")))
	info.WriteString(fmt.Sprintf("Hash:      %s\n", m.commitDetails.Hash))
	if len(m.commitDetails.Parents) > 0 {
		info.WriteString(fmt.Sprintf("Parents:   %s\n", strings.Join(m.commitDetails.Parents, ", ")[:40]+"..."))
	}
	info.WriteString("\n")
	info.WriteString(fmt.Sprintf("Message:   %s\n", m.commitDetails.Message))

	if m.commitDetails.FullMessage != m.commitDetails.Message {
		// Show first few lines of full message if different from summary
		fullLines := strings.Split(m.commitDetails.FullMessage, "\n")
		if len(fullLines) > 1 {
			info.WriteString("\nFull Message:\n")
			for i, line := range fullLines[:min(5, len(fullLines))] {
				if i == 0 {
					continue // Skip first line as it's already shown
				}
				info.WriteString(fmt.Sprintf("           %s\n", line))
			}
			if len(fullLines) > 5 {
				info.WriteString("           ...\n")
			}
		}
	}

	content.WriteString(infoStyle.Render(info.String()))

	// Stats summary
	statsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("242")).
		MarginBottom(1)

	stats := fmt.Sprintf("Files: %d • Additions: +%d • Deletions: -%d • Total: %d",
		m.commitDetails.Stats.FilesChanged,
		m.commitDetails.Stats.Additions,
		m.commitDetails.Stats.Deletions,
		m.commitDetails.Stats.TotalChanges)

	content.WriteString(statsStyle.Render(stats))
	content.WriteString("\n")

	// File changes list
	if len(m.commitDetails.FilesChanged) > 0 {
		content.WriteString(m.commitList.View())
		content.WriteString("\n")
	}

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	help := "1: files • 2: blame • 3: history • 4: authors • enter: file diff • esc: back • q: quit"
	content.WriteString(helpStyle.Render(help))

	return content.String()
}

func (m model) renderFileDiffView() string {
	var content strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	statusIcon := "📝"
	if m.selectedFileChange.Status == "added" {
		statusIcon = "✅"
	} else if m.selectedFileChange.Status == "deleted" {
		statusIcon = "❌"
	}

	title := fmt.Sprintf("%s %s (%s)", statusIcon, m.selectedFileChange.Path, m.selectedFileChange.Status)
	content.WriteString(headerStyle.Render(title))
	content.WriteString("\n")

	// File stats
	statsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("242")).
		MarginBottom(1)

	stats := fmt.Sprintf("Additions: +%d • Deletions: -%d • Total: %d",
		m.selectedFileChange.Additions,
		m.selectedFileChange.Deletions,
		m.selectedFileChange.Additions+m.selectedFileChange.Deletions)

	content.WriteString(statsStyle.Render(stats))
	content.WriteString("\n")

	// Diff content
	if len(m.selectedFileChange.Changes) > 0 {
		diffStyle := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("238")).
			Padding(1, 2).
			MarginBottom(1)

		var diff strings.Builder

		for _, change := range m.selectedFileChange.Changes {
			var lineStyle lipgloss.Style
			switch change.Type {
			case "added":
				lineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("34")) // green
			case "deleted":
				lineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("31")) // red
			case "context":
				lineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245")) // gray
			case "info":
				lineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true) // yellow, bold
			default:
				lineStyle = lipgloss.NewStyle()
			}

			diff.WriteString(lineStyle.Render(change.Content))
			diff.WriteString("\n")
		}

		content.WriteString(diffStyle.Render(diff.String()))
	} else {
		// No detailed changes available
		noChangesStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true).
			MarginBottom(1)

		if m.selectedFileChange.Status == "added" {
			content.WriteString(noChangesStyle.Render("New file - entire file was added"))
		} else if m.selectedFileChange.Status == "deleted" {
			content.WriteString(noChangesStyle.Render("File was deleted"))
		} else {
			content.WriteString(noChangesStyle.Render("Binary file or changes could not be displayed"))
		}
		content.WriteString("\n")
	}

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	help := "1: files • 2: blame • 3: history • 4: authors • 5: commit details • esc: back • q: quit"
	content.WriteString(helpStyle.Render(help))

	return content.String()
}

// Helper functions
func isFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func formatFileSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.1fGB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.1fMB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.1fKB", float64(size)/KB)
	default:
		return fmt.Sprintf("%dB", size)
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
