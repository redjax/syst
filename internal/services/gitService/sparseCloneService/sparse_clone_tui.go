package sparsecloneservice

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type viewState int

const (
	formView viewState = iota
	confirmationView
)

type inputField int

const (
	providerInput inputField = iota
	protocolInput
	userInput
	repositoryInput
	outputInput
	branchInput
	pathsInput
	confirmInput
)

type model struct {
	inputs         []textinput.Model
	focused        inputField
	err            error
	submitted      bool
	pathsList      []string
	pathCursor     int
	pathEditMode   bool
	terminalWidth  int
	terminalHeight int
	options        SparseCloneOptions
	currentView    viewState
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")).
			MarginBottom(1)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true)

	pathItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	selectedPathStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("196")).
				Bold(true)
)

func NewSparseCloneTUI() model {
	inputs := make([]textinput.Model, 8)

	// Provider input
	inputs[providerInput] = textinput.New()
	inputs[providerInput].Placeholder = "github"
	inputs[providerInput].SetValue("github")
	inputs[providerInput].CharLimit = 20
	inputs[providerInput].Width = 30

	// Protocol input
	inputs[protocolInput] = textinput.New()
	inputs[protocolInput].Placeholder = "ssh"
	inputs[protocolInput].SetValue("ssh")
	inputs[protocolInput].CharLimit = 10
	inputs[protocolInput].Width = 30

	// Username input
	inputs[userInput] = textinput.New()
	inputs[userInput].Placeholder = "username or organization"
	inputs[userInput].CharLimit = 50
	inputs[userInput].Width = 30

	// Repository input
	inputs[repositoryInput] = textinput.New()
	inputs[repositoryInput].Placeholder = "repository name"
	inputs[repositoryInput].CharLimit = 100
	inputs[repositoryInput].Width = 30

	// Output directory input
	inputs[outputInput] = textinput.New()
	inputs[outputInput].Placeholder = "output directory (optional)"
	inputs[outputInput].CharLimit = 100
	inputs[outputInput].Width = 30

	// Branch input
	inputs[branchInput] = textinput.New()
	inputs[branchInput].Placeholder = "main"
	inputs[branchInput].SetValue("main")
	inputs[branchInput].CharLimit = 50
	inputs[branchInput].Width = 30

	// Paths input
	inputs[pathsInput] = textinput.New()
	inputs[pathsInput].Placeholder = "path to checkout (press Enter to add)"
	inputs[pathsInput].CharLimit = 200
	inputs[pathsInput].Width = 50

	// Confirm input
	inputs[confirmInput] = textinput.New()
	inputs[confirmInput].Placeholder = "y/N"
	inputs[confirmInput].CharLimit = 1
	inputs[confirmInput].Width = 10

	// Focus first input
	inputs[providerInput].Focus()

	return model{
		inputs:         inputs,
		focused:        providerInput,
		pathsList:      []string{},
		pathCursor:     0,
		pathEditMode:   false,
		terminalWidth:  80,
		terminalHeight: 24,
		currentView:    formView,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width
		m.terminalHeight = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "tab", "down":
			if m.currentView == confirmationView {
				// Move down in paths list in confirmation view
				if m.pathCursor < len(m.pathsList)-1 {
					m.pathCursor++
				}
			} else if m.pathEditMode {
				// Move down in paths list
				if m.pathCursor < len(m.pathsList)-1 {
					m.pathCursor++
				}
			} else {
				m = m.nextInput()
			}
			return m, nil

		case "shift+tab", "up":
			if m.currentView == confirmationView {
				// Move up in paths list in confirmation view
				if m.pathCursor > 0 {
					m.pathCursor--
				}
			} else if m.pathEditMode {
				// Move up in paths list
				if m.pathCursor > 0 {
					m.pathCursor--
				}
			} else {
				m = m.prevInput()
			}
			return m, nil

		case "p":
			// Toggle path edit mode when in pathsInput, but only if input is empty
			if m.focused == pathsInput && len(m.pathsList) > 0 && strings.TrimSpace(m.inputs[pathsInput].Value()) == "" {
				m.pathEditMode = !m.pathEditMode
				if m.pathEditMode && m.pathCursor >= len(m.pathsList) {
					m.pathCursor = len(m.pathsList) - 1
				}
				return m, nil
			}

		case "enter":
			// Handle different behaviors based on current view and field
			if m.currentView == confirmationView {
				// In confirmation view, Enter submits the form
				m.submitted = true
				m.buildOptions()
				return m, tea.Quit
			}

			// Handle form view actions
			switch m.focused {
			case pathsInput:
				if m.pathEditMode {
					// Exit path edit mode
					m.pathEditMode = false
				} else {
					// Add path to list
					path := strings.TrimSpace(m.inputs[pathsInput].Value())
					if path != "" {
						m.pathsList = append(m.pathsList, path)
						m.inputs[pathsInput].SetValue("")
					}
				}
				return m, nil

			case confirmInput:
				// Transition to confirmation view
				m.currentView = confirmationView
				m.pathCursor = 0 // Reset cursor for confirmation view
				return m, nil

			default:
				// Move to next input for other fields
				if m.focused < confirmInput {
					m = m.nextInput()
				}
				return m, nil
			}

		case "d":
			// Delete path when in path edit mode OR in confirmation view
			if (m.pathEditMode || m.currentView == confirmationView) && len(m.pathsList) > 0 && m.pathCursor < len(m.pathsList) {
				m.pathsList = append(m.pathsList[:m.pathCursor], m.pathsList[m.pathCursor+1:]...)
				if m.pathCursor >= len(m.pathsList) && len(m.pathsList) > 0 {
					m.pathCursor = len(m.pathsList) - 1
				}
				if len(m.pathsList) == 0 {
					m.pathEditMode = false
				}
				return m, nil
			}

		case "backspace", "delete":
			// Handle different actions based on current view
			if m.currentView == confirmationView {
				// In confirmation view, backspace goes back to form
				m.currentView = formView
				return m, nil
			}

			// Allow removing paths when focused on path input and there are paths
			if m.focused == pathsInput && len(m.pathsList) > 0 && m.inputs[pathsInput].Value() == "" && !m.pathEditMode {
				// Remove the last path
				m.pathsList = m.pathsList[:len(m.pathsList)-1]
				return m, nil
			}
		}
	}

	// Update the current input only if not in path edit mode
	if !m.pathEditMode {
		var cmd tea.Cmd
		m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	if m.submitted {
		return successStyle.Render("✓ Sparse clone configuration complete! Executing clone...\n")
	}

	switch m.currentView {
	case formView:
		return m.renderFormView()
	case confirmationView:
		return m.renderConfirmationView()
	default:
		return m.renderFormView()
	}
}

