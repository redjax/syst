package compareService

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/redjax/syst/internal/utils/terminal"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type ViewMode int

const (
	OverviewView ViewMode = iota
	DivergenceView
	SharedHistoryView
	MergeBaseView
	BranchInfoView
)

type ComparisonAnalysis struct {
	Ref1          string
	Ref2          string
	Ref1Commit    string
	Ref2Commit    string
	MergeBase     string
	MergeBaseInfo *object.Commit

	// Divergence analysis
	Ref1Ahead     []CommitInfo
	Ref2Ahead     []CommitInfo
	SharedCommits []CommitInfo

	// Statistics
	Stats ComparisonStats
}

type CommitInfo struct {
	Hash      string
	ShortHash string
	Message   string
	Author    string
	Date      time.Time
	Parents   []string
}

type ComparisonStats struct {
	Ref1AheadBy   int
	Ref2AheadBy   int
	SharedCommits int
	DaysSinceBase int
	TotalCommits  int
}

type model struct {
	// Current state
	currentView ViewMode
	analysis    ComparisonAnalysis

	// UI components
	overviewList   list.Model
	divergenceList list.Model
	sharedList     list.Model
	mergeBaseList  list.Model
	branchInfoList list.Model
	searchInput    textinput.Model

	// UI state
	loading    bool
	err        error
	tuiHelper *terminal.ResponsiveTUIHelper
	showSearch bool
}

// Messages
type comparisonAnalysisMsg struct {
	analysis ComparisonAnalysis
}

type errMsg struct {
	err error
}

