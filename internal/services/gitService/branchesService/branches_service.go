package branchesService

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/redjax/syst/internal/utils/terminal"
)

type ViewMode int

const (
	BranchListView ViewMode = iota
	BranchDetailView
	CommitDetailView
)

type BranchInfo struct {
	Name         string
	IsRemote     bool
	IsCurrent    bool
	LastCommit   *object.Commit
	CommitCount  int
	AheadBehind  string
	LastActivity time.Time
}

type CommitInfo struct {
	Hash      string
	Message   string
	Author    string
	Date      time.Time
	ShortHash string
}

type model struct {
	repo           *git.Repository
	branches       []BranchInfo
	commits        []CommitInfo
	selectedBranch *BranchInfo
	selectedCommit *CommitInfo
	branchList     list.Model
	commitList     list.Model
	viewMode       ViewMode
	tuiHelper      *terminal.ResponsiveTUIHelper
	err            error
	loading        bool
	directBranch   string
}

type branchItem struct {
	branch BranchInfo
}

func (i branchItem) FilterValue() string { return i.branch.Name }
func (i branchItem) Title() string {
	prefix := "  "
	if i.branch.IsCurrent {
		prefix = "* "
	}
	status := ""
	if i.branch.IsRemote {
		status = " (remote)"
	}
	if i.branch.AheadBehind != "" {
		status += " " + i.branch.AheadBehind
	}
	return prefix + i.branch.Name + status
}
func (i branchItem) Description() string {
	if i.branch.LastCommit != nil {
		return fmt.Sprintf("Last: %s - %s",
			i.branch.LastCommit.Author.Name,
			i.branch.LastActivity.Format("2006-01-02 15:04"))
	}
	return "No commits"
}

type commitItem struct {
	commit CommitInfo
}

func (i commitItem) FilterValue() string { return i.commit.Message }
func (i commitItem) Title() string {
	return fmt.Sprintf("%s %s", i.commit.ShortHash, i.commit.Message)
}
func (i commitItem) Description() string {
	return fmt.Sprintf("%s - %s", i.commit.Author, i.commit.Date.Format("2006-01-02 15:04"))
}

type dataLoadedMsg struct {
	branches []BranchInfo
}

type commitsLoadedMsg struct {
	commits []CommitInfo
}

type errMsg struct {
	err error
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginBottom(1)

	infoStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 2).
			MarginTop(1)

	statsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)
)

func (m model) Init() tea.Cmd {
	if m.directBranch != "" {
		return tea.Batch(
			loadBranchData,
			func() tea.Msg {
				// Load commits for the direct branch after data loads
				return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
					return nil
				})
			},
		)
	}
	return loadBranchData
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.tuiHelper.HandleWindowSizeMsg(msg)
		m.branchList.SetWidth(m.tuiHelper.GetWidth())
		m.branchList.SetHeight(m.tuiHelper.GetHeight() - 10)
		m.commitList.SetWidth(m.tuiHelper.GetWidth())
		m.commitList.SetHeight(m.tuiHelper.GetHeight() - 15)
		return m, nil

	case dataLoadedMsg:
		m.branches = msg.branches
		m.loading = false

		items := make([]list.Item, len(m.branches))
		for i, branch := range m.branches {
			items[i] = branchItem{branch: branch}
		}
		m.branchList.SetItems(items)

		// If direct branch specified, find and select it
		if m.directBranch != "" {
			for i, branch := range m.branches {
				if branch.Name == m.directBranch || strings.HasSuffix(branch.Name, "/"+m.directBranch) {
					m.branchList.Select(i)
					m.selectedBranch = &m.branches[i]
					m.viewMode = BranchDetailView
					return m, loadCommitsForBranch(m.selectedBranch.Name)
				}
			}
		}

		return m, nil

	case commitsLoadedMsg:
		m.commits = msg.commits
		items := make([]list.Item, len(m.commits))
		for i, commit := range m.commits {
			items[i] = commitItem{commit: commit}
		}
		m.commitList.SetItems(items)
		return m, nil

	case errMsg:
		m.err = msg.err
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		switch m.viewMode {
		case BranchListView:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c", "esc"))):
				return m, tea.Quit
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				if selected := m.branchList.SelectedItem(); selected != nil {
					branchItem := selected.(branchItem)
					m.selectedBranch = &branchItem.branch
					m.viewMode = BranchDetailView
					return m, loadCommitsForBranch(m.selectedBranch.Name)
				}
			default:
				var cmd tea.Cmd
				m.branchList, cmd = m.branchList.Update(msg)
				return m, cmd
			}

		case BranchDetailView:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))):
				return m, tea.Quit
			case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "backspace"))):
				m.viewMode = BranchListView
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				if selected := m.commitList.SelectedItem(); selected != nil {
					commitItem := selected.(commitItem)
					m.selectedCommit = &commitItem.commit
					m.viewMode = CommitDetailView
					return m, nil
				}
			default:
				var cmd tea.Cmd
				m.commitList, cmd = m.commitList.Update(msg)
				return m, cmd
			}

		case CommitDetailView:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))):
				return m, tea.Quit
			case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "backspace"))):
				m.viewMode = BranchDetailView
				return m, nil
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.loading {
		return "\n  Loading branch data...\n"
	}

	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("\n  Error: %v\n", m.err))
	}

	switch m.viewMode {
	case BranchListView:
		return m.renderBranchList()
	case BranchDetailView:
		return m.renderBranchDetail()
	case CommitDetailView:
		return m.renderCommitDetail()
	}

	return ""
}

