package scan

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/redjax/syst/internal/services/pathScanService/tbl"
	"github.com/redjax/syst/internal/utils/terminal"
)

// FileItem represents a file or directory with metadata
type FileItem struct {
	Name        string
	Path        string
	Size        int64
	SizeParsed  string
	Created     string
	Modified    string
	Owner       string
	Permissions string
	IsDir       bool
}

// ScanOptions holds the scan configuration
type ScanOptions struct {
	Path      string
	Limit     int
	SortBy    string
	Order     string
	Filter    string
	Recursive bool
}

// TUI Model and related types
type viewState int

const (
	fileListView viewState = iota
	sortDialogView
)

type model struct {
	files       []FileItem
	list        list.Model
	tuiHelper   *terminal.ResponsiveTUIHelper
	err         error
	loading     bool
	options     ScanOptions
	sortBy      string
	state       viewState
	sortDialog  sortDialogModel
	currentPath string
	pathHistory []string
}

type sortDialogModel struct {
	options   []string
	cursor    int
	selected  string
	orderOpts []string
	orderIdx  int
}

type fileItem struct {
	file FileItem
}

func (i fileItem) FilterValue() string { return i.file.Name }
func (i fileItem) Title() string {
	icon := "📄"
	if i.file.IsDir {
		icon = "📁"
	}
	return fmt.Sprintf("%s %s", icon, i.file.Name)
}
func (i fileItem) Description() string {
	desc := fmt.Sprintf("%s • %s • %s", i.file.SizeParsed, i.file.Modified, i.file.Permissions)
	if i.file.IsDir {
		desc += " • [enter to explore]"
	}
	return desc
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
			BorderForeground(lipgloss.Color("#25A065")).
			Padding(1, 2).
			Margin(1, 0)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)

	pathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFA500")).
			Bold(true)

	dialogStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#25A065")).
			Background(lipgloss.Color("#1A1A1A")).
			Padding(1, 2).
			Width(50)

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#25A065")).
			Foreground(lipgloss.Color("#FFFDF5")).
			Padding(0, 1)

	unselectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DDDDDD")).
			Padding(0, 1)
)

// TUI Messages
type dataLoadedMsg struct {
	files []FileItem
}

type errMsg struct {
	err error
}

type sortSelectedMsg struct {
	sortBy string
	order  string
}

type navigateToPathMsg struct {
	path string
}

