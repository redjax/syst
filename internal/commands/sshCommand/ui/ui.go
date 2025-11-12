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
	l.SetShowHelp(false)
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
}

func newKeygenModel() keygenModel {
	alg := textinput.New()
	alg.Placeholder = "rsa or ed25519"
	alg.Focus()
	alg.CharLimit = 10
	alg.Width = 15

	bits := textinput.New()
	bits.Placeholder = "Bits (RSA only)"
	bits.CharLimit = 5
	bits.Width = 10

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

	return keygenModel{
		algorithmInput: alg,
		bitsInput:      bits,
		passwordInput:  pass,
		commentInput:   comment,
		outputInput:    output,
		cursor:         0,
	}
}

func (m keygenModel) Init() tea.Cmd { return nil }

func (m keygenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:

		totalInputs := 5

		// Submit when Enter is pressed on the last input
		if m.cursor == totalInputs-1 && msg.String() == "enter" {
			// Gather options
			opts := sshservice.KeyGenOptions{
				Algorithm: strings.ToLower(m.algorithmInput.Value()),
				Password:  m.passwordInput.Value(),
				Comment:   m.commentInput.Value(),
				FilePath:  m.outputInput.Value(),
			}
			if opts.Algorithm == "rsa" {
				fmt.Sscanf(m.bitsInput.Value(), "%d", &opts.Bits)
			}

			priv, pub, err := sshservice.GenerateKey(opts)
			if err != nil {
				m.message = fmt.Sprintf("Error: %v", err)
			} else {
				m.message = fmt.Sprintf("Private key saved to %s\nPublic key saved to %s", priv, pub)
			}
			m.done = true
			return m, nil
		}

		// Navigation
		switch msg.String() {
		case "esc", "ctrl+c":
			return m, tea.Quit
		case "tab", "down", "j":
			m.cursor = (m.cursor + 1) % totalInputs
		case "shift+tab", "up", "k":
			m.cursor = (m.cursor - 1 + totalInputs) % totalInputs
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

func (m keygenModel) View() string {
	if m.done {
		return m.message + "\nPress any key to exit."
	}
	return fmt.Sprintf(
		"SSH Keygen\n\nAlgorithm: %s\nBits: %s\nPassword: %s\nComment: %s\nOutput: %s\n\nPress Enter/Tab to navigate, ESC to quit.",
		m.algorithmInput.View(),
		m.bitsInput.View(),
		m.passwordInput.View(),
		m.commentInput.View(),
		m.outputInput.View(),
	)
}

// --- Launch function ---
func RunSSHUI() {
	p := tea.NewProgram(
		newMainMenuModel(),
		tea.WithAltScreen(),       // switch to alternate screen buffer
		tea.WithMouseCellMotion(), // optional, allows mouse hover
	)

	finalModel, err := p.Run()
	if err != nil {
		fmt.Println("Error running SSH UI:", err)
		os.Exit(1)
	}

	_ = finalModel
}
