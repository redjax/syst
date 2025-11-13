package worktreeservice

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/redjax/syst/internal/utils/terminal"
)

type viewMode int

const (
	listView viewMode = iota
	formView
	confirmView
)

type model struct {
	worktrees       []Worktree
	manager         *WorktreeManager
	cursor          int
	err             error
	currentView     viewMode
	formInputs      []textinput.Model
	focusedInput    int
	confirmAction   string
	confirmTarget   string
	tuiHelper       *terminal.ResponsiveTUIHelper
	message         string
	createNewBranch bool
}

type worktreesLoadedMsg struct {
	worktrees []Worktree
}

type errMsg struct {
	err error
}

type successMsg struct {
	message string
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#874BFD")).
			Padding(0, 1)

	normalStyle = lipgloss.NewStyle().
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Padding(1, 0)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87")).
			Padding(1, 2)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Padding(1, 2)

	formStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 2).
			MarginTop(1)
)

func initialModel(manager *WorktreeManager) model {
	tuiHelper := terminal.NewResponsiveTUIHelper()

	return model{
		manager:     manager,
		currentView: listView,
		tuiHelper:   tuiHelper,
	}
}

func (m model) Init() tea.Cmd {
	return loadWorktrees(m.manager)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.tuiHelper.HandleWindowSizeMsg(msg)
		return m, nil

	case worktreesLoadedMsg:
		m.worktrees = msg.worktrees
		m.message = ""
		return m, nil

	case successMsg:
		m.message = msg.message
		m.currentView = listView
		m.err = nil
		// Only reload worktrees if the message doesn't say "Opened"
		if !strings.Contains(msg.message, "Opened") {
			return m, loadWorktrees(m.manager)
		}
		return m, nil

	case errMsg:
		m.err = msg.err
		m.message = ""
		m.currentView = listView
		return m, nil

	case tea.KeyMsg:
		// If we're in form view, handle form input updates first
		if m.currentView == formView {
			// Let text inputs handle the message first
			m2, cmd := m.updateFormInputs(msg)
			m = m2.(model)
			// Then handle form navigation keys
			m3, cmd2 := m.handleFormViewKeys(msg)
			return m3, tea.Batch(cmd, cmd2)
		}
		return m.handleKeyPress(msg)
	}

	return m, nil
}

func (m model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case listView:
		return m.handleListViewKeys(msg)
	case formView:
		return m.handleFormViewKeys(msg)
	case confirmView:
		return m.handleConfirmViewKeys(msg)
	}
	return m, nil
}

func (m model) handleListViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))):
		return m, tea.Quit
	case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
		if m.cursor < len(m.worktrees)-1 {
			m.cursor++
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("n"))):
		m.currentView = formView
		m.formInputs = m.createAddForm()
		m.focusedInput = 0
		if len(m.formInputs) > 0 {
			m.formInputs[0].Focus()
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("d"))):
		if len(m.worktrees) > 0 && m.cursor < len(m.worktrees) {
			m.currentView = confirmView
			m.confirmAction = "delete"
			m.confirmTarget = m.worktrees[m.cursor].Path
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("o"))):
		if len(m.worktrees) > 0 && m.cursor < len(m.worktrees) {
			return m, openWorktree(m.worktrees[m.cursor].Path)
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
		return m, loadWorktrees(m.manager)
	}
	return m, nil
}

func (m model) handleFormViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc":
		m.currentView = listView
		m.formInputs = nil
		return m, nil
	case "tab", "shift+tab":
		if msg.String() == "tab" {
			m.focusedInput++
		} else {
			m.focusedInput--
		}
		if m.focusedInput >= len(m.formInputs) {
			m.focusedInput = 0
		}
		if m.focusedInput < 0 {
			m.focusedInput = len(m.formInputs) - 1
		}
		for i := range m.formInputs {
			if i == m.focusedInput {
				m.formInputs[i].Focus()
			} else {
				m.formInputs[i].Blur()
			}
		}
		return m, nil
	case "enter":
		return m, m.submitAddForm()
	case "ctrl+n":
		m.createNewBranch = !m.createNewBranch
		return m, nil
	}
	return m, nil
}

func (m model) handleConfirmViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if m.confirmAction == "delete" {
			return m, deleteWorktree(m.manager, m.confirmTarget)
		}
	case "n", "N", "esc", "ctrl+c":
		m.currentView = listView
		m.confirmAction = ""
		m.confirmTarget = ""
	}
	return m, nil
}

func (m model) updateFormInputs(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	for i := range m.formInputs {
		var cmd tea.Cmd
		m.formInputs[i], cmd = m.formInputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m model) createAddForm() []textinput.Model {
	inputs := make([]textinput.Model, 2)

	// Path input
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Path to new worktree"
	inputs[0].CharLimit = 256
	inputs[0].Width = 50
	inputs[0].Prompt = "Path: "

	// Branch input
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Branch name (optional)"
	inputs[1].CharLimit = 128
	inputs[1].Width = 50
	inputs[1].Prompt = "Branch: "

	return inputs
}

func (m model) View() string {
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v\n\nPress 'q' to quit", m.err))
	}

	switch m.currentView {
	case listView:
		return m.renderListView()
	case formView:
		return m.renderFormView()
	case confirmView:
		return m.renderConfirmView()
	}

	return ""
}