// Model methods
func (m model) Init() tea.Cmd {
	return loadScanData(m.options)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.tuiHelper.HandleWindowSizeMsg(msg)
		width, height := m.tuiHelper.GetSize()
		m.list.SetWidth(width)
		m.list.SetHeight(height - 12) // Leave space for title, summary, and help
		return m, nil

	case dataLoadedMsg:
		m.files = msg.files
		m.loading = false

		// Create list items
		var items []list.Item
		for _, file := range m.files {
			items = append(items, fileItem{file: file})
		}

		m.list.SetItems(items)
		return m, nil

	case errMsg:
		m.err = msg.err
		m.loading = false
		return m, nil

	case sortSelectedMsg:
		m.sortBy = msg.sortBy
		m.options.Order = msg.order
		m.state = fileListView

		// Re-sort the files
		sortFiles(m.files, m.sortBy, m.options.Order)

		// Update list items
		var items []list.Item
		for _, file := range m.files {
			items = append(items, fileItem{file: file})
		}
		m.list.SetItems(items)
		return m, nil

	case navigateToPathMsg:
		m.currentPath = msg.path
		m.options.Path = msg.path
		m.loading = true
		m.list.SetItems([]list.Item{})
		return m, loadScanData(m.options)

	case tea.KeyMsg:
		if m.state == sortDialogView {
			return m.updateSortDialog(msg)
		}

		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))):
			return m, tea.Quit
		case key.Matches(msg, key.NewBinding(key.WithKeys("backspace", "u"))):
			// Go back to parent directory
			if len(m.pathHistory) > 0 {
				// Pop the last path from history
				parentPath := m.pathHistory[len(m.pathHistory)-1]
				m.pathHistory = m.pathHistory[:len(m.pathHistory)-1]
				return m, func() tea.Msg {
					return navigateToPathMsg{path: parentPath}
				}
			} else {
				// Try to go up one directory level
				parentPath := filepath.Dir(m.currentPath)
				if parentPath != m.currentPath { // Make sure we're not at root
					return m, func() tea.Msg {
						return navigateToPathMsg{path: parentPath}
					}
				}
			}
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			// Navigate into directory or open file
			if selected := m.list.SelectedItem(); selected != nil {
				if item, ok := selected.(fileItem); ok {
					if item.file.IsDir {
						// Navigate into directory
						m.pathHistory = append(m.pathHistory, m.currentPath)
						return m, func() tea.Msg {
							return navigateToPathMsg{path: item.file.Path}
						}
					} else {
						// Open file
						go openInDefaultApp(item.file.Path)
					}
				}
			}
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("o"))):
			// Always open in external file manager/application
			if selected := m.list.SelectedItem(); selected != nil {
				if item, ok := selected.(fileItem); ok {
					go openInDefaultApp(item.file.Path)
				}
			}
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("s"))):
			// Show sort dialog
			m.state = sortDialogView
			m.sortDialog = newSortDialog(m.sortBy, m.options.Order)
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
			// Toggle sort order
			if m.options.Order == "asc" {
				m.options.Order = "desc"
			} else {
				m.options.Order = "asc"
			}

			// Re-sort the files
			sortFiles(m.files, m.sortBy, m.options.Order)

			// Update list items
			var items []list.Item
			for _, file := range m.files {
				items = append(items, fileItem{file: file})
			}
			m.list.SetItems(items)
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.loading {
		return "Scanning directory..."
	}

	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}

	if m.state == sortDialogView {
		return m.renderSortDialog()
	}

	var sections []string

	// Title
	title := titleStyle.Render("📂 Directory Scanner")
	sections = append(sections, title)

	// Summary in styled section
	summaryContent := fmt.Sprintf("Path: %s\nTotal items: %d",
		pathStyle.Render(m.currentPath),
		len(m.files))

	if len(m.files) > 0 {
		// Add breakdown by type
		dirs := 0
		files := 0
		var totalSize int64
		for _, file := range m.files {
			if file.IsDir {
				dirs++
			} else {
				files++
				totalSize += file.Size
			}
		}
		summaryContent += fmt.Sprintf("\nDirectories: %d  Files: %d", dirs, files)
		if totalSize > 0 {
			summaryContent += fmt.Sprintf("  Total size: %s", tbl.ByteCountIEC(totalSize))
		}
		summaryContent += fmt.Sprintf("\nSorted by: %s (%s)", m.sortBy, m.options.Order)
	}

	// Add navigation breadcrumb if we have history
	if len(m.pathHistory) > 0 {
		// Show just the immediate parent in breadcrumb to save space
		parent := filepath.Base(m.pathHistory[len(m.pathHistory)-1])
		current := filepath.Base(m.currentPath)
		summaryContent += fmt.Sprintf("\nNavigation: %s → %s (backspace to go back)", parent, current)
	} else if filepath.Dir(m.currentPath) != m.currentPath {
		// Show that we can go up one level
		parent := filepath.Base(filepath.Dir(m.currentPath))
		current := filepath.Base(m.currentPath)
		summaryContent += fmt.Sprintf("\nNavigation: %s → %s (backspace to go up)", parent, current)
	}

	summary := sectionStyle.Render(summaryContent)
	sections = append(sections, summary)

	// File list
	sections = append(sections, m.list.View())

	// Help
	helpText := "↑/↓: navigate • enter: open file/explore dir • o: open in explorer • backspace/u: go back • s: sort • r: toggle order • q: quit"
	help := helpStyle.Render(helpText)
	sections = append(sections, help)

	return strings.Join(sections, "\n")
}

// Helper methods
func (m model) cycleSortBy() tea.Cmd {
	sortOptions := []string{"name", "size", "created", "modified", "owner", "permissions"}
	currentIndex := 0
	for i, option := range sortOptions {
		if option == m.sortBy {
			currentIndex = i
			break
		}
	}

	nextIndex := (currentIndex + 1) % len(sortOptions)
	m.sortBy = sortOptions[nextIndex]

	// Re-sort the files
	sortFiles(m.files, m.sortBy, m.options.Order)

	// Update list items
	var items []list.Item
	for _, file := range m.files {
		items = append(items, fileItem{file: file})
	}

	return func() tea.Msg {
		return dataLoadedMsg{files: m.files}
	}
}