// RunComparison starts the comparison tools TUI
func RunComparison(args []string) error {
	// Parse arguments to determine what to compare
	ref1 := "main"
	ref2 := "HEAD"

	if len(args) >= 1 {
		ref1 = args[0]
	}
	if len(args) >= 2 {
		ref2 = args[1]
	}

	// Initialize model
	m := model{
		currentView: OverviewView,
		loading:     true,
		tuiHelper: terminal.NewResponsiveTUIHelper(),
	}

	// Initialize UI components
	m.overviewList = list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	m.overviewList.Title = "⚖️ Comparison Overview"
	m.overviewList.SetShowHelp(false)

	m.divergenceList = list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	m.divergenceList.Title = "🔀 Divergence Analysis"
	m.divergenceList.SetShowHelp(false)

	m.sharedList = list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	m.sharedList.Title = "🤝 Shared History"
	m.sharedList.SetShowHelp(false)

	m.mergeBaseList = list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	m.mergeBaseList.Title = "🔗 Merge Base"
	m.mergeBaseList.SetShowHelp(false)

	m.branchInfoList = list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	m.branchInfoList.Title = "📊 Branch Information"
	m.branchInfoList.SetShowHelp(false)

	m.searchInput = textinput.New()
	m.searchInput.Placeholder = "Search commits..."
	m.searchInput.CharLimit = 100

	// Start the TUI
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Load comparison analysis
	go func() {
		p.Send(loadComparisonAnalysis(ref1, ref2))
	}()

	_, err := p.Run()
	return err
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.tuiHelper.HandleWindowSizeMsg(msg)

		listHeight := m.tuiHelper.GetHeight() - 10
		listWidth := m.tuiHelper.GetWidth() - 4

		m.overviewList.SetSize(listWidth, listHeight)
		m.divergenceList.SetSize(listWidth, listHeight)
		m.sharedList.SetSize(listWidth, listHeight)
		m.mergeBaseList.SetSize(listWidth, listHeight)
		m.branchInfoList.SetSize(listWidth, listHeight)

	case comparisonAnalysisMsg:
		m.loading = false
		m.analysis = msg.analysis

		// Update overview list
		overviewItems := []list.Item{
			OverviewItem{title: "⚖️ Comparison", desc: fmt.Sprintf("%s ↔ %s", m.analysis.Ref1, m.analysis.Ref2)},
			OverviewItem{title: fmt.Sprintf("📈 %s ahead", m.analysis.Ref1), desc: fmt.Sprintf("%d commits", m.analysis.Stats.Ref1AheadBy)},
			OverviewItem{title: fmt.Sprintf("📈 %s ahead", m.analysis.Ref2), desc: fmt.Sprintf("%d commits", m.analysis.Stats.Ref2AheadBy)},
			OverviewItem{title: "🤝 Shared commits", desc: fmt.Sprintf("%d commits", m.analysis.Stats.SharedCommits)},
			OverviewItem{title: "🔗 Merge base", desc: m.analysis.MergeBase[:8]},
		}
		if m.analysis.Stats.DaysSinceBase > 0 {
			overviewItems = append(overviewItems, OverviewItem{
				title: "📅 Days since base",
				desc:  fmt.Sprintf("%d days", m.analysis.Stats.DaysSinceBase),
			})
		}
		m.overviewList.SetItems(overviewItems)

		// Update divergence list (combine both ahead commits)
		var divergenceItems []list.Item
		for _, commit := range m.analysis.Ref1Ahead {
			divergenceItems = append(divergenceItems, CommitInfoItem{
				commit: commit,
				branch: m.analysis.Ref1,
				icon:   "📈",
			})
		}
		for _, commit := range m.analysis.Ref2Ahead {
			divergenceItems = append(divergenceItems, CommitInfoItem{
				commit: commit,
				branch: m.analysis.Ref2,
				icon:   "📈",
			})
		}
		// Sort by date (newest first)
		sort.Slice(divergenceItems, func(i, j int) bool {
			return divergenceItems[i].(CommitInfoItem).commit.Date.After(divergenceItems[j].(CommitInfoItem).commit.Date)
		})
		m.divergenceList.SetItems(divergenceItems)

		// Update shared history list
		sharedItems := make([]list.Item, len(m.analysis.SharedCommits))
		for i, commit := range m.analysis.SharedCommits {
			sharedItems[i] = CommitInfoItem{
				commit: commit,
				branch: "shared",
				icon:   "🤝",
			}
		}
		m.sharedList.SetItems(sharedItems)

		// Update merge base info
		var mergeBaseItems []list.Item
		if m.analysis.MergeBaseInfo != nil {
			mergeBaseItems = []list.Item{
				MergeBaseItem{title: "📝 Commit", desc: m.analysis.MergeBase[:8]},
				MergeBaseItem{title: "👤 Author", desc: m.analysis.MergeBaseInfo.Author.Name},
				MergeBaseItem{title: "📅 Date", desc: m.analysis.MergeBaseInfo.Author.When.Format("2006-01-02 15:04:05")},
				MergeBaseItem{title: "💬 Message", desc: strings.Split(m.analysis.MergeBaseInfo.Message, "\n")[0]},
			}
		}
		m.mergeBaseList.SetItems(mergeBaseItems)

		// Update branch info list
		branchInfoItems := []list.Item{
			BranchInfoItem{title: fmt.Sprintf("📊 %s Info", m.analysis.Ref1), desc: fmt.Sprintf("Commit: %s", m.analysis.Ref1Commit[:8])},
			BranchInfoItem{title: fmt.Sprintf("📊 %s Info", m.analysis.Ref2), desc: fmt.Sprintf("Commit: %s", m.analysis.Ref2Commit[:8])},
			BranchInfoItem{title: "🔄 Total diverged", desc: fmt.Sprintf("%d commits", m.analysis.Stats.Ref1AheadBy+m.analysis.Stats.Ref2AheadBy)},
			BranchInfoItem{title: "📈 Total analyzed", desc: fmt.Sprintf("%d commits", m.analysis.Stats.TotalCommits)},
		}
		m.branchInfoList.SetItems(branchInfoItems)

	case errMsg:
		m.loading = false
		m.err = msg.err

	case tea.KeyMsg:
		// Handle global keys first
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))):
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			if m.showSearch {
				m.showSearch = false
				m.searchInput.Blur()
				return m, nil
			}
			if m.currentView != OverviewView {
				m.currentView = OverviewView
				return m, nil
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("/"))):
			if m.currentView == DivergenceView || m.currentView == SharedHistoryView {
				m.showSearch = !m.showSearch
				if m.showSearch {
					m.searchInput.Focus()
				} else {
					m.searchInput.Blur()
				}
				return m, nil
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("1"))):
			m.currentView = OverviewView
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("2"))):
			m.currentView = DivergenceView
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("3"))):
			m.currentView = SharedHistoryView
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("4"))):
			m.currentView = MergeBaseView
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("5"))):
			m.currentView = BranchInfoView
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
			m.loading = true
			return m, func() tea.Msg {
				return loadComparisonAnalysis(m.analysis.Ref1, m.analysis.Ref2)
			}
		}

		// Handle view-specific keys
		if m.showSearch {
			switch msg.Type {
			case tea.KeyEnter:
				// Perform search
				query := m.searchInput.Value()
				if query != "" {
					// Filter the current list based on the view
					switch m.currentView {
					case DivergenceView:
						m.filterDivergenceList(query)
					case SharedHistoryView:
						m.filterSharedList(query)
					}
				}
				m.showSearch = false
				m.searchInput.Blur()
				return m, nil
			}
			m.searchInput, cmd = m.searchInput.Update(msg)
			return m, cmd
		}

		switch m.currentView {
		case OverviewView:
			m.overviewList, cmd = m.overviewList.Update(msg)
		case DivergenceView:
			m.divergenceList, cmd = m.divergenceList.Update(msg)
		case SharedHistoryView:
			m.sharedList, cmd = m.sharedList.Update(msg)
		case MergeBaseView:
			m.mergeBaseList, cmd = m.mergeBaseList.Update(msg)
		case BranchInfoView:
			m.branchInfoList, cmd = m.branchInfoList.Update(msg)
		}
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *model) filterDivergenceList(query string) {
	var filtered []list.Item
	for _, item := range m.divergenceList.Items() {
		if commitItem, ok := item.(CommitInfoItem); ok {
			if strings.Contains(strings.ToLower(commitItem.commit.Message), strings.ToLower(query)) ||
				strings.Contains(strings.ToLower(commitItem.commit.Author), strings.ToLower(query)) ||
				strings.Contains(strings.ToLower(commitItem.commit.Hash), strings.ToLower(query)) {
				filtered = append(filtered, item)
			}
		}
	}
	m.divergenceList.SetItems(filtered)
}

