package infoService

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

var (
	titleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Margin(1, 0)
	sectionStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true)
	valueStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	labelStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	quitTextStyle = lipgloss.NewStyle().Margin(1, 0, 2, 4)
	contentStyle  = lipgloss.NewStyle().Padding(1, 2)
)

type SectionItem struct {
	name        string
	description string
	content     string
}

func (s SectionItem) Title() string       { return s.name }
func (s SectionItem) Description() string { return s.description }
func (s SectionItem) FilterValue() string { return s.name }

type model struct {
	list       list.Model
	sections   []SectionItem
	detailMode bool
	selected   SectionItem
	quitting   bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if !m.detailMode {
				if selectedItem := m.list.SelectedItem(); selectedItem != nil {
					if section, ok := selectedItem.(SectionItem); ok {
						m.selected = section
						m.detailMode = true
						return m, nil
					}
				}
			}

		case "esc":
			if m.detailMode {
				m.detailMode = false
				return m, nil
			}
		}
	}

	if !m.detailMode {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return quitTextStyle.Render("Thanks for using syst!")
	}

	if m.detailMode {
		return m.renderSectionDetail()
	}

	content := titleStyle.Render("Repository Information")
	content += "\n\n" + m.list.View()
	content += "\n\nPress Enter to view section details, q to quit"

	return content
}

func (m model) renderSectionDetail() string {
	var content strings.Builder

	content.WriteString(titleStyle.Render(m.selected.name))
	content.WriteString("\n\n")
	content.WriteString(contentStyle.Render(m.selected.content))
	content.WriteString("\n\nPress Esc to go back")

	return content.String()
}

type RepoStats struct {
	Name         string
	Path         string
	Remotes      []RemoteInfo
	Branches     []BranchInfo
	Contributors []ContributorInfo
	TotalCommits int
	FileCount    int
	Size         string
	LastCommit   CommitInfo
}

type RemoteInfo struct {
	Name string
	URL  string
}

type BranchInfo struct {
	Name       string
	IsCurrent  bool
	LastCommit string
}

type ContributorInfo struct {
	Name        string
	Email       string
	CommitCount int
	LastCommit  time.Time
}

type CommitInfo struct {
	Hash    string
	Author  string
	Date    time.Time
	Message string
}

func RunRepoInfoTUI() error {
	stats, err := gatherRepoStats()
	if err != nil {
		return err
	}

	sections := createSections(stats)

	// Convert to list items
	items := make([]list.Item, len(sections))
	for i, section := range sections {
		items[i] = section
	}

	// Create list
	const defaultWidth = 20
	const listHeight = 12

	l := list.New(items, list.NewDefaultDelegate(), defaultWidth, listHeight)
	l.Title = "Repository Sections"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	m := model{
		list:       l,
		sections:   sections,
		detailMode: false,
	}

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		return err
	}

	return nil
}

func gatherRepoStats() (*RepoStats, error) {
	r, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{})
	if err != nil {
		return nil, err
	}

	stats := &RepoStats{}

	// Get repository path and name
	workTree, _ := r.Worktree()
	stats.Path = workTree.Filesystem.Root()
	stats.Name = filepath.Base(stats.Path)

	// Get remotes
	remotes, _ := r.Remotes()
	for _, remote := range remotes {
		config := remote.Config()
		for _, url := range config.URLs {
			stats.Remotes = append(stats.Remotes, RemoteInfo{
				Name: config.Name,
				URL:  url,
			})
		}
	}

	// Get branches
	branches, _ := r.Branches()
	head, _ := r.Head()
	currentBranch := head.Name().Short()

	// #nosec G104 - ForEach callback errors are handled by returning nil in all cases
	branches.ForEach(func(ref *plumbing.Reference) error {
		branchName := ref.Name().Short()
		isCurrent := branchName == currentBranch

		// Get last commit for branch
		commit, err := r.CommitObject(ref.Hash())
		var lastCommitHash string
		if err == nil {
			lastCommitHash = commit.Hash.String()[:8]
		}

		stats.Branches = append(stats.Branches, BranchInfo{
			Name:       branchName,
			IsCurrent:  isCurrent,
			LastCommit: lastCommitHash,
		})
		return nil
	})

	// Get contributors and commit count
	commitIter, _ := r.Log(&git.LogOptions{})
	contributors := make(map[string]*ContributorInfo)
	commitCount := 0

	var lastCommit *object.Commit
	// #nosec G104 - ForEach callback errors are handled by returning nil in all cases
	commitIter.ForEach(func(c *object.Commit) error {
		if lastCommit == nil {
			lastCommit = c
		}

		commitCount++
		key := c.Author.Email
		if contrib, exists := contributors[key]; exists {
			contrib.CommitCount++
			if c.Author.When.After(contrib.LastCommit) {
				contrib.LastCommit = c.Author.When
			}
		} else {
			contributors[key] = &ContributorInfo{
				Name:        c.Author.Name,
				Email:       c.Author.Email,
				CommitCount: 1,
				LastCommit:  c.Author.When,
			}
		}
		return nil
	})

	// Convert contributors map to slice and sort
	for _, contrib := range contributors {
		stats.Contributors = append(stats.Contributors, *contrib)
	}
	sort.Slice(stats.Contributors, func(i, j int) bool {
		return stats.Contributors[i].CommitCount > stats.Contributors[j].CommitCount
	})

	stats.TotalCommits = commitCount

	// Set last commit info
	if lastCommit != nil {
		stats.LastCommit = CommitInfo{
			Hash:    lastCommit.Hash.String(),
			Author:  lastCommit.Author.Name,
			Date:    lastCommit.Author.When,
			Message: strings.TrimSpace(lastCommit.Message),
		}
	}

	// Calculate repository size (only tracked files)
	stats.Size = calculateTrackedRepoSize(r, stats.Path)

	return stats, nil
}

