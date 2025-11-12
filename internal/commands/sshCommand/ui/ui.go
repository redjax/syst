package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	sshservice "github.com/redjax/syst/internal/services/sshService"
)

// --- Main menu ---

var subcommands = []string{"keygen"}

type mainMenuModel struct {
	list list.Model
}

func newMainMenuModel() mainMenuModel {
	items := make([]list.Item, len(subcommands))
	for i, cmd := range subcommands {
		items[i] = listItem(cmd)
	}

	l := list.New(items, list.NewDefaultDelegate(), 30, 10)
	l.Title = "Select SSH subcommand"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(true)
	l.Styles.Title = l.Styles.Title.Bold(true)

	return mainMenuModel{list: l}
}

func (m mainMenuModel) Init() tea.Cmd { return nil }

func (m mainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			i, ok := m.list.SelectedItem().(listItem)
			if ok && string(i) == "keygen" {
				return newKeygenModel(), nil
			}
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	return m, cmd
}

func (m mainMenuModel) View() string {
	return m.list.View()
}

type listItem string

func (i listItem) FilterValue() string { return string(i) }
func (i listItem) Title() string       { return string(i) }
func (i listItem) Description() string { return "" }

// --- Keygen form ---

type keygenModel struct {
	algorithmInput textinput.Model
	bitsInput      textinput.Model
	passwordInput  textinput.Model
	commentInput   textinput.Model
	outputInput    textinput.Model
	cursor         int
	done           bool
	message        string
	returnToMain   bool
}

func newKeygenModel() *keygenModel {
	alg := textinput.New()
	alg.Placeholder = "rsa or ed25519"
	alg.Focus()
	alg.CharLimit = 10
	alg.Width = 15

	bits := textinput.New()
	bits.Placeholder = "Bits (RSA only)"
	bits.CharLimit = 5
	bits.Width = 20

	pass := textinput.New()
	pass.Placeholder = "Password (optional)"
	pass.EchoMode = textinput.EchoPassword
	pass.EchoCharacter = '*'
	pass.Width = 20

	comment := textinput.New()
	comment.Placeholder = "Comment (optional)"
	comment.Width = 20

	output := textinput.New()
	output.Placeholder = "~/.ssh/id_rsa or id_ed25519"
	output.Width = 30

	return &keygenModel{
		algorithmInput: alg,
		bitsInput:      bits,
		passwordInput:  pass,
		commentInput:   comment,
		outputInput:    output,
		cursor:         0,
	}
}

func (m keygenModel) Init() tea.Cmd { return nil }

func (m *keygenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	totalInputs := 5

	switch msg := msg.(type) {
	case tea.KeyMsg:

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			// return new main menu model
			return newMainMenuModel(), nil
		case "tab", "down":
			m.cursor = (m.cursor + 1) % totalInputs
		case "shift+tab", "up":
			m.cursor = (m.cursor - 1 + totalInputs) % totalInputs
		case "enter":
			// Read inputs
			algo := strings.ToLower(strings.TrimSpace(m.algorithmInput.Value()))
			output := strings.TrimSpace(m.outputInput.Value())

			// Validate inputs
			if algo != "rsa" && algo != "ed25519" {
				m.message = "Error: algorithm must be 'rsa' or 'ed25519'"
				return m, nil
			}
			if output == "" {
				m.message = "Error: output file path cannot be empty"
				return m, nil
			}

			opts := sshservice.KeyGenOptions{
				Algorithm: algo,
				Password:  m.passwordInput.Value(),
				Comment:   m.commentInput.Value(),
				FilePath:  output,
			}

			if algo == "rsa" {
				var bits int
				if n, err := fmt.Sscanf(m.bitsInput.Value(), "%d", &bits); n == 1 && err == nil {
					opts.Bits = bits
				} else {
					opts.Bits = 4096
				}
			}

			// Attempt to generate key
			priv, pub, err := sshservice.GenerateKey(opts)
			if err != nil {
				m.message = fmt.Sprintf("Error: %v", err)
				return m, nil // keep form active
			}

			m.message = fmt.Sprintf("Private key saved to %s\nPublic key saved to %s", priv, pub)
			// Form stays active so user can press Esc to go back
			return m, nil
		}
	}

	// Update focus after cursor change
	m.updateFocus()

	// Update all inputs
	var cmd tea.Cmd
	m.algorithmInput, cmd = m.algorithmInput.Update(msg)
	m.bitsInput, _ = m.bitsInput.Update(msg)
	m.passwordInput, _ = m.passwordInput.Update(msg)
	m.commentInput, _ = m.commentInput.Update(msg)
	m.outputInput, _ = m.outputInput.Update(msg)

	return m, cmd
}

func (m *keygenModel) updateFocus() {
	m.algorithmInput.Blur()
	m.bitsInput.Blur()
	m.passwordInput.Blur()
	m.commentInput.Blur()
	m.outputInput.Blur()

	switch m.cursor {
	case 0:
		m.algorithmInput.Focus()
	case 1:
		m.bitsInput.Focus()
	case 2:
		m.passwordInput.Focus()
	case 3:
		m.commentInput.Focus()
	case 4:
		m.outputInput.Focus()
	}
}

func (m *keygenModel) View() string {
	if m.returnToMain {
		return ""
	}

	// Build input fields UI
	ui := fmt.Sprintf(
		"SSH Keygen\n\nAlgorithm: %s\nBits: %s\nPassword: %s\nComment: %s\nOutput: %s\n\n",
		m.algorithmInput.View(),
		m.bitsInput.View(),
		m.passwordInput.View(),
		m.commentInput.View(),
		m.outputInput.View(),
	)

	// Message area
	if m.message != "" {
		ui += fmt.Sprintf("%s\n\n", m.message)
	}

	// Navigation hints in dim gray
	hints := "\033[90mNavigation:\n" +
		"  Tab / Down  → Next field\n" +
		"  Shift+Tab / Up  → Previous field\n" +
		"  Enter → Submit\n" +
		"  Esc → Back to main menu, Ctrl+C → Quit\033[0m\n"

	return ui + hints
}

// --- Launch function ---
func RunSSHUI() {
	// Start with main menu
	var currentModel tea.Model = newMainMenuModel()

	p := tea.NewProgram(
		currentModel,
		tea.WithAltScreen(),
	)

	for {
		var err error
		currentModel, err = p.Run()
		if err != nil {
			fmt.Println("Error running SSH UI:", err)
			os.Exit(1)
		}

		switch m := currentModel.(type) {
		case *keygenModel:
			if m.returnToMain {
				// Return to main menu
				currentModel = newMainMenuModel()
				p = tea.NewProgram(currentModel, tea.WithAltScreen())
				continue
			}
		}

		// Otherwise exit
		break
	}
}