func (m *model) filterSharedList(query string) {
	var filtered []list.Item
	for _, item := range m.sharedList.Items() {
		if commitItem, ok := item.(CommitInfoItem); ok {
			if strings.Contains(strings.ToLower(commitItem.commit.Message), strings.ToLower(query)) ||
				strings.Contains(strings.ToLower(commitItem.commit.Author), strings.ToLower(query)) ||
				strings.Contains(strings.ToLower(commitItem.commit.Hash), strings.ToLower(query)) {
				filtered = append(filtered, item)
			}
		}
	}
	m.sharedList.SetItems(filtered)
}

func (m model) View() string {
	if m.loading {
		return m.renderLoading()
	}

	if m.err != nil {
		return m.renderError()
	}

	switch m.currentView {
	case OverviewView:
		return m.renderOverview()
	case DivergenceView:
		return m.renderDivergenceView()
	case SharedHistoryView:
		return m.renderSharedHistoryView()
	case MergeBaseView:
		return m.renderMergeBaseView()
	case BranchInfoView:
		return m.renderBranchInfoView()
	default:
		return m.renderOverview()
	}
}

func loadComparisonAnalysis(ref1, ref2 string) tea.Msg {
	analysis, err := analyzeComparison(ref1, ref2)
	if err != nil {
		return errMsg{err}
	}
	return comparisonAnalysisMsg{analysis}
}

