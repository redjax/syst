package contributorsService

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
	"github.com/go-git/go-git/v5/plumbing/object"
)

type ViewMode int

const (
	ContributorListView ViewMode = iota
	ContributorDetailView
	TimelineView
)

type ContributorData struct {
	Name              string
	Email             string
	TotalCommits      int
	FirstCommit       time.Time
	LastCommit        time.Time
	LinesAdded        int
	LinesDeleted      int
	FilesModified     int
	CommitsByMonth    map[string]int
	CommitsByHour     map[int]int
	CommitsByDay      map[int]int
	RecentCommits     []CommitSummary
	TopFiles          []FileStat
	CommitMessages    []string
	AverageCommitSize int
	LargestCommit     CommitSummary
	Percentage        float64
}

type CommitSummary struct {
	Hash         string
	ShortHash    string
	Message      string
	Date         time.Time
	FilesChanged int
	Additions    int
	Deletions    int
}

type FileStat struct {
	Path          string
	Modifications int
}

type OverallStats struct {
	TotalContributors int
	TotalCommits      int
	DateRange         string
	MostActive        string
	RecentActivity    []ContributorActivity
}

type ContributorActivity struct {
	Name   string
	Period string
	Count  int
}

type model struct {
	contributors    []ContributorData
	selectedIndex   int
	overallStats    OverallStats
	contributorList list.Model
	viewMode        ViewMode
	width           int
	height          int
	err             error
	loading         bool
}

type contributorItem struct {
	contributor ContributorData
}

func (i contributorItem) FilterValue() string { return i.contributor.Name }
func (i contributorItem) Title() string {
	commits := i.contributor.TotalCommits
	percentage := i.contributor.Percentage
	return fmt.Sprintf("%s (%d commits, %.1f%%)", i.contributor.Name, commits, percentage)
}
func (i contributorItem) Description() string {
	lastActive := i.contributor.LastCommit.Format("2006-01-02")
	linesChanged := i.contributor.LinesAdded + i.contributor.LinesDeleted
	return fmt.Sprintf("Last active: %s • %d lines changed • %d files",
		lastActive, linesChanged, i.contributor.FilesModified)
}

type dataLoadedMsg struct {
	contributors []ContributorData
	overallStats OverallStats
}

type errMsg struct {
	err error
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#FF6B35")).
			Padding(0, 1)

	sectionStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 2).
			MarginTop(1)

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginBottom(1)

	statsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	highlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#01FAC6")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)
)

func (m model) Init() tea.Cmd {
	return loadContributorData
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.contributorList.SetWidth(msg.Width)
		m.contributorList.SetHeight(msg.Height - 10)
		return m, nil

	case dataLoadedMsg:
		m.contributors = msg.contributors
		m.overallStats = msg.overallStats
		m.loading = false

		items := make([]list.Item, len(m.contributors))
		for i, contributor := range m.contributors {
			items[i] = contributorItem{contributor: contributor}
		}
		m.contributorList.SetItems(items)

		return m, nil

	case errMsg:
		m.err = msg.err
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		switch m.viewMode {
		case ContributorListView:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c", "esc"))):
				return m, tea.Quit
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				if selected := m.contributorList.SelectedItem(); selected != nil {
					m.selectedIndex = m.contributorList.Index()
					m.viewMode = ContributorDetailView
					return m, nil
				}
			case key.Matches(msg, key.NewBinding(key.WithKeys("t"))):
				m.viewMode = TimelineView
				return m, nil
			default:
				var cmd tea.Cmd
				m.contributorList, cmd = m.contributorList.Update(msg)
				return m, cmd
			}

		case ContributorDetailView, TimelineView:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))):
				return m, tea.Quit
			case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "backspace"))):
				m.viewMode = ContributorListView
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("t"))):
				if m.viewMode == ContributorDetailView {
					m.viewMode = TimelineView
				} else {
					m.viewMode = ContributorDetailView
				}
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
				if m.selectedIndex > 0 {
					m.selectedIndex--
				}
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
				if m.selectedIndex < len(m.contributors)-1 {
					m.selectedIndex++
				}
				return m, nil
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.loading {
		return "\n  Analyzing contributor data...\n"
	}

	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("\n  Error: %v\n", m.err))
	}

	switch m.viewMode {
	case ContributorListView:
		return m.renderContributorList()
	case ContributorDetailView:
		return m.renderContributorDetail()
	case TimelineView:
		return m.renderTimelineView()
	}

	return ""
}

