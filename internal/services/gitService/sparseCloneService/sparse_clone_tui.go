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

	b.WriteString(titleStyle.Render("🔧 Git Sparse Clone Configuration"))
	b.WriteString("\n\n")

	// Provider
	b.WriteString(labelStyle.Render("Git Provider:"))
	b.WriteString("\n")
	b.WriteString(m.inputs[providerInput].View())
	if m.focused == providerInput {
		b.WriteString(helpStyle.Render(" (github, gitlab, codeberg)"))
	}
	b.WriteString("\n\n")

	// Protocol
	b.WriteString(labelStyle.Render("Protocol:"))
	b.WriteString("\n")
	b.WriteString(m.inputs[protocolInput].View())
	if m.focused == protocolInput {
		b.WriteString(helpStyle.Render(" (ssh, https)"))
	}
	b.WriteString("\n\n")

	// Username
	b.WriteString(labelStyle.Render("Username/Organization:"))
	b.WriteString("\n")
	b.WriteString(m.inputs[userInput].View())
	b.WriteString("\n\n")

	// Repository
	b.WriteString(labelStyle.Render("Repository Name:"))
	b.WriteString("\n")
	b.WriteString(m.inputs[repositoryInput].View())
	b.WriteString("\n\n")

	// Output Directory
	b.WriteString(labelStyle.Render("Output Directory:"))
	b.WriteString("\n")
	b.WriteString(m.inputs[outputInput].View())
	if m.focused == outputInput {
		b.WriteString(helpStyle.Render(" (optional)"))
	}
	b.WriteString("\n\n")

	// Branch
	b.WriteString(labelStyle.Render("Branch:"))
	b.WriteString("\n")
	b.WriteString(m.inputs[branchInput].View())
	b.WriteString("\n\n")

	// Paths
	b.WriteString(labelStyle.Render("Sparse Checkout Paths:"))
	b.WriteString("\n")
	if len(m.pathsList) > 0 {
		for i, path := range m.pathsList {
			if m.pathEditMode && i == m.pathCursor {
				b.WriteString(selectedPathStyle.Render(fmt.Sprintf("► %d. %s", i+1, path)))
			} else {
				b.WriteString(pathItemStyle.Render(fmt.Sprintf("  %d. %s", i+1, path)))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	b.WriteString(m.inputs[pathsInput].View())
	b.WriteString("\n")
	if m.focused == pathsInput {
		if len(m.pathsList) > 0 {
			if m.pathEditMode {
				b.WriteString(helpStyle.Render("Path Edit Mode: ↑/↓: navigate • d: delete • Enter: exit edit"))
			} else {
				currentInput := strings.TrimSpace(m.inputs[pathsInput].Value())
				if currentInput == "" {
					b.WriteString(helpStyle.Render("Enter: add path • p: edit existing paths • Backspace: remove last"))
				} else {
					b.WriteString(helpStyle.Render("Enter: add path • Backspace: remove last"))
				}
			}
		} else {
			b.WriteString(helpStyle.Render("Enter: add path"))
		}
	}
	b.WriteString("\n\n")

	// Confirmation
	if m.focused >= confirmInput && len(m.pathsList) > 0 {
		b.WriteString(labelStyle.Render("Proceed with sparse clone? (y/N):"))
		b.WriteString("\n")
		b.WriteString(m.inputs[confirmInput].View())
		b.WriteString("\n\n")
	}

	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n")
	}

	// Help text
	if m.pathEditMode {
		// Don't show additional help when in path edit mode
		// (help is already shown above near the paths)
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