func analyzeComparison(ref1, ref2 string) (ComparisonAnalysis, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return ComparisonAnalysis{}, err
	}

	// Resolve references to commits
	ref1Hash, err := resolveRef(repo, ref1)
	if err != nil {
		return ComparisonAnalysis{}, fmt.Errorf("failed to resolve '%s': %w", ref1, err)
	}

	ref2Hash, err := resolveRef(repo, ref2)
	if err != nil {
		return ComparisonAnalysis{}, fmt.Errorf("failed to resolve '%s': %w", ref2, err)
	}

	// Get commit objects
	ref1Commit, err := repo.CommitObject(ref1Hash)
	if err != nil {
		return ComparisonAnalysis{}, err
	}

	ref2Commit, err := repo.CommitObject(ref2Hash)
	if err != nil {
		return ComparisonAnalysis{}, err
	}

	// Find merge base
	mergeBaseHashes, err := ref1Commit.MergeBase(ref2Commit)
	if err != nil {
		return ComparisonAnalysis{}, fmt.Errorf("failed to find merge base: %w", err)
	}

	var mergeBase string
	var mergeBaseCommit *object.Commit
	if len(mergeBaseHashes) > 0 {
		mergeBaseCommit = mergeBaseHashes[0]
		mergeBase = mergeBaseCommit.Hash.String()
	}

	// Get commits that are in ref1 but not in ref2 (ref1 ahead)
	ref1Ahead, err := getCommitRange(repo, mergeBase, ref1Hash.String())
	if err != nil {
		return ComparisonAnalysis{}, fmt.Errorf("failed to get ref1 ahead commits: %w", err)
	}

	// Get commits that are in ref2 but not in ref1 (ref2 ahead)
	ref2Ahead, err := getCommitRange(repo, mergeBase, ref2Hash.String())
	if err != nil {
		return ComparisonAnalysis{}, fmt.Errorf("failed to get ref2 ahead commits: %w", err)
	}

	// Get shared commits (from merge base backwards)
	sharedCommits, err := getSharedCommits(repo, mergeBase, 20) // Limit to recent 20
	if err != nil {
		return ComparisonAnalysis{}, fmt.Errorf("failed to get shared commits: %w", err)
	}

	// Calculate statistics
	daysSinceBase := 0
	if mergeBaseCommit != nil {
		daysSinceBase = int(time.Since(mergeBaseCommit.Author.When).Hours() / 24)
	}

	stats := ComparisonStats{
		Ref1AheadBy:   len(ref1Ahead),
		Ref2AheadBy:   len(ref2Ahead),
		SharedCommits: len(sharedCommits),
		DaysSinceBase: daysSinceBase,
		TotalCommits:  len(ref1Ahead) + len(ref2Ahead) + len(sharedCommits),
	}

	return ComparisonAnalysis{
		Ref1:          ref1,
		Ref2:          ref2,
		Ref1Commit:    ref1Hash.String(),
		Ref2Commit:    ref2Hash.String(),
		MergeBase:     mergeBase,
		MergeBaseInfo: mergeBaseCommit,
		Ref1Ahead:     ref1Ahead,
		Ref2Ahead:     ref2Ahead,
		SharedCommits: sharedCommits,
		Stats:         stats,
	}, nil
}

func resolveRef(repo *git.Repository, ref string) (plumbing.Hash, error) {
	// Try to resolve as a hash first
	if hash := plumbing.NewHash(ref); !hash.IsZero() {
		// Check if it's a valid commit
		if _, err := repo.CommitObject(hash); err == nil {
			return hash, nil
		}
	}

	// Try to resolve as a reference
	resolved, err := repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return plumbing.ZeroHash, err
	}

	return *resolved, nil
}

func getCommitRange(repo *git.Repository, fromCommit, toCommit string) ([]CommitInfo, error) {
	var commits []CommitInfo

	// If fromCommit is empty, get all commits up to toCommit
	if fromCommit == "" {
		return getCommitHistory(repo, toCommit, 50) // Limit to 50 commits
	}

	toHash := plumbing.NewHash(toCommit)
	fromHash := plumbing.NewHash(fromCommit)

	// Get commit iterator from toCommit
	toCommitObj, err := repo.CommitObject(toHash)
	if err != nil {
		return commits, err
	}

	// Traverse commits from toCommit backwards until we reach fromCommit
	iter := object.NewCommitPreorderIter(toCommitObj, nil, nil)
	defer iter.Close()

	err = iter.ForEach(func(commit *object.Commit) error {
		// Stop if we reached the fromCommit
		if commit.Hash == fromHash {
			return nil
		}

		commits = append(commits, CommitInfo{
			Hash:      commit.Hash.String(),
			ShortHash: commit.Hash.String()[:8],
			Message:   strings.Split(commit.Message, "\n")[0],
			Author:    commit.Author.Name,
			Date:      commit.Author.When,
			Parents:   getParentHashes(commit),
		})

		// Limit to prevent too many commits
		if len(commits) > 100 {
			return fmt.Errorf("commit_limit_reached")
		}

		return nil
	})

	// Ignore the "commit_limit_reached" error as it's expected
	if err != nil && err.Error() == "commit_limit_reached" {
		err = nil
	}

	return commits, err
}

func getSharedCommits(repo *git.Repository, fromCommit string, limit int) ([]CommitInfo, error) {
	if fromCommit == "" {
		return []CommitInfo{}, nil
	}

	return getCommitHistory(repo, fromCommit, limit)
}