func (m model) renderFormView() string {
	var b strings.Builder

	// Title - ALWAYS visible, never scrolls
	b.WriteString(titleStyle.Render("🔧 Git Sparse Clone Configuration"))
	b.WriteString("\n\n")

	// Calculate available height for scrolling content
	// Reserve lines for: title (3), scroll indicators (2), help text (1), margins (2)
	availableHeight := m.terminalHeight - 8
	if availableHeight < 6 {
		availableHeight = 6
	}

	// Build all form content into a slice of lines
	allLines := []string{}

	// Provider (lines 0-2)
	allLines = append(allLines, labelStyle.Render("Git Provider:"))
	inputLine := m.inputs[providerInput].View()
	if m.focused == providerInput {
		inputLine += helpStyle.Render(" (github, gitlab, codeberg)")
	}
	allLines = append(allLines, inputLine)
	allLines = append(allLines, "") // spacing

	// Protocol (lines 3-5)
	allLines = append(allLines, labelStyle.Render("Protocol:"))
	inputLine = m.inputs[protocolInput].View()
	if m.focused == protocolInput {
		inputLine += helpStyle.Render(" (ssh, https)")
	}
	allLines = append(allLines, inputLine)
	allLines = append(allLines, "")

	// Username (lines 6-8)
	allLines = append(allLines, labelStyle.Render("Username/Organization:"))
	allLines = append(allLines, m.inputs[userInput].View())
	allLines = append(allLines, "")

	// Repository (lines 9-11)
	allLines = append(allLines, labelStyle.Render("Repository Name:"))
	allLines = append(allLines, m.inputs[repositoryInput].View())
	allLines = append(allLines, "")

	// Output Directory (lines 12-14)
	allLines = append(allLines, labelStyle.Render("Output Directory:"))
	inputLine = m.inputs[outputInput].View()
	if m.focused == outputInput {
		inputLine += helpStyle.Render(" (optional)")
	}
	allLines = append(allLines, inputLine)
	allLines = append(allLines, "")

	// Branch (lines 15-17)
	allLines = append(allLines, labelStyle.Render("Branch:"))
	allLines = append(allLines, m.inputs[branchInput].View())
	allLines = append(allLines, "")

	// Track where paths section starts for scroll calculation
	pathsSectionStart := len(allLines)

	// Paths section
	allLines = append(allLines, labelStyle.Render("Sparse Checkout Paths:"))
	if len(m.pathsList) > 0 {
		// Limit displayed paths to fit
		maxPathsToShow := availableHeight / 3
		if maxPathsToShow < 3 {
			maxPathsToShow = 3
		}

		startIdx := 0
		if len(m.pathsList) > maxPathsToShow {
			if m.pathEditMode {
				startIdx = m.pathCursor - maxPathsToShow/2
				if startIdx < 0 {
					startIdx = 0
				}
				if startIdx+maxPathsToShow > len(m.pathsList) {
					startIdx = len(m.pathsList) - maxPathsToShow
				}
			}
		}

		endIdx := startIdx + maxPathsToShow
		if endIdx > len(m.pathsList) {
			endIdx = len(m.pathsList)
		}

		if startIdx > 0 {
			allLines = append(allLines, helpStyle.Render(fmt.Sprintf("  ... (%d more above)", startIdx)))
		}

		for i := startIdx; i < endIdx; i++ {
			if m.pathEditMode && i == m.pathCursor {
				allLines = append(allLines, selectedPathStyle.Render(fmt.Sprintf("► %d. %s", i+1, m.pathsList[i])))
			} else {
				allLines = append(allLines, pathItemStyle.Render(fmt.Sprintf("  %d. %s", i+1, m.pathsList[i])))
			}
		}

		if endIdx < len(m.pathsList) {
			allLines = append(allLines, helpStyle.Render(fmt.Sprintf("  ... (%d more below)", len(m.pathsList)-endIdx)))
		}
		allLines = append(allLines, "")
	}

	inputLine = m.inputs[pathsInput].View()
	allLines = append(allLines, inputLine)

	if m.focused == pathsInput {
		var helpText string
		if len(m.pathsList) > 0 {
			if m.pathEditMode {
				helpText = "Path Edit Mode: ↑/↓: navigate • d: delete • Enter: exit edit"
			} else {
				currentInput := strings.TrimSpace(m.inputs[pathsInput].Value())
				if currentInput == "" {
					helpText = "Enter: add path • p: edit existing paths • Backspace: remove last"
				} else {
					helpText = "Enter: add path • Backspace: remove last"
				}
			}
		} else {
			helpText = "Enter: add path"
		}
		allLines = append(allLines, helpStyle.Render(helpText))
	}
	allLines = append(allLines, "")

	// Confirmation
	confirmSectionStart := len(allLines)
	if len(m.pathsList) > 0 {
		allLines = append(allLines, labelStyle.Render("Proceed with sparse clone? (y/N):"))
		allLines = append(allLines, m.inputs[confirmInput].View())
		allLines = append(allLines, "")
	}

	// Calculate which line the focused input is on
	focusedInputLine := 0
	switch m.focused {
	case providerInput:
		focusedInputLine = 1 // The input line, not the label
	case protocolInput:
		focusedInputLine = 4
	case userInput:
		focusedInputLine = 7
	case repositoryInput:
		focusedInputLine = 10
	case outputInput:
		focusedInputLine = 13
	case branchInput:
		focusedInputLine = 16
	case pathsInput:
		focusedInputLine = pathsSectionStart + 1 + len(m.pathsList)
		if len(m.pathsList) > 0 {
			focusedInputLine++ // Extra line for paths display separator
		}
	case confirmInput:
		focusedInputLine = confirmSectionStart + 1
	}

	// Calculate scroll offset to keep focused input visible
	// Keep the focused line in the middle third of the viewport
	scrollOffset := 0
	if focusedInputLine > availableHeight/3 {
		scrollOffset = focusedInputLine - availableHeight/3
	}

	// Don't scroll past the end
	maxScrollOffset := len(allLines) - availableHeight
	if maxScrollOffset < 0 {
		maxScrollOffset = 0
	}
	if scrollOffset > maxScrollOffset {
		scrollOffset = maxScrollOffset
	}

	// Adjust available height to account for scroll indicators
	// If we're scrolled, we'll show indicators which take up lines
	displayHeight := availableHeight
	if scrollOffset > 0 {
		displayHeight-- // "▲ More above" takes 1 line
	}
	if scrollOffset+availableHeight < len(allLines) {
		displayHeight-- // "▼ More below" takes 1 line
	}

	// Calculate visible range using adjusted height
	endLine := scrollOffset + displayHeight
	if endLine > len(allLines) {
		endLine = len(allLines)
	}

	// Add scroll indicator at top of content area if scrolled down
	if scrollOffset > 0 {
		b.WriteString(helpStyle.Render("▲ More above"))
		b.WriteString("\n")
	}

	// Render visible content lines
	visibleLines := allLines[scrollOffset:endLine]
	for _, line := range visibleLines {
		b.WriteString(line)
		b.WriteString("\n")
	}

	// Add scroll indicator at bottom if there's more content
	if endLine < len(allLines) {
		b.WriteString(helpStyle.Render("▼ More below"))
		b.WriteString("\n")
	}

	// Error message (if any)
	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n")
	}

	// Help text
	if m.pathEditMode {
		// Don't show additional help when in path edit mode
	} else if m.focused == pathsInput {
		// Don't show global help when focused on paths
		// (help is already shown above near the paths)
	} else {
		b.WriteString(helpStyle.Render("tab/↓: next • shift+tab/↑: previous • enter: confirm/add • esc: quit"))
	}

	return b.String()
}