func (m model) reverseSortOrder() tea.Cmd {
	if m.options.Order == "asc" {
		m.options.Order = "desc"
	} else {
		m.options.Order = "asc"
	}

	// Re-sort the files
	sortFiles(m.files, m.sortBy, m.options.Order)

	// Update list items
	var items []list.Item
	for _, file := range m.files {
		items = append(items, fileItem{file: file})
	}

	return func() tea.Msg {
		return dataLoadedMsg{files: m.files}
	}
}

// Sort dialog methods
func newSortDialog(currentSort, currentOrder string) sortDialogModel {
	options := []string{"name", "size", "created", "modified", "owner", "permissions"}
	orderOpts := []string{"asc", "desc"}

	cursor := 0
	for i, opt := range options {
		if opt == currentSort {
			cursor = i
			break
		}
	}

	orderIdx := 0
	if currentOrder == "desc" {
		orderIdx = 1
	}

	return sortDialogModel{
		options:   options,
		cursor:    cursor,
		selected:  currentSort,
		orderOpts: orderOpts,
		orderIdx:  orderIdx,
	}
}

func (m model) updateSortDialog(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("escape", "q"))):
		m.state = fileListView
		return m, nil
	case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		selectedSort := m.sortDialog.options[m.sortDialog.cursor]
		selectedOrder := m.sortDialog.orderOpts[m.sortDialog.orderIdx]
		return m, func() tea.Msg {
			return sortSelectedMsg{
				sortBy: selectedSort,
				order:  selectedOrder,
			}
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
		if m.sortDialog.cursor > 0 {
			m.sortDialog.cursor--
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
		if m.sortDialog.cursor < len(m.sortDialog.options)-1 {
			m.sortDialog.cursor++
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("tab", "right", "l"))):
		m.sortDialog.orderIdx = (m.sortDialog.orderIdx + 1) % len(m.sortDialog.orderOpts)
	case key.Matches(msg, key.NewBinding(key.WithKeys("shift+tab", "left", "h"))):
		m.sortDialog.orderIdx = (m.sortDialog.orderIdx - 1 + len(m.sortDialog.orderOpts)) % len(m.sortDialog.orderOpts)
	}

	return m, nil
}

func (m model) renderSortDialog() string {
	var content strings.Builder

	content.WriteString("Sort Options\n\n")
	content.WriteString("Sort by:\n")

	for i, option := range m.sortDialog.options {
		style := unselectedStyle
		if i == m.sortDialog.cursor {
			style = selectedStyle
		}
		content.WriteString(style.Render(fmt.Sprintf("  %s", option)) + "\n")
	}

	content.WriteString("\nOrder:\n")
	for i, order := range m.sortDialog.orderOpts {
		style := unselectedStyle
		if i == m.sortDialog.orderIdx {
			style = selectedStyle
		}
		content.WriteString(style.Render(fmt.Sprintf("  %s", order)) + "\n")
	}

	content.WriteString("\n")
	content.WriteString(helpStyle.Render("↑/↓: navigate sort options • tab/←/→: toggle order • enter: apply • esc/q: cancel"))

	dialog := dialogStyle.Render(content.String())

	// Center the dialog
	termWidth, termHeight := m.tuiHelper.GetSize()
	return lipgloss.Place(termWidth, termHeight, lipgloss.Center, lipgloss.Center, dialog)
}