func (m model) renderListView() string {
	var s strings.Builder

	// Title
	title := titleStyle.Render("Git Worktrees")
	s.WriteString(title + "\n\n")

	// Message
	if m.message != "" {
		s.WriteString(successStyle.Render(m.message) + "\n\n")
	}

	// Worktree list
	if len(m.worktrees) == 0 {
		s.WriteString("  No worktrees found\n\n")
	} else {
		for i, wt := range m.worktrees {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}

			branch := wt.Branch
			if branch == "" {
				branch = "(detached)"
			}

			line := fmt.Sprintf("%s %s [%s]", cursor, wt.Path, branch)

			if m.cursor == i {
				s.WriteString(selectedStyle.Render(line) + "\n")
			} else {
				s.WriteString(normalStyle.Render(line) + "\n")
			}
		}
	}

	// Help
	s.WriteString("\n")
	s.WriteString(helpStyle.Render("(n) new worktree  (d) delete  (o) open  (r) refresh  (q) quit"))
	s.WriteString("\n")
	s.WriteString(helpStyle.Render("(↑/k) up  (↓/j) down"))

	return s.String()
}

func (m model) renderFormView() string {
	var s strings.Builder

	title := titleStyle.Render("New Worktree")
	s.WriteString(title + "\n\n")

	for i, input := range m.formInputs {
		s.WriteString(input.View() + "\n")
		if i < len(m.formInputs)-1 {
			s.WriteString("\n")
		}
	}

	s.WriteString("\n\n")

	checkBox := "[ ]"
	if m.createNewBranch {
		checkBox = "[x]"
	}
	s.WriteString(fmt.Sprintf("%s Force new branch creation (Ctrl+N to toggle)\n", checkBox))
	s.WriteString(helpStyle.Render("   (branches are auto-created if they don't exist)\n\n"))

	s.WriteString(helpStyle.Render("(Tab) next field  (Enter) create  (Esc) cancel"))

	return formStyle.Render(s.String())
}

func (m model) renderConfirmView() string {
	var s strings.Builder

	title := titleStyle.Render("Confirm")
	s.WriteString(title + "\n\n")

	s.WriteString(fmt.Sprintf("Are you sure you want to delete worktree:\n%s\n\n", m.confirmTarget))
	s.WriteString(helpStyle.Render("(y) yes  (n) no"))

	return formStyle.Render(s.String())
}

func (m model) submitAddForm() tea.Cmd {
	path := strings.TrimSpace(m.formInputs[0].Value())
	branch := strings.TrimSpace(m.formInputs[1].Value())

	if path == "" {
		return func() tea.Msg {
			return errMsg{err: fmt.Errorf("path cannot be empty")}
		}
	}

	// Generate default path if only branch is provided
	if branch != "" && path == branch {
		path = m.manager.GenerateWorktreePath(branch)
	}

	// Auto-detect if we need to create a new branch
	createNewBranch := m.createNewBranch
	if branch != "" && !createNewBranch {
		// Check if branch exists
		exists, err := m.manager.BranchExists(branch)
		if err != nil {
			return func() tea.Msg {
				return errMsg{err: fmt.Errorf("failed to check branch: %w", err)}
			}
		}
		// If branch doesn't exist, automatically create it
		if !exists {
			createNewBranch = true
		}
	}

	opts := AddWorktreeOptions{
		Path:      path,
		Branch:    branch,
		NewBranch: createNewBranch,
		Checkout:  true,
	}

	return addWorktree(m.manager, opts)
}

// Commands
func loadWorktrees(manager *WorktreeManager) tea.Cmd {
	return func() tea.Msg {
		worktrees, err := manager.ListWorktrees()
		if err != nil {
			return errMsg{err: err}
		}
		return worktreesLoadedMsg{worktrees: worktrees}
	}
}

func addWorktree(manager *WorktreeManager, opts AddWorktreeOptions) tea.Cmd {
	return func() tea.Msg {
		if err := manager.AddWorktree(opts); err != nil {
			return errMsg{err: err}
		}
		return successMsg{message: fmt.Sprintf("Created worktree at %s", opts.Path)}
	}
}

func deleteWorktree(manager *WorktreeManager, path string) tea.Cmd {
	return func() tea.Msg {
		if err := manager.RemoveWorktree(path, false); err != nil {
			return errMsg{err: err}
		}
		return successMsg{message: fmt.Sprintf("Deleted worktree %s", path)}
	}
}

func openWorktree(path string) tea.Cmd {
	return func() tea.Msg {
		if err := OpenInEditor(path); err != nil {
			return errMsg{err: err}
		}
		return successMsg{message: fmt.Sprintf("Opened %s in editor", path)}
	}
}

// RunWorktreeTUI starts the worktree TUI
func RunWorktreeTUI(repoPath string) error {
	manager, err := NewWorktreeManager(repoPath)
	if err != nil {
		return err
	}

	p := tea.NewProgram(initialModel(manager), tea.WithAltScreen())
	_, err = p.Run()
	return err
}