func (m model) renderBranchList() string {
	var sections []string

	title := titleStyle.Render("🌿 Branch Explorer")
	sections = append(sections, title)

	sections = append(sections, m.branchList.View())

	help := helpStyle.Render("↑/↓: navigate • enter: select branch • q: quit")
	sections = append(sections, help)

	return strings.Join(sections, "\n")
}

func (m model) renderBranchDetail() string {
	if m.selectedBranch == nil {
		return "No branch selected"
	}

	var sections []string

	title := titleStyle.Render(fmt.Sprintf("🌿 Branch: %s", m.selectedBranch.Name))
	sections = append(sections, title)

	// Branch info
	info := m.renderBranchInfo()
	sections = append(sections, infoStyle.Render(info))

	// Commits subtitle
	subtitle := subtitleStyle.Render("📝 Commits")
	sections = append(sections, subtitle)

	// Commits list
	if len(m.commits) > 0 {
		sections = append(sections, m.commitList.View())
	} else {
		sections = append(sections, "  No commits found")
	}

	help := helpStyle.Render("↑/↓: navigate commits • enter: view commit • esc: back to branches • q: quit")
	sections = append(sections, help)

	return strings.Join(sections, "\n")
}

func (m model) renderBranchInfo() string {
	branch := m.selectedBranch
	var content strings.Builder

	content.WriteString(fmt.Sprintf("Name: %s\n", statsStyle.Render(branch.Name)))
	content.WriteString(fmt.Sprintf("Type: %s\n",
		statsStyle.Render(func() string {
			if branch.IsRemote {
				return "Remote"
			}
			return "Local"
		}())))

	if branch.IsCurrent {
		content.WriteString(fmt.Sprintf("Status: %s\n", statsStyle.Render("Current Branch")))
	}

	content.WriteString(fmt.Sprintf("Total Commits: %s\n",
		statsStyle.Render(fmt.Sprintf("%d", branch.CommitCount))))

	if branch.AheadBehind != "" {
		content.WriteString(fmt.Sprintf("Divergence: %s\n", statsStyle.Render(branch.AheadBehind)))
	}

	if branch.LastCommit != nil {
		content.WriteString(fmt.Sprintf("Last Commit: %s\n",
			statsStyle.Render(branch.LastActivity.Format("2006-01-02 15:04"))))
		content.WriteString(fmt.Sprintf("Last Author: %s\n",
			statsStyle.Render(branch.LastCommit.Author.Name)))
	}

	return content.String()
}

func (m model) renderCommitDetail() string {
	if m.selectedCommit == nil {
		return "No commit selected"
	}

	var sections []string

	title := titleStyle.Render(fmt.Sprintf("📝 Commit: %s", m.selectedCommit.ShortHash))
	sections = append(sections, title)

	// Commit details
	info := m.renderCommitInfo()
	sections = append(sections, infoStyle.Render(info))

	help := helpStyle.Render("esc: back to commits • q: quit")
	sections = append(sections, help)

	return strings.Join(sections, "\n")
}

func (m model) renderCommitInfo() string {
	commit := m.selectedCommit
	var content strings.Builder

	content.WriteString(fmt.Sprintf("Hash: %s\n", statsStyle.Render(commit.Hash)))
	content.WriteString(fmt.Sprintf("Short Hash: %s\n", statsStyle.Render(commit.ShortHash)))
	content.WriteString(fmt.Sprintf("Author: %s\n", statsStyle.Render(commit.Author)))
	content.WriteString(fmt.Sprintf("Date: %s\n", statsStyle.Render(commit.Date.Format("2006-01-02 15:04:05"))))
	content.WriteString(fmt.Sprintf("Message:\n%s\n", statsStyle.Render(commit.Message)))

	return content.String()
}

func loadBranchData() tea.Msg {
	branches, err := gatherBranchData()
	if err != nil {
		return errMsg{err}
	}
	return dataLoadedMsg{branches}
}

func loadCommitsForBranch(branchName string) tea.Cmd {
	return func() tea.Msg {
		commits, err := gatherCommitsForBranch(branchName)
		if err != nil {
			return errMsg{err}
		}
		return commitsLoadedMsg{commits}
	}
}

