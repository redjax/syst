package statusService

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

		// Create combined list of only files with changes
		var items []list.Item

		// Add staged files first (highest priority)
		for _, file := range m.statusInfo.StagedFiles {
			items = append(items, statusItem{file: file})
		}
		// Add modified files
		for _, file := range m.statusInfo.ModifiedFiles {
			items = append(items, statusItem{file: file})
		}
		// Add deleted files
		for _, file := range m.statusInfo.DeletedFiles {
			items = append(items, statusItem{file: file})
		}
		// Add untracked files
		for _, file := range m.statusInfo.UntrackedFiles {
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
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			// Open selected file in editor
			if selected := m.list.SelectedItem(); selected != nil {
				if item, ok := selected.(statusItem); ok {
					go openFileInEditor(item.file.Path)
				}
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.loading {
		return "Loading git status..."
	}

	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
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

	help := helpStyle.Render("↑/↓: navigate • enter: open file • q: quit")
	sections = append(sections, help)

	return strings.Join(sections, "\n")
}

func (m model) renderStatusSummary() string {
	var content strings.Builder

	// Calculate total changes (excluding clean files)
	totalChanges := len(m.statusInfo.ModifiedFiles) + len(m.statusInfo.StagedFiles) + 
		len(m.statusInfo.DeletedFiles) + len(m.statusInfo.UntrackedFiles)

	content.WriteString("📊 Summary:\n")
	content.WriteString(fmt.Sprintf("Total Changes: %s  ",
		titleStyle.Render(fmt.Sprintf("%d", totalChanges))))
	
	if len(m.statusInfo.StagedFiles) > 0 {
		content.WriteString(fmt.Sprintf("Staged: %s  ",
			stagedStyle.Render(fmt.Sprintf("%d", len(m.statusInfo.StagedFiles)))))
	}
	if len(m.statusInfo.ModifiedFiles) > 0 {
		content.WriteString(fmt.Sprintf("Modified: %s  ",
			modifiedStyle.Render(fmt.Sprintf("%d", len(m.statusInfo.ModifiedFiles)))))
	}
	if len(m.statusInfo.DeletedFiles) > 0 {
		content.WriteString(fmt.Sprintf("Deleted: %s  ",
			deletedStyle.Render(fmt.Sprintf("%d", len(m.statusInfo.DeletedFiles)))))
	}
	if len(m.statusInfo.UntrackedFiles) > 0 {
		content.WriteString(fmt.Sprintf("Untracked: %s  ",
			untrackedStyle.Render(fmt.Sprintf("%d", len(m.statusInfo.UntrackedFiles)))))
	}
	
	if totalChanges == 0 {
		content.WriteString(stagedStyle.Render("✅ Working directory clean"))
	}

	return content.String()
}

// openFileInEditor opens a file in the default editor cross-platform
func openFileInEditor(filePath string) error {
	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "windows":
		// Try VS Code first, then notepad
		if _, err := exec.LookPath("code"); err == nil {
			cmd = exec.Command("code", filePath)
		} else {
			cmd = exec.Command("notepad", filePath)
		}
	case "darwin": // macOS
		// Try VS Code first, then default app
		if _, err := exec.LookPath("code"); err == nil {
			cmd = exec.Command("code", filePath)
		} else {
			cmd = exec.Command("open", filePath)
		}
	default: // Linux and other Unix-like systems
		// Try various editors in order of preference
		editors := []string{"code", "$EDITOR", "nano", "vi"}
		for _, editor := range editors {
			if editor == "$EDITOR" {
				if envEditor := os.Getenv("EDITOR"); envEditor != "" {
					if _, err := exec.LookPath(envEditor); err == nil {
						cmd = exec.Command(envEditor, filePath)
						break
					}
				}
				continue
			}
			if _, err := exec.LookPath(editor); err == nil {
				cmd = exec.Command(editor, filePath)
				break
			}
		}
		// Fallback to xdg-open
		if cmd == nil {
			cmd = exec.Command("xdg-open", filePath)
		}
	}
	
	if cmd == nil {
		return fmt.Errorf("no suitable editor found")
	}
	
	return cmd.Start()
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

// gatherStatusInfo collects all file status information using git porcelain format
func gatherStatusInfo() (*StatusInfo, error) {
	// Use git status --porcelain to get the exact same output as git status
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run git status: %w", err)
	}

	statusInfo := &StatusInfo{
		CleanFiles:     []FileStatus{},
		UntrackedFiles: []FileStatus{},
		ModifiedFiles:  []FileStatus{},
		StagedFiles:    []FileStatus{},
		DeletedFiles:   []FileStatus{},
	}

	// Parse porcelain output
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		// No changes  
		return statusInfo, nil
	}

	for _, line := range lines {
		if len(line) < 3 {
			continue
		}

		// Porcelain format: XY filename
		// X = staging area status, Y = working tree status
		// ' ' = unmodified, M = modified, A = added, D = deleted, R = renamed, C = copied, U = updated but unmerged
		stagingStatus := line[0]
		worktreeStatus := line[1]  
		// Take everything after position 2 (skip the two status chars)
		filePath := line[2:]
		// Trim leading space if present
		filePath = strings.TrimLeft(filePath, " ")

		// Get file info
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

		// Categorize based on porcelain status codes
		// Priority order: untracked, then working tree changes, then staged changes
		if stagingStatus == '?' && worktreeStatus == '?' {
			// File is untracked
			fs.Status = "untracked"
			statusInfo.UntrackedFiles = append(statusInfo.UntrackedFiles, fs)
		} else if worktreeStatus == 'M' {
			// File is modified in working tree
			fs.Status = "modified"
			statusInfo.ModifiedFiles = append(statusInfo.ModifiedFiles, fs)
		} else if worktreeStatus == 'D' {
			// File is deleted in working tree
			fs.Status = "deleted"
			statusInfo.DeletedFiles = append(statusInfo.DeletedFiles, fs)
		} else if stagingStatus != ' ' {
			// File has staged changes (A, M, D, R, C, etc. in staging area)
			fs.Status = "staged"
			statusInfo.StagedFiles = append(statusInfo.StagedFiles, fs)
		}
	}

	return statusInfo, nil
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
