package statusService

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/redjax/syst/internal/utils/terminal"
)

// StatusOptions configures the git status display
type StatusOptions struct {
	ShowAll    bool
	ShowColors bool
}

// FileStatus represents the status of a file in the repository
type FileStatus struct {
	Path    string
	Status  string // "tracked", "untracked", "modified", "staged", "deleted"
	IsDir   bool
	Size    int64
	ModTime string
}

// StatusInfo contains all file status information
type StatusInfo struct {
	CleanFiles     []FileStatus // Tracked files with no changes (show as normal text)
	UntrackedFiles []FileStatus // New files not tracked by git
	ModifiedFiles  []FileStatus // Files with changes in working directory
	StagedFiles    []FileStatus // Files staged for commit
	DeletedFiles   []FileStatus // Files deleted from working directory
}

// TUI Model and related types
type model struct {
	statusInfo *StatusInfo
	list       list.Model
	tuiHelper  *terminal.ResponsiveTUIHelper
	err        error
	loading    bool
}

type statusItem struct {
	file FileStatus
}

func (i statusItem) FilterValue() string { return i.file.Path }
func (i statusItem) Title() string {
	icon := getStatusIcon(i.file.Status)
	return fmt.Sprintf("%s %s", icon, i.file.Path)
}
func (i statusItem) Description() string {
	sizeStr := ""
	if i.file.Size > 0 {
		sizeStr = fmt.Sprintf(" (%s)", formatSize(i.file.Size))
	}
	dirIndicator := ""
	if i.file.IsDir {
		dirIndicator = "/"
	}
	return fmt.Sprintf("%s%s%s - %s", i.file.Status, dirIndicator, sizeStr, i.file.ModTime)
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1).
			Bold(true)

	sectionStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 2).
			Margin(1, 0)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)

	statsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	modifiedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00")) // Yellow

	deletedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")) // Red

	untrackedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#808080")) // Gray

	stagedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")) // Green
)

// TUI Messages
type dataLoadedMsg struct {
	statusInfo *StatusInfo
}

type errMsg struct {
	err error
}

// Model methods
func (m model) Init() tea.Cmd {
	return loadStatusData
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.tuiHelper.HandleWindowSizeMsg(msg)
		width, height := m.tuiHelper.GetSize()
		m.list.SetWidth(width)
		m.list.SetHeight(height - 10)
		return m, nil

	case dataLoadedMsg:
		m.statusInfo = msg.statusInfo
		m.loading = false

		// Create combined list of all files with status
		var items []list.Item

		// Add files with status first (modified, staged, deleted, untracked)
		for _, file := range m.statusInfo.ModifiedFiles {
			items = append(items, statusItem{file: file})
		}
		for _, file := range m.statusInfo.StagedFiles {
			items = append(items, statusItem{file: file})
		}
		for _, file := range m.statusInfo.DeletedFiles {
			items = append(items, statusItem{file: file})
		}
		for _, file := range m.statusInfo.UntrackedFiles {
			items = append(items, statusItem{file: file})
		}

		// Optionally add clean files at the end
		for _, file := range m.statusInfo.CleanFiles {
			items = append(items, statusItem{file: file})
		}

		m.list.SetItems(items)
		return m, nil

	case errMsg:
		m.err = msg.err
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))):
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.loading {
		return m.tuiHelper.CenterContent("Loading git status...")
	}

	if m.err != nil {
		return m.tuiHelper.CenterContent(fmt.Sprintf("Error: %v", m.err))
	}

	var sections []string

	title := titleStyle.Render("📋 Git Repository Status")
	sections = append(sections, title)

	// Status summary
	if m.statusInfo != nil {
		summary := m.renderStatusSummary()
		sections = append(sections, sectionStyle.Render(summary))
	}

	// File list
	sections = append(sections, m.list.View())

	help := helpStyle.Render("↑/↓: navigate • q: quit")
	sections = append(sections, help)

	return m.tuiHelper.CenterContent(strings.Join(sections, "\n"))
}

func (m model) renderStatusSummary() string {
	var content strings.Builder

	content.WriteString("📊 Summary:\n")
	content.WriteString(fmt.Sprintf("Modified: %s  ",
		modifiedStyle.Render(fmt.Sprintf("%d", len(m.statusInfo.ModifiedFiles)))))
	content.WriteString(fmt.Sprintf("Staged: %s  ",
		stagedStyle.Render(fmt.Sprintf("%d", len(m.statusInfo.StagedFiles)))))
	content.WriteString(fmt.Sprintf("Deleted: %s  ",
		deletedStyle.Render(fmt.Sprintf("%d", len(m.statusInfo.DeletedFiles)))))
	content.WriteString(fmt.Sprintf("Untracked: %s  ",
		untrackedStyle.Render(fmt.Sprintf("%d", len(m.statusInfo.UntrackedFiles)))))
	content.WriteString(fmt.Sprintf("Clean: %s",
		statsStyle.Render(fmt.Sprintf("%d", len(m.statusInfo.CleanFiles)))))

	return content.String()
}