func gatherBranchData() ([]BranchInfo, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	// Get current branch
	head, err := repo.Head()
	var currentBranchName string
	if err == nil {
		currentBranchName = head.Name().Short()
	}

	var branches []BranchInfo

	// Get local branches
	branchIter, err := repo.Branches()
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %w", err)
	}

	err = branchIter.ForEach(func(ref *plumbing.Reference) error {
		branchName := ref.Name().Short()

		// Get last commit
		commit, err := repo.CommitObject(ref.Hash())
		if err != nil {
			return nil // Skip if can't get commit
		}

		// Count commits
		commitCount, err := countCommitsInBranch(repo, ref.Hash())
		if err != nil {
			commitCount = 0
		}

		// Calculate ahead/behind (simplified)
		aheadBehind := ""
		if currentBranchName != "" && branchName != currentBranchName {
			aheadBehind = calculateAheadBehind(repo, currentBranchName, branchName)
		}

		branches = append(branches, BranchInfo{
			Name:         branchName,
			IsRemote:     false,
			IsCurrent:    branchName == currentBranchName,
			LastCommit:   commit,
			CommitCount:  commitCount,
			AheadBehind:  aheadBehind,
			LastActivity: commit.Author.When,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate branches: %w", err)
	}

	// Get remote branches
	remotes, err := repo.Remotes()
	if err == nil {
		for _, remote := range remotes {
			refs, err := remote.List(&git.ListOptions{})
			if err != nil {
				continue
			}

			for _, ref := range refs {
				if ref.Name().IsBranch() {
					branchName := ref.Name().Short()

					// Skip if we already have this as a local branch
					exists := false
					for _, branch := range branches {
						if branch.Name == branchName {
							exists = true
							break
						}
					}
					if exists {
						continue
					}

					// Try to get commit (may not be available locally)
					commit, err := repo.CommitObject(ref.Hash())
					if err != nil {
						continue // Skip if can't get commit
					}

					commitCount, err := countCommitsInBranch(repo, ref.Hash())
					if err != nil {
						commitCount = 0
					}

					branches = append(branches, BranchInfo{
						Name:         branchName,
						IsRemote:     true,
						IsCurrent:    false,
						LastCommit:   commit,
						CommitCount:  commitCount,
						AheadBehind:  "",
						LastActivity: commit.Author.When,
					})
				}
			}
		}
	}

	// Sort branches by last activity
	sort.Slice(branches, func(i, j int) bool {
		// Current branch first
		if branches[i].IsCurrent {
			return true
		}
		if branches[j].IsCurrent {
			return false
		}
		// Then by last activity
		return branches[i].LastActivity.After(branches[j].LastActivity)
	})

	return branches, nil
}

func gatherCommitsForBranch(branchName string) ([]CommitInfo, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	// Get branch reference
	ref, err := repo.Reference(plumbing.NewBranchReferenceName(branchName), true)
	if err != nil {
		// Try as remote branch
		ref, err = repo.Reference(plumbing.NewRemoteReferenceName("origin", branchName), true)
		if err != nil {
			return nil, fmt.Errorf("failed to find branch %s: %w", branchName, err)
		}
	}

	// Get commit iterator
	cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, fmt.Errorf("failed to get log for branch %s: %w", branchName, err)
	}

	var commits []CommitInfo
	count := 0
	err = cIter.ForEach(func(c *object.Commit) error {
		if count >= 100 { // Limit to 100 commits
			return fmt.Errorf("limit reached")
		}

		commits = append(commits, CommitInfo{
			Hash:      c.Hash.String(),
			ShortHash: c.Hash.String()[:8],
			Message:   strings.Split(c.Message, "\n")[0], // First line only
			Author:    c.Author.Name,
			Date:      c.Author.When,
		})

		count++
		return nil
	})

	if err != nil && err.Error() != "limit reached" {
		return nil, fmt.Errorf("failed to iterate commits: %w", err)
	}

	return commits, nil
}

func countCommitsInBranch(repo *git.Repository, hash plumbing.Hash) (int, error) {
	cIter, err := repo.Log(&git.LogOptions{From: hash})
	if err != nil {
		return 0, err
	}

	count := 0
	err = cIter.ForEach(func(c *object.Commit) error {
		count++
		return nil
	})

	return count, err
}

func calculateAheadBehind(repo *git.Repository, baseBranch, compareBranch string) string {
	// This is a simplified implementation
	// In a real implementation, you'd calculate the exact ahead/behind counts
	baseRef, err := repo.Reference(plumbing.NewBranchReferenceName(baseBranch), true)
	if err != nil {
		return ""
	}

	compareRef, err := repo.Reference(plumbing.NewBranchReferenceName(compareBranch), true)
	if err != nil {
		return ""
	}

	if baseRef.Hash() == compareRef.Hash() {
		return "up to date"
	}

	return "diverged"
}

// RunBranchesExplorer starts the interactive branch explorer TUI
func RunBranchesExplorer(directBranch string) error {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#01FAC6")).
		BorderLeftForeground(lipgloss.Color("#01FAC6"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#DDDDDD"))

	branchList := list.New([]list.Item{}, delegate, 0, 0)
	branchList.Title = "Branches"
	branchList.SetShowStatusBar(false)
	branchList.SetShowHelp(false)

	commitList := list.New([]list.Item{}, delegate, 0, 0)
	commitList.Title = "Commits"
	commitList.SetShowStatusBar(false)
	commitList.SetShowHelp(false)

	m := model{
		branchList:   branchList,
		commitList:   commitList,
		viewMode:     BranchListView,
		loading:      true,
		directBranch: directBranch,
		tuiHelper: terminal.NewResponsiveTUIHelper(),
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