func (m model) renderContributorList() string {
	var sections []string

	title := titleStyle.Render("👥 Contributors Analysis")
	sections = append(sections, title)

	// Overall stats
	overviewContent := m.renderOverallStats()
	sections = append(sections, sectionStyle.Render(overviewContent))

	// Contributors list
	sections = append(sections, m.contributorList.View())

	help := helpStyle.Render("↑/↓: navigate • enter: details • t: timeline • q: quit")
	sections = append(sections, help)

	return strings.Join(sections, "\n")
}

func (m model) renderOverallStats() string {
	stats := m.overallStats
	var content strings.Builder

	content.WriteString(headerStyle.Render("📊 Repository Overview"))
	content.WriteString("\n\n")

	content.WriteString(fmt.Sprintf("Total Contributors: %s\n",
		statsStyle.Render(fmt.Sprintf("%d", stats.TotalContributors))))
	content.WriteString(fmt.Sprintf("Total Commits: %s\n",
		statsStyle.Render(fmt.Sprintf("%d", stats.TotalCommits))))
	content.WriteString(fmt.Sprintf("Date Range: %s\n",
		statsStyle.Render(stats.DateRange)))
	content.WriteString(fmt.Sprintf("Most Active: %s\n",
		highlightStyle.Render(stats.MostActive)))

	if len(stats.RecentActivity) > 0 {
		content.WriteString("\nRecent Activity (last 30 days):\n")
		for i, activity := range stats.RecentActivity {
			if i >= 3 { // Show top 3
				break
			}
			content.WriteString(fmt.Sprintf("  %s: %s commits\n",
				activity.Name, statsStyle.Render(fmt.Sprintf("%d", activity.Count))))
		}
	}

	return content.String()
}

func (m model) renderContributorDetail() string {
	if m.selectedIndex >= len(m.contributors) {
		return "No contributor selected"
	}

	contributor := m.contributors[m.selectedIndex]
	var sections []string

	title := titleStyle.Render(fmt.Sprintf("👤 %s", contributor.Name))
	sections = append(sections, title)

	// Contributor stats
	detailContent := m.renderContributorStats(contributor)
	sections = append(sections, sectionStyle.Render(detailContent))

	// Activity patterns
	activityContent := m.renderActivityPatterns(contributor)
	sections = append(sections, sectionStyle.Render(activityContent))

	// Recent work
	recentContent := m.renderRecentWork(contributor)
	sections = append(sections, sectionStyle.Render(recentContent))

	help := helpStyle.Render("↑/↓: switch contributor • t: timeline • esc: back • q: quit")
	sections = append(sections, help)

	return strings.Join(sections, "\n")
}

func (m model) renderContributorStats(contributor ContributorData) string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("📈 Statistics"))
	content.WriteString("\n\n")

	content.WriteString(fmt.Sprintf("Email: %s\n", contributor.Email))
	content.WriteString(fmt.Sprintf("Total Commits: %s (%.1f%%)\n",
		statsStyle.Render(fmt.Sprintf("%d", contributor.TotalCommits)), contributor.Percentage))
	content.WriteString(fmt.Sprintf("Lines Added: %s\n",
		statsStyle.Render(fmt.Sprintf("%d", contributor.LinesAdded))))
	content.WriteString(fmt.Sprintf("Lines Deleted: %s\n",
		statsStyle.Render(fmt.Sprintf("%d", contributor.LinesDeleted))))
	content.WriteString(fmt.Sprintf("Files Modified: %s\n",
		statsStyle.Render(fmt.Sprintf("%d", contributor.FilesModified))))
	content.WriteString(fmt.Sprintf("Average Commit Size: %s lines\n",
		statsStyle.Render(fmt.Sprintf("%d", contributor.AverageCommitSize))))
	content.WriteString(fmt.Sprintf("First Commit: %s\n",
		contributor.FirstCommit.Format("2006-01-02")))
	content.WriteString(fmt.Sprintf("Last Commit: %s\n",
		contributor.LastCommit.Format("2006-01-02")))

	return content.String()
}