func getCommitHistory(repo *git.Repository, commitHash string, limit int) ([]CommitInfo, error) {
	var commits []CommitInfo

	hash := plumbing.NewHash(commitHash)
	commit, err := repo.CommitObject(hash)
	if err != nil {
		return commits, err
	}

	iter := object.NewCommitPreorderIter(commit, nil, nil)
	defer iter.Close()

	count := 0
	err = iter.ForEach(func(c *object.Commit) error {
		if count >= limit {
			// Use a special error to indicate we've reached our limit
			return fmt.Errorf("limit_reached")
		}

		commits = append(commits, CommitInfo{
			Hash:      c.Hash.String(),
			ShortHash: c.Hash.String()[:8],
			Message:   strings.Split(c.Message, "\n")[0],
			Author:    c.Author.Name,
			Date:      c.Author.When,
			Parents:   getParentHashes(c),
		})

		count++
		return nil
	})

	// Ignore the "limit_reached" error as it's expected
	if err != nil && err.Error() == "limit_reached" {
		err = nil
	}

	return commits, err
}

func getParentHashes(commit *object.Commit) []string {
	var parents []string
	for _, parent := range commit.ParentHashes {
		parents = append(parents, parent.String())
	}
	return parents
}

// List item types
type OverviewItem struct {
	title string
	desc  string
}

func (o OverviewItem) Title() string       { return o.title }
func (o OverviewItem) Description() string { return o.desc }
func (o OverviewItem) FilterValue() string { return o.title + " " + o.desc }

type CommitInfoItem struct {
	commit CommitInfo
	branch string
	icon   string
}

func (c CommitInfoItem) Title() string {
	var branchIndicator string
	if c.branch != "shared" {
		branchIndicator = fmt.Sprintf(" [%s]", c.branch)
	}
	return fmt.Sprintf("%s %s • %s%s", c.icon, c.commit.ShortHash, c.commit.Message, branchIndicator)
}

func (c CommitInfoItem) Description() string {
	return fmt.Sprintf("%s • %s", c.commit.Author, c.commit.Date.Format("2006-01-02 15:04"))
}

func (c CommitInfoItem) FilterValue() string {
	return c.commit.Hash + " " + c.commit.Message + " " + c.commit.Author
}

type MergeBaseItem struct {
	title string
	desc  string
}

func (m MergeBaseItem) Title() string       { return m.title }
func (m MergeBaseItem) Description() string { return m.desc }
func (m MergeBaseItem) FilterValue() string { return m.title + " " + m.desc }

type BranchInfoItem struct {
	title string
	desc  string
}

func (b BranchInfoItem) Title() string       { return b.title }
func (b BranchInfoItem) Description() string { return b.desc }
func (b BranchInfoItem) FilterValue() string { return b.title + " " + b.desc }

// Render functions
func (m model) renderLoading() string {
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginTop(2).
		MarginLeft(2)

	return style.Render("⚖️ Analyzing comparison...")
}

func (m model) renderError() string {
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("196")).
		MarginTop(2).
		MarginLeft(2)

	return style.Render(fmt.Sprintf("❌ Error: %v", m.err))
}

func (m model) renderOverview() string {
	var content strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	title := fmt.Sprintf("⚖️ Comparison: %s ↔ %s", m.analysis.Ref1, m.analysis.Ref2)
	content.WriteString(headerStyle.Render(title))
	content.WriteString("\n")

	// Overview list
	content.WriteString(m.overviewList.View())
	content.WriteString("\n")

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	help := "1: overview • 2: divergence • 3: shared • 4: merge base • 5: info • r: refresh • q: quit"
	content.WriteString(helpStyle.Render(help))

	return content.String()
}

func (m model) renderDivergenceView() string {
	var content strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	title := fmt.Sprintf("🔀 Divergence: %s (%d) ↔ %s (%d)",
		m.analysis.Ref1, m.analysis.Stats.Ref1AheadBy,
		m.analysis.Ref2, m.analysis.Stats.Ref2AheadBy)
	content.WriteString(headerStyle.Render(title))
	content.WriteString("\n")

	// Search input
	if m.showSearch {
		searchStyle := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("39")).
			Padding(0, 1).
			MarginBottom(1)

		content.WriteString(searchStyle.Render("🔍 " + m.searchInput.View()))
		content.WriteString("\n")
	}

	// Divergence list
	content.WriteString(m.divergenceList.View())
	content.WriteString("\n")

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	help := "1: overview • 2: divergence • 3: shared • /: search • esc: back • q: quit"
	content.WriteString(helpStyle.Render(help))

	return content.String()
}