func calculateTrackedRepoSize(repo *git.Repository, repoPath string) string {
	// Get the HEAD commit
	ref, err := repo.Head()
	if err != nil {
		return "Unknown"
	}

	// Get the commit object
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return "Unknown"
	}

	// Get the tree (files in the commit)
	tree, err := commit.Tree()
	if err != nil {
		return "Unknown"
	}

	var totalSize int64
	var fileCount int

	// Walk through all files in the git tree (only tracked files)
	err = tree.Files().ForEach(func(file *object.File) error {
		// Get file size
		totalSize += file.Size
		fileCount++
		return nil
	})

	if err != nil {
		return "Unknown"
	}

	return formatBytes(totalSize)
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB"}
	if exp < len(units) {
		return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
	}

	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), "PB")
}

func createSections(stats *RepoStats) []SectionItem {
	sections := []SectionItem{
		{
			name:        "📋 General Info",
			description: "Repository name, path, and basic details",
			content:     formatGeneralInfo(stats),
		},
		{
			name:        "🌐 Remotes",
			description: fmt.Sprintf("%d configured remotes", len(stats.Remotes)),
			content:     formatRemotes(stats.Remotes),
		},
		{
			name:        "🌿 Branches",
			description: fmt.Sprintf("%d branches", len(stats.Branches)),
			content:     formatBranches(stats.Branches),
		},
		{
			name:        "👥 Contributors",
			description: fmt.Sprintf("%d contributors", len(stats.Contributors)),
			content:     formatContributors(stats.Contributors),
		},
		{
			name:        "📊 Commit Statistics",
			description: fmt.Sprintf("%d total commits", stats.TotalCommits),
			content:     formatCommitStats(stats),
		},
	}

	return sections
}

func formatGeneralInfo(stats *RepoStats) string {
	var content strings.Builder
	content.WriteString(sectionStyle.Render("Repository Details"))
	content.WriteString("\n\n")
	content.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Name:"), valueStyle.Render(stats.Name)))
	content.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Path:"), valueStyle.Render(stats.Path)))
	content.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Size:"), valueStyle.Render(stats.Size)))
	content.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Total Commits:"), valueStyle.Render(fmt.Sprintf("%d", stats.TotalCommits))))

	if stats.LastCommit.Hash != "" {
		content.WriteString("\n")
		content.WriteString(sectionStyle.Render("Last Commit"))
		content.WriteString("\n\n")
		content.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Hash:"), valueStyle.Render(stats.LastCommit.Hash[:8])))
		content.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Author:"), valueStyle.Render(stats.LastCommit.Author)))
		content.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Date:"), valueStyle.Render(stats.LastCommit.Date.Format("2006-01-02 15:04:05"))))
		content.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Message:"), valueStyle.Render(stats.LastCommit.Message)))
	}

	return content.String()
}

func formatRemotes(remotes []RemoteInfo) string {
	var content strings.Builder
	content.WriteString(sectionStyle.Render("Configured Remotes"))
	content.WriteString("\n\n")

	if len(remotes) == 0 {
		content.WriteString(labelStyle.Render("No remotes configured"))
		return content.String()
	}

	for _, remote := range remotes {
		content.WriteString(fmt.Sprintf("%s %s\n", valueStyle.Render(remote.Name+":"), labelStyle.Render(remote.URL)))
	}

	return content.String()
}

func formatBranches(branches []BranchInfo) string {
	var content strings.Builder
	content.WriteString(sectionStyle.Render("Repository Branches"))
	content.WriteString("\n\n")

	for _, branch := range branches {
		marker := " "
		if branch.IsCurrent {
			marker = "*"
		}
		content.WriteString(fmt.Sprintf("%s %s %s\n",
			valueStyle.Render(marker),
			valueStyle.Render(branch.Name),
			labelStyle.Render("("+branch.LastCommit+")")))
	}

	return content.String()
}

func formatContributors(contributors []ContributorInfo) string {
	var content strings.Builder
	content.WriteString(sectionStyle.Render("Top Contributors"))
	content.WriteString("\n\n")

	for i, contrib := range contributors {
		if i >= 10 { // Show top 10
			break
		}
		content.WriteString(fmt.Sprintf("%s (%d commits)\n",
			valueStyle.Render(contrib.Name),
			contrib.CommitCount))
		content.WriteString(fmt.Sprintf("  %s\n", labelStyle.Render(contrib.Email)))
		content.WriteString(fmt.Sprintf("  Last: %s\n\n", labelStyle.Render(contrib.LastCommit.Format("2006-01-02"))))
	}

	return content.String()
}

func formatCommitStats(stats *RepoStats) string {
	var content strings.Builder
	content.WriteString(sectionStyle.Render("Commit Statistics"))
	content.WriteString("\n\n")
	content.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Total Commits:"), valueStyle.Render(fmt.Sprintf("%d", stats.TotalCommits))))
	content.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Contributors:"), valueStyle.Render(fmt.Sprintf("%d", len(stats.Contributors)))))
	content.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Branches:"), valueStyle.Render(fmt.Sprintf("%d", len(stats.Branches)))))
	content.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Remotes:"), valueStyle.Render(fmt.Sprintf("%d", len(stats.Remotes)))))

	return content.String()
}
