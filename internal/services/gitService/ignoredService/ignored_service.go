package ignoredService

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/redjax/syst/internal/utils/terminal"
)

// IgnoredOptions configures the ignored files display
type IgnoredOptions struct {
	ShowAll    bool
	ShowSizes  bool
	OutputPath string
}

// IgnoredFile represents an ignored file or directory
type IgnoredFile struct {
	Path    string
	IsDir   bool
	Size    int64
	ModTime string
}

// TUI Model and related types
type model struct {
	ignoredFiles []IgnoredFile
	list         list.Model
	tuiHelper    *terminal.ResponsiveTUIHelper
	err          error
	loading      bool
	options      IgnoredOptions
}

type ignoredItem struct {
	file IgnoredFile
}

func (i ignoredItem) FilterValue() string { return i.file.Path }
func (i ignoredItem) Title() string {
	icon := "📄"
	if i.file.IsDir {
		icon = "📁"
	}
	return fmt.Sprintf("%s %s", icon, i.file.Path)
}
func (i ignoredItem) Description() string {
	sizeStr := ""
	if i.file.Size > 0 && !i.file.IsDir {
		sizeStr = fmt.Sprintf(" (%s)", formatSize(i.file.Size))
	}
	dirIndicator := ""
	if i.file.IsDir {
		dirIndicator = "/"
	}
	return fmt.Sprintf("ignored%s%s - %s", dirIndicator, sizeStr, i.file.ModTime)
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#874BFD")).
			Padding(0, 1).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)

	ignoredStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#808080")) // Gray for ignored files
)

// TUI Messages
type dataLoadedMsg struct {
	ignoredFiles []IgnoredFile
}

type errMsg struct {
	err error
}

// Model methods
func (m model) Init() tea.Cmd {
	return loadIgnoredData(m.options)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.tuiHelper.HandleWindowSizeMsg(msg)
		width, height := m.tuiHelper.GetSize()
		m.list.SetWidth(width)
		m.list.SetHeight(height - 8)
		return m, nil

	case dataLoadedMsg:
		m.ignoredFiles = msg.ignoredFiles
		m.loading = false

		// Create list items
		var items []list.Item
		for _, file := range m.ignoredFiles {
			items = append(items, ignoredItem{file: file})
		}

		m.list.SetItems(items)
		return m, nil

	case errMsg:
		m.err = msg.err
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			// Open selected file/directory in default editor/file manager
			if selected := m.list.SelectedItem(); selected != nil {
				if item, ok := selected.(ignoredItem); ok {
					go openInDefaultApp(item.file.Path)
				}
			}
			return m, nil
		case "e":
			// Export to file
			if m.options.OutputPath != "" {
				return m, func() tea.Msg {
					err := exportIgnoredFiles(m.ignoredFiles, m.options.OutputPath)
					if err != nil {
						return errMsg{err}
					}
					return nil
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
		return "Loading ignored files..."
	}

	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}

	var sections []string

	// Title
	title := titleStyle.Render("🚫 Git Ignored Files")
	sections = append(sections, title)

	// Summary
	summary := fmt.Sprintf("Total ignored files: %d", len(m.ignoredFiles))
	sections = append(sections, summary)

	// File list
	sections = append(sections, m.list.View())

	// Help
	helpText := "↑/↓: navigate • enter: open • e: export • q: quit"
	if m.options.OutputPath != "" {
		helpText = "↑/↓: navigate • enter: open • e: export to " + m.options.OutputPath + " • q: quit"
	}
	help := helpStyle.Render(helpText)
	sections = append(sections, help)

	return strings.Join(sections, "\n")
}

// RunIgnoredFiles starts the ignored files TUI
func RunIgnoredFiles(opts IgnoredOptions) error {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#874BFD")).
		BorderLeftForeground(lipgloss.Color("#874BFD"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#DDDDDD"))

	ignoredList := list.New([]list.Item{}, delegate, 0, 0)
	ignoredList.Title = "Ignored Files"
	ignoredList.SetShowStatusBar(false)
	ignoredList.SetShowHelp(false)

	m := model{
		list:      ignoredList,
		loading:   true,
		options:   opts,
		tuiHelper: terminal.NewResponsiveTUIHelper(),
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// Helper functions
func loadIgnoredData(opts IgnoredOptions) tea.Cmd {
	return func() tea.Msg {
		ignoredFiles, err := gatherIgnoredFiles(opts)
		if err != nil {
			return errMsg{err}
		}
		return dataLoadedMsg{ignoredFiles: ignoredFiles}
	}
}

func gatherIgnoredFiles(opts IgnoredOptions) ([]IgnoredFile, error) {
	// Get ignored files using git ls-files
	args := []string{"ls-files", "--ignored", "--exclude-standard", "--others"}
	if opts.ShowAll {
		args = append(args, "--directory")
	}

	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run git ls-files: %w", err)
	}

	var ignoredFiles []IgnoredFile
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		file := IgnoredFile{
			Path: line,
		}

		// Get file info if it exists
		if info, err := os.Stat(line); err == nil {
			file.IsDir = info.IsDir()
			if opts.ShowSizes && !file.IsDir {
				file.Size = info.Size()
			}
			file.ModTime = info.ModTime().Format("2006-01-02 15:04")
		} else {
			// File might not exist (deleted but still in index)
			file.ModTime = "unknown"
		}

		ignoredFiles = append(ignoredFiles, file)
	}

	// Sort by path
	sort.Slice(ignoredFiles, func(i, j int) bool {
		return ignoredFiles[i].Path < ignoredFiles[j].Path
	})

	return ignoredFiles, nil
}

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

func openInDefaultApp(path string) error {
	var cmd *exec.Cmd

	// Check if it's a directory or file
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	switch {
	case info.IsDir():
		// Open directory in file manager
		if cmd = getFileManagerCommand(path); cmd == nil {
			return fmt.Errorf("no file manager available")
		}
	default:
		// Open file in default editor
		if cmd = getEditorCommand(path); cmd == nil {
			return fmt.Errorf("no editor available")
		}
	}

	return cmd.Start()
}

func getFileManagerCommand(path string) *exec.Cmd {
	switch {
	case isCommandAvailable("explorer"):
		return exec.Command("explorer", path)
	case isCommandAvailable("xdg-open"):
		return exec.Command("xdg-open", path)
	case isCommandAvailable("open"):
		return exec.Command("open", path)
	default:
		return nil
	}
}

func getEditorCommand(path string) *exec.Cmd {
	editors := []string{"code", "notepad", "nano", "vi"}
	for _, editor := range editors {
		if isCommandAvailable(editor) {
			return exec.Command(editor, path)
		}
	}
	return nil
}

func isCommandAvailable(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func exportIgnoredFiles(files []IgnoredFile, outputPath string) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	// Write header
	fmt.Fprintf(f, "# Git Ignored Files\n")
	fmt.Fprintf(f, "# Generated on %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(f, "# Total files: %d\n\n", len(files))

	// Write file list
	for _, file := range files {
		sizeInfo := ""
		if file.Size > 0 && !file.IsDir {
			sizeInfo = fmt.Sprintf(" (%s)", formatSize(file.Size))
		}
		typeInfo := ""
		if file.IsDir {
			typeInfo = " [DIR]"
		}
		fmt.Fprintf(f, "%s%s%s - %s\n", file.Path, typeInfo, sizeInfo, file.ModTime)
	}

	return nil
}