// ScanDirectoryTUI starts the directory scanner TUI
func ScanDirectoryTUI(path string, limit int, sortBy, order, filter string, recursive bool) error {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#25A065")).
		BorderLeftForeground(lipgloss.Color("#25A065"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#DDDDDD"))

	fileList := list.New([]list.Item{}, delegate, 0, 0)
	fileList.Title = "Files"
	fileList.SetShowStatusBar(false)
	fileList.SetShowHelp(false)

	options := ScanOptions{
		Path:      path,
		Limit:     limit,
		SortBy:    sortBy,
		Order:     order,
		Filter:    filter,
		Recursive: recursive,
	}

	m := model{
		list:        fileList,
		loading:     true,
		options:     options,
		sortBy:      sortBy,
		state:       fileListView,
		currentPath: path,
		pathHistory: []string{},
		tuiHelper:   terminal.NewResponsiveTUIHelper(),
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// Helper functions
func loadScanData(opts ScanOptions) tea.Cmd {
	return func() tea.Msg {
		files, err := gatherFileData(opts)
		if err != nil {
			return errMsg{err}
		}
		return dataLoadedMsg{files: files}
	}
}

func gatherFileData(opts ScanOptions) ([]FileItem, error) {
	var files []FileItem
	count := 0

	if opts.Recursive {
		err := filepath.Walk(opts.Path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip files we can't access
			}

			// Skip .git directories
			if info.IsDir() && info.Name() == ".git" {
				return filepath.SkipDir
			}

			// Check limit
			if opts.Limit > 0 && count >= opts.Limit {
				return fmt.Errorf("limit_reached")
			}

			file := createFileItem(info, path, opts.Path)
			files = append(files, file)
			count++

			return nil
		})

		if err != nil && err.Error() != "limit_reached" {
			return nil, err
		}
	} else {
		entries, err := os.ReadDir(opts.Path)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			file := createFileItem(info, filepath.Join(opts.Path, entry.Name()), opts.Path)
			files = append(files, file)
			count++

			if opts.Limit > 0 && count >= opts.Limit {
				break
			}
		}
	}

	// Apply filter if specified
	if opts.Filter != "" {
		files = applyFilter(files, opts.Filter)
	}

	// Sort files
	sortFiles(files, opts.SortBy, opts.Order)

	return files, nil
}

func createFileItem(info os.FileInfo, fullPath, rootPath string) FileItem {
	ctime, owner := getMeta(info, fullPath)
	size := info.Size()

	// Calculate directory size if it's a directory
	if info.IsDir() {
		size = calculateDirectorySize(fullPath)
	}

	sizeParsed := tbl.ByteCountIEC(size)

	return FileItem{
		Name:        info.Name(),
		Path:        fullPath,
		Size:        size,
		SizeParsed:  sizeParsed,
		Created:     ctime,
		Modified:    info.ModTime().Format("2006-01-02 15:04:05"),
		Owner:       owner,
		Permissions: info.Mode().String(),
		IsDir:       info.IsDir(),
	}
}

func sortFiles(files []FileItem, sortBy, order string) {
	sort.Slice(files, func(i, j int) bool {
		var result bool

		switch sortBy {
		case "name":
			result = files[i].Name < files[j].Name
		case "size":
			result = files[i].Size < files[j].Size
		case "created":
			result = files[i].Created < files[j].Created
		case "modified":
			result = files[i].Modified < files[j].Modified
		case "owner":
			result = files[i].Owner < files[j].Owner
		case "permissions":
			result = files[i].Permissions < files[j].Permissions
		default:
			result = files[i].Name < files[j].Name
		}

		if order == "desc" {
			return !result
		}
		return result
	})
}

func applyFilter(files []FileItem, filterString string) []FileItem {
	// Simple filtering implementation
	// This could be enhanced to match the existing tbl.ParseFilter functionality
	var filtered []FileItem

	for _, file := range files {
		// For now, simple name filtering
		if strings.Contains(strings.ToLower(file.Name), strings.ToLower(filterString)) {
			filtered = append(filtered, file)
		}
	}

	return filtered
}

func openInDefaultApp(path string) error {
	// Same implementation as in the ignored service
	var cmd *exec.Cmd

	switch {
	case isCommandAvailable("explorer"):
		cmd = exec.Command("explorer", path)
	case isCommandAvailable("xdg-open"):
		cmd = exec.Command("xdg-open", path)
	case isCommandAvailable("open"):
		cmd = exec.Command("open", path)
	default:
		return fmt.Errorf("no file manager available")
	}

	return cmd.Start()
}

func isCommandAvailable(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func calculateDirectorySize(dirPath string) int64 {
	var size int64

	// Use filepath.Walk to traverse directory and sum file sizes
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip files we can't access
			return nil
		}

		// Skip .git directories to avoid large sizes
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		// Only count regular files, not directories
		if !info.IsDir() {
			size += info.Size()
		}

		return nil
	})

	if err != nil {
		// If we can't walk the directory, return 0
		return 0
	}

	return size
}