func (m model) renderConfirmationView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("📋 Confirmation - Review Your Configuration"))
	b.WriteString("\n\n")

	provider := m.getFieldValue(providerInput, "github")
	protocol := m.getFieldValue(protocolInput, "ssh")
	user := m.getFieldValue(userInput, "")
	repo := m.getFieldValue(repositoryInput, "")
	output := m.getFieldValue(outputInput, repo)
	branch := m.getFieldValue(branchInput, "main")

	b.WriteString(labelStyle.Render("Configuration Summary:"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Provider: %s\n", provider))
	b.WriteString(fmt.Sprintf("  Protocol: %s\n", protocol))
	b.WriteString(fmt.Sprintf("  Repository: %s/%s\n", user, repo))
	b.WriteString(fmt.Sprintf("  Output Directory: %s\n", output))
	b.WriteString(fmt.Sprintf("  Branch: %s\n", branch))
	b.WriteString("\n")

	// Paths list with cursor navigation for editing
	b.WriteString(labelStyle.Render("Sparse Checkout Paths:"))
	b.WriteString("\n")
	if len(m.pathsList) > 0 {
		for i, path := range m.pathsList {
			if i == m.pathCursor {
				b.WriteString(selectedPathStyle.Render(fmt.Sprintf("► %d. %s", i+1, path)))
			} else {
				b.WriteString(pathItemStyle.Render(fmt.Sprintf("  %d. %s", i+1, path)))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("↑/↓: navigate paths • d: delete selected path"))
	} else {
		b.WriteString(helpStyle.Render("  (no paths added yet)"))
	}
	b.WriteString("\n\n")

	// Command preview
	if user != "" && repo != "" && len(m.pathsList) > 0 {
		b.WriteString(labelStyle.Render("Equivalent command:"))
		b.WriteString("\n")
		cmdParts := []string{"syst git sparse-clone"}
		cmdParts = append(cmdParts, fmt.Sprintf("--provider %s", provider))
		cmdParts = append(cmdParts, fmt.Sprintf("--protocol %s", protocol))
		cmdParts = append(cmdParts, fmt.Sprintf("-u %s", user))
		cmdParts = append(cmdParts, fmt.Sprintf("-r %s", repo))
		if output != "" && output != repo {
			cmdParts = append(cmdParts, fmt.Sprintf("-o %s", output))
		}
		cmdParts = append(cmdParts, fmt.Sprintf("-b %s", branch))
		for _, path := range m.pathsList {
			cmdParts = append(cmdParts, fmt.Sprintf("-p %s", path))
		}

		cmdStr := strings.Join(cmdParts, " ")
		b.WriteString(helpStyle.Render(cmdStr))
		b.WriteString("\n\n")
	}

	// Action buttons
	b.WriteString(labelStyle.Render("Actions:"))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Enter: Proceed with clone • Backspace: Go back to edit • esc: quit"))

	if m.err != nil {
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	}

	return b.String()
}

// getFieldValue returns the current value of an input field, or the default if empty
func (m model) getFieldValue(field inputField, defaultValue string) string {
	value := strings.TrimSpace(m.inputs[field].Value())
	if value == "" {
		return defaultValue
	}
	return value
}

func (m model) nextInput() model {
	m.inputs[m.focused].Blur()
	m.focused = (m.focused + 1) % inputField(len(m.inputs))
	m.inputs[m.focused].Focus()
	return m
}

func (m model) prevInput() model {
	m.inputs[m.focused].Blur()
	if m.focused == 0 {
		m.focused = inputField(len(m.inputs) - 1)
	} else {
		m.focused--
	}
	m.inputs[m.focused].Focus()
	return m
}

func (m *model) buildOptions() {
	m.options = SparseCloneOptions{
		Provider:   m.getFieldValue(providerInput, "github"),
		Protocol:   m.getFieldValue(protocolInput, "ssh"),
		User:       m.getFieldValue(userInput, ""),
		Repository: m.getFieldValue(repositoryInput, ""),
		Output:     m.getFieldValue(outputInput, ""),
		Branch:     m.getFieldValue(branchInput, "main"),
		Paths:      m.pathsList,
	}
}

func (m model) GetOptions() SparseCloneOptions {
	return m.options
}

func (m model) IsSubmitted() bool {
	return m.submitted
}

// RunSparseCloneTUI runs the interactive TUI and returns the configured options
func RunSparseCloneTUI() (*SparseCloneOptions, error) {
	tuiModel := NewSparseCloneTUI()

	p := tea.NewProgram(tuiModel, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run TUI: %w", err)
	}

	resultModel, ok := finalModel.(model)
	if !ok {
		return nil, fmt.Errorf("unexpected model type")
	}

	if !resultModel.IsSubmitted() {
		return nil, fmt.Errorf("operation cancelled")
	}

	opts := resultModel.GetOptions()

	// Validate required fields
	if opts.User == "" {
		return nil, fmt.Errorf("username is required")
	}
	if opts.Repository == "" {
		return nil, fmt.Errorf("repository name is required")
	}
	if len(opts.Paths) == 0 {
		return nil, fmt.Errorf("at least one checkout path is required")
	}

	return &opts, nil
}