func (m model) renderActivityPatterns(contributor ContributorData) string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("📅 Activity Patterns"))
	content.WriteString("\n\n")

	// Day of week pattern
	days := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	content.WriteString("Weekly Pattern:\n")

	maxDaily := 0
	for _, count := range contributor.CommitsByDay {
		if count > maxDaily {
			maxDaily = count
		}
	}

	for i, day := range days {
		count := contributor.CommitsByDay[i]
		if maxDaily > 0 {
			bars := strings.Repeat("█", (count*10)/maxDaily+1)
			content.WriteString(fmt.Sprintf("%s %s %d\n", day, bars, count))
		}
	}

	// Most active hours
	content.WriteString("\nPeak Hours:\n")
	maxHourly := 0
	peakHour := 0
	for hour, count := range contributor.CommitsByHour {
		if count > maxHourly {
			maxHourly = count
			peakHour = hour
		}
	}
	content.WriteString(fmt.Sprintf("Most active at %s (%d commits)\n",
		statsStyle.Render(fmt.Sprintf("%02d:00", peakHour)), maxHourly))

	return content.String()
}

func (m model) renderRecentWork(contributor ContributorData) string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("🔨 Recent Work"))
	content.WriteString("\n\n")

	// Recent commits
	content.WriteString("Recent Commits:\n")
	for i, commit := range contributor.RecentCommits {
		if i >= 5 { // Show 5 most recent
			break
		}
		content.WriteString(fmt.Sprintf("%s %s (%d files)\n",
			commit.ShortHash, commit.Message, commit.FilesChanged))
	}

	// Top modified files
	if len(contributor.TopFiles) > 0 {
		content.WriteString("\nFrequently Modified Files:\n")
		for i, file := range contributor.TopFiles {
			if i >= 5 { // Show top 5
				break
			}
			content.WriteString(fmt.Sprintf("%s (%d times)\n",
				file.Path, file.Modifications))
		}
	}

	return content.String()
}

func (m model) renderTimelineView() string {
	var sections []string

	title := titleStyle.Render("📈 Activity Timeline")
	sections = append(sections, title)

	// Monthly activity for all contributors
	monthlyData := make(map[string]int)
	for _, contributor := range m.contributors {
		for month, count := range contributor.CommitsByMonth {
			monthlyData[month] += count
		}
	}

	// Sort months
	var months []string
	for month := range monthlyData {
		months = append(months, month)
	}
	sort.Strings(months)

	var content strings.Builder
	content.WriteString(headerStyle.Render("📊 Monthly Activity"))
	content.WriteString("\n\n")

	maxMonthly := 0
	for _, count := range monthlyData {
		if count > maxMonthly {
			maxMonthly = count
		}
	}

	for _, month := range months {
		count := monthlyData[month]
		if maxMonthly > 0 {
			bars := strings.Repeat("█", (count*20)/maxMonthly+1)
			content.WriteString(fmt.Sprintf("%s %s %d\n", month, bars, count))
		}
	}

	sections = append(sections, sectionStyle.Render(content.String()))

	help := helpStyle.Render("↑/↓: navigate • t: details • esc: back • q: quit")
	sections = append(sections, help)

	return strings.Join(sections, "\n")
}

func loadContributorData() tea.Msg {
	contributors, overallStats, err := analyzeContributors()
	if err != nil {
		return errMsg{err}
	}
	return dataLoadedMsg{contributors, overallStats}
}