// Helper functions
func getStatusIcon(status string) string {
	switch status {
	case "modified":
		return "🟡" // Yellow circle for modified
	case "staged":
		return "🟢" // Green circle for staged
	case "deleted":
		return "🔴" // Red circle for deleted
	case "untracked":
		return "⚫" // Gray circle for untracked
	default:
		return "  " // No icon for clean files
	}
}

func loadStatusData() tea.Msg {
	statusInfo, err := gatherStatusInfo()
	if err != nil {
		return errMsg{err}
	}
	return dataLoadedMsg{statusInfo}
}

// RunGitStatus displays the git status with tracked/untracked indicators
func RunGitStatus(opts StatusOptions) error {
	// Check if we're in a git repository
	if !isGitRepository() {
		return fmt.Errorf("not a git repository")
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#01FAC6")).
		BorderLeftForeground(lipgloss.Color("#01FAC6"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#DDDDDD"))

	statusList := list.New([]list.Item{}, delegate, 0, 0)
	statusList.Title = "Git Status"
	statusList.SetShowStatusBar(false)
	statusList.SetShowHelp(false)

	m := model{
		list:      statusList,
		loading:   true,
		tuiHelper: terminal.NewResponsiveTUIHelper(),
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// isGitRepository checks if we're in a git repository
func isGitRepository() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	err := cmd.Run()
	return err == nil
}

// gatherStatusInfo collects all file status information
func gatherStatusInfo() (*StatusInfo, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return nil, err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	// Get git status
	status, err := worktree.Status()
	if err != nil {
		return nil, err
	}

	statusInfo := &StatusInfo{
		CleanFiles:     []FileStatus{},
		UntrackedFiles: []FileStatus{},
		ModifiedFiles:  []FileStatus{},
		StagedFiles:    []FileStatus{},
		DeletedFiles:   []FileStatus{},
	}

	// Process git status
	for filePath, fileStatus := range status {
		info, err := os.Stat(filePath)
		var size int64
		var modTime string
		var isDir bool

		if err == nil {
			size = info.Size()
			modTime = info.ModTime().Format("2006-01-02 15:04")
			isDir = info.IsDir()
		}

		fs := FileStatus{
			Path:    filePath,
			IsDir:   isDir,
			Size:    size,
			ModTime: modTime,
		}

		// Categorize by status
		if fileStatus.Staging != git.Unmodified {
			fs.Status = "staged"
			statusInfo.StagedFiles = append(statusInfo.StagedFiles, fs)
		}

		if fileStatus.Worktree == git.Modified {
			fs.Status = "modified"
			statusInfo.ModifiedFiles = append(statusInfo.ModifiedFiles, fs)
		}

		if fileStatus.Worktree == git.Untracked {
			fs.Status = "untracked"
			statusInfo.UntrackedFiles = append(statusInfo.UntrackedFiles, fs)
		}

		if fileStatus.Worktree == git.Deleted {
			fs.Status = "deleted"
			statusInfo.DeletedFiles = append(statusInfo.DeletedFiles, fs)
		}
	}

	// Also get all tracked files
	cleanFiles, err := getAllTrackedFiles(repo)
	if err == nil {
		statusInfo.CleanFiles = cleanFiles
	}

	return statusInfo, nil
}

// getAllTrackedFiles gets all files tracked by git
func getAllTrackedFiles(repo *git.Repository) ([]FileStatus, error) {
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

	var trackedFiles []FileStatus

	err = tree.Files().ForEach(func(file *object.File) error {
		info, err := os.Stat(file.Name)
		var size int64
		var modTime string
		var isDir bool

		if err == nil {
			size = info.Size()
			modTime = info.ModTime().Format("2006-01-02 15:04")
			isDir = info.IsDir()
		} else {
			// File might have been deleted
			size = file.Size
		}

		trackedFiles = append(trackedFiles, FileStatus{
			Path:    file.Name,
			Status:  "tracked",
			IsDir:   isDir,
			Size:    size,
			ModTime: modTime,
		})
		return nil
	})

	return trackedFiles, err
}

// formatSize formats file size in human readable format
func formatSize(bytes int64) string {
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