func (m model) renderSharedHistoryView() string {
	var content strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	title := fmt.Sprintf("🤝 Shared History (%d commits)", m.analysis.Stats.SharedCommits)
	content.WriteString(headerStyle.Render(title))
	content.WriteString("\n")

	// Search input
	if m.showSearch {
		searchStyle := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("39")).
			Padding(0, 1).
			MarginBottom(1)

		content.WriteString(searchStyle.Render("🔍 " + m.searchInput.View()))
		content.WriteString("\n")
	}

	// Shared commits list
	content.WriteString(m.sharedList.View())
	content.WriteString("\n")

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	help := "1: overview • 2: divergence • 3: shared • /: search • esc: back • q: quit"
	content.WriteString(helpStyle.Render(help))

	return content.String()
}

func (m model) renderMergeBaseView() string {
	var content strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	title := "🔗 Merge Base Analysis"
	content.WriteString(headerStyle.Render(title))
	content.WriteString("\n")

	// Merge base information
	if m.analysis.MergeBase != "" {
		infoStyle := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("238")).
			Padding(1, 2).
			MarginBottom(1)

		var info strings.Builder
		info.WriteString(fmt.Sprintf("🔗 Merge Base: %s\n", m.analysis.MergeBase[:8]))
		if m.analysis.MergeBaseInfo != nil {
			info.WriteString(fmt.Sprintf("👤 Author: %s\n", m.analysis.MergeBaseInfo.Author.Name))
			info.WriteString(fmt.Sprintf("📅 Date: %s\n", m.analysis.MergeBaseInfo.Author.When.Format("2006-01-02 15:04:05")))
			info.WriteString(fmt.Sprintf("💬 Message: %s\n", strings.Split(m.analysis.MergeBaseInfo.Message, "\n")[0]))
			info.WriteString(fmt.Sprintf("📅 Days ago: %d\n", m.analysis.Stats.DaysSinceBase))
		}

		content.WriteString(infoStyle.Render(info.String()))
	} else {
		noBaseStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true).
			MarginBottom(1)
		content.WriteString(noBaseStyle.Render("No common merge base found"))
		content.WriteString("\n")
	}

	// Additional merge base details
	content.WriteString(m.mergeBaseList.View())
	content.WriteString("\n")

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	help := "1: overview • 2: divergence • 3: shared • 4: merge base • 5: info • esc: back • q: quit"
	content.WriteString(helpStyle.Render(help))

	return content.String()
}

func (m model) renderBranchInfoView() string {
	var content strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	title := "📊 Branch Information"
	content.WriteString(headerStyle.Render(title))
	content.WriteString("\n")

	// Detailed comparison stats
	statsStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("238")).
		Padding(1, 2).
		MarginBottom(1)

	var stats strings.Builder
	stats.WriteString("📊 Comparison Summary:\n\n")
	stats.WriteString(fmt.Sprintf("📈 %s is ahead by: %d commits\n", m.analysis.Ref1, m.analysis.Stats.Ref1AheadBy))
	stats.WriteString(fmt.Sprintf("📈 %s is ahead by: %d commits\n", m.analysis.Ref2, m.analysis.Stats.Ref2AheadBy))
	stats.WriteString(fmt.Sprintf("🤝 Shared commits: %d\n", m.analysis.Stats.SharedCommits))
	stats.WriteString(fmt.Sprintf("🔄 Total divergence: %d commits\n", m.analysis.Stats.Ref1AheadBy+m.analysis.Stats.Ref2AheadBy))
	stats.WriteString(fmt.Sprintf("📈 Total analyzed: %d commits\n", m.analysis.Stats.TotalCommits))
	if m.analysis.Stats.DaysSinceBase > 0 {
		stats.WriteString(fmt.Sprintf("📅 Days since divergence: %d\n", m.analysis.Stats.DaysSinceBase))
	}

	content.WriteString(statsStyle.Render(stats.String()))

	// Branch info list
	content.WriteString(m.branchInfoList.View())
	content.WriteString("\n")

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	help := "1: overview • 2: divergence • 3: shared • 4: merge base • 5: info • r: refresh • q: quit"
	content.WriteString(helpStyle.Render(help))

	return content.String()
}