func analyzeContributors() ([]ContributorData, OverallStats, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return nil, OverallStats{}, fmt.Errorf("failed to open repository: %w", err)
	}

	ref, err := repo.Head()
	if err != nil {
		return nil, OverallStats{}, fmt.Errorf("failed to get HEAD: %w", err)
	}

	cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, OverallStats{}, fmt.Errorf("failed to get log: %w", err)
	}

	contributorMap := make(map[string]*ContributorData)
	var totalCommits int
	var oldestCommit, newestCommit time.Time
	recentCutoff := time.Now().AddDate(0, 0, -30) // Last 30 days

	err = cIter.ForEach(func(c *object.Commit) error {
		totalCommits++
		authorName := c.Author.Name
		authorEmail := c.Author.Email
		commitTime := c.Author.When

		// Track date range
		if oldestCommit.IsZero() || commitTime.Before(oldestCommit) {
			oldestCommit = commitTime
		}
		if newestCommit.IsZero() || commitTime.After(newestCommit) {
			newestCommit = commitTime
		}

		// Get or create contributor data
		if contributorMap[authorName] == nil {
			contributorMap[authorName] = &ContributorData{
				Name:           authorName,
				Email:          authorEmail,
				CommitsByMonth: make(map[string]int),
				CommitsByHour:  make(map[int]int),
				CommitsByDay:   make(map[int]int),
				FirstCommit:    commitTime,
				LastCommit:     commitTime,
			}
		}

		contributor := contributorMap[authorName]
		contributor.TotalCommits++

		// Update date range
		if commitTime.Before(contributor.FirstCommit) {
			contributor.FirstCommit = commitTime
		}
		if commitTime.After(contributor.LastCommit) {
			contributor.LastCommit = commitTime
		}

		// Time patterns
		month := commitTime.Format("2006-01")
		contributor.CommitsByMonth[month]++
		contributor.CommitsByHour[commitTime.Hour()]++
		contributor.CommitsByDay[int(commitTime.Weekday())]++

		// Get commit stats
		stats, err := c.Stats()
		if err == nil {
			additions := 0
			deletions := 0
			filesModified := len(stats)

			for _, stat := range stats {
				additions += stat.Addition
				deletions += stat.Deletion
			}

			contributor.LinesAdded += additions
			contributor.LinesDeleted += deletions
			contributor.FilesModified += filesModified

			// Track recent commits
			if commitTime.After(recentCutoff) {
				contributor.RecentCommits = append(contributor.RecentCommits, CommitSummary{
					Hash:         c.Hash.String(),
					ShortHash:    c.Hash.String()[:8],
					Message:      strings.Split(c.Message, "\n")[0],
					Date:         commitTime,
					FilesChanged: filesModified,
					Additions:    additions,
					Deletions:    deletions,
				})
			}

			// Track largest commit
			totalChanges := additions + deletions
			if totalChanges > contributor.LargestCommit.Additions+contributor.LargestCommit.Deletions {
				contributor.LargestCommit = CommitSummary{
					Hash:         c.Hash.String(),
					ShortHash:    c.Hash.String()[:8],
					Message:      strings.Split(c.Message, "\n")[0],
					Date:         commitTime,
					FilesChanged: filesModified,
					Additions:    additions,
					Deletions:    deletions,
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, OverallStats{}, fmt.Errorf("failed to iterate commits: %w", err)
	}

	// Convert map to slice and calculate percentages
	var contributors []ContributorData
	for _, contributor := range contributorMap {
		contributor.Percentage = float64(contributor.TotalCommits) / float64(totalCommits) * 100
		if contributor.TotalCommits > 0 {
			contributor.AverageCommitSize = (contributor.LinesAdded + contributor.LinesDeleted) / contributor.TotalCommits
		}

		// Sort recent commits by date
		sort.Slice(contributor.RecentCommits, func(i, j int) bool {
			return contributor.RecentCommits[i].Date.After(contributor.RecentCommits[j].Date)
		})

		contributors = append(contributors, *contributor)
	}

	// Sort contributors by commit count
	sort.Slice(contributors, func(i, j int) bool {
		return contributors[i].TotalCommits > contributors[j].TotalCommits
	})

	// Calculate overall stats
	var mostActive string
	if len(contributors) > 0 {
		mostActive = contributors[0].Name
	}

	// Recent activity
	var recentActivity []ContributorActivity
	for _, contributor := range contributors {
		recentCount := len(contributor.RecentCommits)
		if recentCount > 0 {
			recentActivity = append(recentActivity, ContributorActivity{
				Name:   contributor.Name,
				Period: "30 days",
				Count:  recentCount,
			})
		}
	}

	// Sort recent activity
	sort.Slice(recentActivity, func(i, j int) bool {
		return recentActivity[i].Count > recentActivity[j].Count
	})

	overallStats := OverallStats{
		TotalContributors: len(contributors),
		TotalCommits:      totalCommits,
		DateRange:         fmt.Sprintf("%s to %s", oldestCommit.Format("2006-01-02"), newestCommit.Format("2006-01-02")),
		MostActive:        mostActive,
		RecentActivity:    recentActivity,
	}

	return contributors, overallStats, nil
}

// RunContributorsAnalysis starts the contributors analysis TUI
func RunContributorsAnalysis() error {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#01FAC6")).
		BorderLeftForeground(lipgloss.Color("#01FAC6"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#DDDDDD"))

	contributorList := list.New([]list.Item{}, delegate, 0, 0)
	contributorList.Title = "Contributors"
	contributorList.SetShowStatusBar(false)
	contributorList.SetShowHelp(false)

	m := model{
		contributorList: contributorList,
		viewMode:        ContributorListView,
		loading:         true,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
