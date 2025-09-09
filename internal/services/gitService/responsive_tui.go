package gitservice

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TUIModel is an interface that all bubbletea models should implement
// to support responsive terminal sizing
type TUIModel interface {
	SetSize(width, height int)
	GetSize() (width, height int)
}

// ResponsiveTUIHelper provides common utilities for responsive TUI design
type ResponsiveTUIHelper struct {
	width  int
	height int
}

// NewResponsiveTUIHelper creates a new responsive TUI helper
func NewResponsiveTUIHelper() *ResponsiveTUIHelper {
	return &ResponsiveTUIHelper{
		width:  80, // Default width
		height: 24, // Default height
	}
}

// SetSize updates the terminal dimensions
func (h *ResponsiveTUIHelper) SetSize(width, height int) {
	h.width = width
	h.height = height
}

// GetSize returns the current terminal dimensions
func (h *ResponsiveTUIHelper) GetSize() (int, int) {
	return h.width, h.height
}

// GetResponsiveSectionStyle returns a section style that adapts to terminal width
func (h *ResponsiveTUIHelper) GetResponsiveSectionStyle(baseStyle lipgloss.Style) lipgloss.Style {
	maxWidth := h.width - 4 // Account for borders and padding
	if maxWidth < 40 {
		maxWidth = 40 // Minimum width
	}

	return baseStyle.Width(maxWidth)
}

// GetResponsiveTitleStyle returns a title style that adapts to terminal width
func (h *ResponsiveTUIHelper) GetResponsiveTitleStyle(baseStyle lipgloss.Style) lipgloss.Style {
	return baseStyle.Width(h.width - 2) // Account for padding
}

// GetContentWidth returns the available width for content (accounting for borders)
func (h *ResponsiveTUIHelper) GetContentWidth() int {
	contentWidth := h.width - 8 // Account for section borders and padding
	if contentWidth < 40 {
		contentWidth = 40
	}
	return contentWidth
}

// CalculateBarLength calculates the appropriate bar length for charts based on terminal width
func (h *ResponsiveTUIHelper) CalculateBarLength(labelWidth int, maxBarLength int) int {
	availableWidth := h.width - labelWidth - 10 // Account for labels, counts, and margins
	if availableWidth < 10 {
		availableWidth = 10
	}
	if availableWidth > maxBarLength {
		availableWidth = maxBarLength
	}
	return availableWidth
}

// CalculateMaxItemsForHeight calculates how many items can fit in the available height
func (h *ResponsiveTUIHelper) CalculateMaxItemsForHeight(linesPerItem int, reservedLines int) int {
	availableLines := h.height - reservedLines
	if availableLines <= 0 {
		return 1
	}

	maxItems := availableLines / linesPerItem
	if maxItems < 1 {
		maxItems = 1
	}
	return maxItems
}

// TruncateContentToHeight ensures content fits within terminal height
func (h *ResponsiveTUIHelper) TruncateContentToHeight(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) <= h.height-1 {
		return content
	}

	// Truncate if too many lines
	lines = lines[:h.height-2]
	lines = append(lines, lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Render("... (content truncated)"))
	return strings.Join(lines, "\n")
}

// CenterContent centers content both horizontally and vertically
func (h *ResponsiveTUIHelper) CenterContent(content string) string {
	style := lipgloss.NewStyle().
		Width(h.width).
		Height(h.height).
		Align(lipgloss.Center, lipgloss.Center)
	return style.Render(content)
}

// CreateTwoColumnLayout creates a responsive two-column layout
// Falls back to single column on smaller terminals
func (h *ResponsiveTUIHelper) CreateTwoColumnLayout(leftItems, rightItems []string) string {
	var result strings.Builder

	if h.width >= 80 {
		// Two-column layout for larger terminals
		contentWidth := h.GetContentWidth()
		leftStyle := lipgloss.NewStyle().Width(contentWidth / 2)
		rightStyle := lipgloss.NewStyle().Width(contentWidth / 2)

		maxItems := len(leftItems)
		if len(rightItems) > maxItems {
			maxItems = len(rightItems)
		}

		for i := 0; i < maxItems; i++ {
			leftText := ""
			rightText := ""

			if i < len(leftItems) {
				leftText = leftStyle.Render(leftItems[i])
			}
			if i < len(rightItems) {
				rightText = rightStyle.Render(rightItems[i])
			}

			result.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, leftText, rightText))
			result.WriteString("\n")
		}
	} else {
		// Single column for smaller terminals
		allItems := make([]string, 0, len(leftItems)+len(rightItems))
		allItems = append(allItems, leftItems...)
		allItems = append(allItems, rightItems...)

		for _, item := range allItems {
			result.WriteString(item + "\n")
		}
	}

	return result.String()
}

// HandleWindowSizeMsg is a helper function to handle tea.WindowSizeMsg
func (h *ResponsiveTUIHelper) HandleWindowSizeMsg(msg tea.WindowSizeMsg) {
	h.SetSize(msg.Width, msg.Height)
}

// CommonResponsiveUpdateHandler provides a standard way to handle window size messages
func CommonResponsiveUpdateHandler(msg tea.Msg, helper *ResponsiveTUIHelper) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		helper.HandleWindowSizeMsg(msg)
		return nil
	}
	return nil
}
