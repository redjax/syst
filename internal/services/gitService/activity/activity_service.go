package activity

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	gitservice "github.com/redjax/syst/internal/services/gitService"
)

type ViewMode int

const (
	OverviewView ViewMode = iota
	TimingView
	PatternsView
	ContributorsView
	TrendsView
)

type ActivityData struct {
	TotalCommits    int
	CommitsByHour   map[int]int    // hour -> count
	CommitsByDay    map[int]int    // weekday -> count
	CommitsByMonth  map[string]int // month -> count
	RecentActivity  []CommitActivity
	TopAuthors      []AuthorStats
	CommitFrequency map[string]int // date -> count
	AveragePerDay   float64
	MostActiveDay   string
	MostActiveHour  int
	LongestStreak   int
	CurrentStreak   int
	MonthlyTrends   []MonthlyTrend
	WeeklyActivity  []WeeklyActivity
	HourlyDistrib   []HourlyActivity
	AuthorTimeline  []AuthorActivity
}

type CommitActivity struct {
	Date   string
	Count  int
	Author string
}

type AuthorStats struct {
	Name        string
	Commits     int
	Percentage  float64
	FirstCommit string
	LastCommit  string
	AvgPerWeek  float64
}

type MonthlyTrend struct {
	Month  string
	Count  int
	Change float64 // percentage change from previous month
}

type WeeklyActivity struct {
	Week    string
	Count   int
	Authors []string
}

type HourlyActivity struct {
	Hour  int
	Count int
	Peak  bool
}

type AuthorActivity struct {
	Author string
	Week   string
	Count  int
}

type model struct {
	data             ActivityData
	currentView      ViewMode
	contributorIndex int
	err              error
	loading          bool
	tuiHelper        *gitservice.ResponsiveTUIHelper
}

type dataLoadedMsg struct {
	data ActivityData
}

type errMsg struct {
	err error
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
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

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87"))
)

// Helper function to get dynamic section style based on terminal width
func (m model) getSectionStyle() lipgloss.Style {
	return m.tuiHelper.GetResponsiveSectionStyle(sectionStyle)
}

// Helper function to get dynamic title style based on terminal width
func (m model) getTitleStyle() lipgloss.Style {
	return m.tuiHelper.GetResponsiveTitleStyle(titleStyle)
}

func (m model) Init() tea.Cmd {
	return loadActivityData
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.tuiHelper.HandleWindowSizeMsg(msg)
		return m, nil

	case dataLoadedMsg:
		m.data = msg.data
		m.loading = false
		return m, nil

	case errMsg:
		m.err = msg.err
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))):
			return m, tea.Quit
		case key.Matches(msg, key.NewBinding(key.WithKeys("1"))):
			m.currentView = OverviewView
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("2"))):
			m.currentView = TimingView
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("3"))):
			m.currentView = PatternsView
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("4"))):
			m.currentView = ContributorsView
			m.contributorIndex = 0
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("5"))):
			m.currentView = TrendsView
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("left", "h"))):
			if m.currentView > 0 {
				m.currentView--
				if m.currentView == ContributorsView {
					m.contributorIndex = 0
				}
			}
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("right", "l"))):
			if m.currentView < TrendsView {
				m.currentView++
				if m.currentView == ContributorsView {
					m.contributorIndex = 0
				}
			}
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			if m.currentView == ContributorsView && len(m.data.TopAuthors) > 0 {
				if m.contributorIndex > 0 {
					m.contributorIndex--
				}
			}
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			if m.currentView == ContributorsView && len(m.data.TopAuthors) > 0 {
				if m.contributorIndex < len(m.data.TopAuthors)-1 {
					m.contributorIndex++
				}
			}
			return m, nil
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.loading {
		loadingMsg := "Loading repository activity data..."
		return m.tuiHelper.CenterContent(loadingMsg)
	}

	if m.err != nil {
		errorMsg := fmt.Sprintf("Error: %v", m.err)
		return m.tuiHelper.CenterContent(errorStyle.Render(errorMsg))
	}

	var content strings.Builder

	// Title with current view indicator
	viewNames := []string{"Overview", "Timing", "Patterns", "Contributors", "Trends"}
	title := fmt.Sprintf("📊 Repository Activity Dashboard - %s", viewNames[m.currentView])
	content.WriteString(m.getTitleStyle().Render(title))
	content.WriteString("\n\n")

	// Render current view
	switch m.currentView {
	case OverviewView:
		content.WriteString(m.renderOverviewView())
	case TimingView:
		content.WriteString(m.renderTimingView())
	case PatternsView:
		content.WriteString(m.renderPatternsView())
	case ContributorsView:
		content.WriteString(m.renderContributorsView())
	case TrendsView:
		content.WriteString(m.renderTrendsView())
	}

	// Navigation help at the bottom
	width, _ := m.tuiHelper.GetSize()
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Width(width).
		Align(lipgloss.Center).
		Render("1: Overview • 2: Timing • 3: Patterns • 4: Contributors • 5: Trends • ←/→: Navigate • q: Quit")

	content.WriteString("\n")
	content.WriteString(help)

	// Ensure content fits within terminal height
	result := content.String()
	return m.tuiHelper.TruncateContentToHeight(result)
}

func (m model) renderOverviewView() string {
	d := m.data
	var content strings.Builder

	// Calculate available width for content layout
	width, height := m.tuiHelper.GetSize()

	// Use responsive section style
	sectionStyleResponsive := m.getSectionStyle()

	content.WriteString(sectionStyleResponsive.Render(headerStyle.Render("📈 Overview Statistics")))
	content.WriteString("\n\n")

	// Create two-column layout for larger terminals
	if width >= 80 {
		leftCol := []string{
			fmt.Sprintf("Total Commits: %s", statsStyle.Render(fmt.Sprintf("%d", d.TotalCommits))),
			fmt.Sprintf("Average per Day: %s", statsStyle.Render(fmt.Sprintf("%.1f", d.AveragePerDay))),
			fmt.Sprintf("Current Streak: %s days", statsStyle.Render(fmt.Sprintf("%d", d.CurrentStreak))),
		}

		rightCol := []string{
			fmt.Sprintf("Longest Streak: %s days", statsStyle.Render(fmt.Sprintf("%d", d.LongestStreak))),
			fmt.Sprintf("Most Active Day: %s", statsStyle.Render(d.MostActiveDay)),
			fmt.Sprintf("Most Active Hour: %s", statsStyle.Render(fmt.Sprintf("%d:00", d.MostActiveHour))),
		}

		content.WriteString(m.tuiHelper.CreateTwoColumnLayout(leftCol, rightCol))
	} else {
		// Single column for smaller terminals
		stats := []string{
			fmt.Sprintf("Total Commits: %s", statsStyle.Render(fmt.Sprintf("%d", d.TotalCommits))),
			fmt.Sprintf("Average per Day: %s", statsStyle.Render(fmt.Sprintf("%.1f", d.AveragePerDay))),
			fmt.Sprintf("Current Streak: %s days", statsStyle.Render(fmt.Sprintf("%d", d.CurrentStreak))),
			fmt.Sprintf("Longest Streak: %s days", statsStyle.Render(fmt.Sprintf("%d", d.LongestStreak))),
			fmt.Sprintf("Most Active Day: %s", statsStyle.Render(d.MostActiveDay)),
			fmt.Sprintf("Most Active Hour: %s", statsStyle.Render(fmt.Sprintf("%d:00", d.MostActiveHour))),
		}

		for _, stat := range stats {
			content.WriteString(stat + "\n")
		}
	}

	content.WriteString("\n")
	content.WriteString(sectionStyleResponsive.Render(headerStyle.Render("📅 Recent Activity")))
	content.WriteString("\n\n")

	if len(d.RecentActivity) > 0 {
		// Adjust number of shown items based on available height
		maxItems := 7
		if height < 20 {
			maxItems = 3
		} else if height < 30 {
			maxItems = 5
		}

		for i, activity := range d.RecentActivity {
			if i >= maxItems {
				break
			}
			content.WriteString(fmt.Sprintf("%s: %s commits\n",
				activity.Date, statsStyle.Render(fmt.Sprintf("%d", activity.Count))))
		}
	} else {
		content.WriteString("No recent activity found\n")
	}

	return content.String()
}

func (m model) renderTimingView() string {
	d := m.data
	var content strings.Builder

	// Use responsive section style
	sectionStyleResponsive := m.getSectionStyle()

	content.WriteString(sectionStyleResponsive.Render(headerStyle.Render("⏰ Hourly Distribution")))
	content.WriteString("\n\n")

	// Hour distribution with enhanced visualization
	maxHourly := 0
	for _, count := range d.CommitsByHour {
		if count > maxHourly {
			maxHourly = count
		}
	}

	if maxHourly > 0 {
		// Calculate bar length based on terminal width
		maxBarLength := m.tuiHelper.CalculateBarLength(30, 40) // 30 for labels, max 40 for bars

		for hour := 0; hour < 24; hour++ {
			count := d.CommitsByHour[hour]
			if count > 0 {
				percentage := float64(count) / float64(maxHourly)
				barLength := int(percentage * float64(maxBarLength))
				bars := strings.Repeat("█", barLength)
				if barLength == 0 && count > 0 {
					bars = "▏"
				}

				timeRange := fmt.Sprintf("%02d:00-%02d:59", hour, hour)
				content.WriteString(fmt.Sprintf("%-11s %s %s (%d)\n",
					timeRange, bars, statsStyle.Render(fmt.Sprintf("%.1f%%", percentage*100)), count))
			}
		}
	}

	content.WriteString("\n")
	content.WriteString(sectionStyleResponsive.Render(headerStyle.Render("📅 Weekly Distribution")))
	content.WriteString("\n\n")

	// Day of week distribution
	days := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	maxDaily := 0
	for _, count := range d.CommitsByDay {
		if count > maxDaily {
			maxDaily = count
		}
	}

	if maxDaily > 0 {
		// Calculate bar length based on terminal width
		maxBarLength := m.tuiHelper.CalculateBarLength(25, 30) // 25 for labels, max 30 for bars

		for i, day := range days {
			count := d.CommitsByDay[i]
			if count > 0 {
				percentage := float64(count) / float64(maxDaily)
				barLength := int(percentage * float64(maxBarLength))
				bars := strings.Repeat("█", barLength)
				if barLength == 0 && count > 0 {
					bars = "▏"
				}
				content.WriteString(fmt.Sprintf("%-10s %s %s (%d)\n",
					day, bars, statsStyle.Render(fmt.Sprintf("%.1f%%", percentage*100)), count))
			}
		}
	}

	return content.String()
}

func (m model) renderPatternsView() string {
	d := m.data
	var content strings.Builder

	// Use responsive section style
	sectionStyleResponsive := m.getSectionStyle()
	_, height := m.tuiHelper.GetSize()

	content.WriteString(sectionStyleResponsive.Render(headerStyle.Render("📅 Monthly Trends")))
	content.WriteString("\n\n")

	if len(d.MonthlyTrends) > 0 {
		// Adjust number of months shown based on terminal height
		maxMonths := 12
		if height < 25 {
			maxMonths = 6
		} else if height < 35 {
			maxMonths = 9
		}

		for i, trend := range d.MonthlyTrends {
			if i >= maxMonths {
				break
			}
			changeIndicator := ""
			if trend.Change > 0 {
				changeIndicator = fmt.Sprintf(" (+%.1f%%)", trend.Change)
			} else if trend.Change < 0 {
				changeIndicator = fmt.Sprintf(" (%.1f%%)", trend.Change)
			}

			content.WriteString(fmt.Sprintf("%s: %s commits%s\n",
				trend.Month, statsStyle.Render(fmt.Sprintf("%d", trend.Count)), changeIndicator))
		}
	} else {
		content.WriteString("Calculating monthly trends...\n")
	}

	content.WriteString("\n")
	content.WriteString(sectionStyleResponsive.Render(headerStyle.Render("📈 Activity Patterns")))
	content.WriteString("\n\n")

	// Peak hours analysis
	peakHours := []int{}
	avgHourly := float64(d.TotalCommits) / 24.0
	for hour, count := range d.CommitsByHour {
		if float64(count) > avgHourly*1.5 {
			peakHours = append(peakHours, hour)
		}
	}

	if len(peakHours) > 0 {
		content.WriteString("Peak Development Hours: ")
		for i, hour := range peakHours {
			if i > 0 {
				content.WriteString(", ")
			}
			content.WriteString(statsStyle.Render(fmt.Sprintf("%d:00", hour)))
		}
		content.WriteString("\n")
	}

	// Workday vs weekend analysis
	weekdayCommits := d.CommitsByDay[1] + d.CommitsByDay[2] + d.CommitsByDay[3] + d.CommitsByDay[4] + d.CommitsByDay[5]
	weekendCommits := d.CommitsByDay[0] + d.CommitsByDay[6]

	if weekdayCommits > 0 || weekendCommits > 0 {
		weekdayPct := float64(weekdayCommits) / float64(weekdayCommits+weekendCommits) * 100
		weekendPct := float64(weekendCommits) / float64(weekdayCommits+weekendCommits) * 100

		content.WriteString(fmt.Sprintf("Weekday Activity: %s (%.1f%%)\n",
			statsStyle.Render(fmt.Sprintf("%d", weekdayCommits)), weekdayPct))
		content.WriteString(fmt.Sprintf("Weekend Activity: %s (%.1f%%)\n",
			statsStyle.Render(fmt.Sprintf("%d", weekendCommits)), weekendPct))
	}

	return content.String()
}

func (m model) renderContributorsView() string {
	var content strings.Builder

	// Use responsive section style
	sectionStyleResponsive := m.getSectionStyle()

	content.WriteString(sectionStyleResponsive.Render(headerStyle.Render("👥 Contributors Analysis")))
	content.WriteString("\n\n")

	if len(m.data.TopAuthors) > 0 {
		content.WriteString("Navigate with ↑/↓ keys\n\n")

		// Calculate how many contributors to show based on terminal height
		maxContributors := m.tuiHelper.CalculateMaxItemsForHeight(5, 10) // 5 lines per contributor, 10 reserved lines

		if maxContributors > len(m.data.TopAuthors) {
			maxContributors = len(m.data.TopAuthors)
		}

		for i, author := range m.data.TopAuthors {
			if i >= maxContributors {
				break
			}

			prefix := "  "
			if i == m.contributorIndex {
				prefix = "▶ "
			}

			content.WriteString(fmt.Sprintf("%s%s\n", prefix, statsStyle.Render(author.Name)))
			content.WriteString(fmt.Sprintf("    %d commits (%.1f%%)\n", author.Commits, author.Percentage))
			content.WriteString(fmt.Sprintf("    Avg: %.1f commits/week\n", author.AvgPerWeek))
			content.WriteString(fmt.Sprintf("    Active: %s to %s\n", author.FirstCommit, author.LastCommit))
			content.WriteString("\n")
		}

		if len(m.data.TopAuthors) > maxContributors {
			content.WriteString(fmt.Sprintf("... and %d more contributors\n", len(m.data.TopAuthors)-maxContributors))
		}
	} else {
		content.WriteString("No contributor data available\n")
	}

	return content.String()
}

func (m model) renderTrendsView() string {
	d := m.data
	var content strings.Builder

	// Use responsive section style
	sectionStyleResponsive := m.getSectionStyle()
	_, height := m.tuiHelper.GetSize()

	content.WriteString(sectionStyleResponsive.Render(headerStyle.Render("📈 Long-term Trends")))
	content.WriteString("\n\n")

	// Commit frequency analysis
	if len(d.CommitFrequency) > 0 {
		totalDays := len(d.CommitFrequency)
		activeDays := 0
		for _, count := range d.CommitFrequency {
			if count > 0 {
				activeDays++
			}
		}

		activityRate := float64(activeDays) / float64(totalDays) * 100
		content.WriteString(fmt.Sprintf("Activity Rate: %s of tracked days\n",
			statsStyle.Render(fmt.Sprintf("%.1f%%", activityRate))))
		content.WriteString(fmt.Sprintf("Total Active Days: %s\n",
			statsStyle.Render(fmt.Sprintf("%d", activeDays))))
	}

	content.WriteString("\n")
	content.WriteString(sectionStyleResponsive.Render(headerStyle.Render("🏆 Repository Milestones")))
	content.WriteString("\n\n")

	// Milestone analysis
	milestones := []string{
		fmt.Sprintf("🎯 Reached %d total commits", d.TotalCommits),
	}

	if d.LongestStreak >= 7 {
		milestones = append(milestones, fmt.Sprintf("🔥 Achieved %d-day coding streak", d.LongestStreak))
	}

	if len(d.TopAuthors) > 1 {
		milestones = append(milestones, fmt.Sprintf("👥 %d contributors collaboration", len(d.TopAuthors)))
	}

	if d.AveragePerDay >= 1.0 {
		milestones = append(milestones, fmt.Sprintf("⚡ Sustained %.1f commits/day average", d.AveragePerDay))
	}

	// Limit milestones shown based on terminal height
	maxMilestones := len(milestones)
	if height < 20 {
		maxMilestones = 2
	} else if height < 30 {
		maxMilestones = 3
	}

	for i, milestone := range milestones {
		if i >= maxMilestones {
			break
		}
		content.WriteString(milestone + "\n")
	}

	if len(milestones) > maxMilestones {
		content.WriteString(fmt.Sprintf("... and %d more milestones\n", len(milestones)-maxMilestones))
	}

	content.WriteString("\n")
	content.WriteString(sectionStyleResponsive.Render(headerStyle.Render("📊 Activity Summary")))
	content.WriteString("\n\n")

	// Summary insights
	busyHours := 0
	for _, count := range d.CommitsByHour {
		if count > 0 {
			busyHours++
		}
	}

	busyDays := 0
	for _, count := range d.CommitsByDay {
		if count > 0 {
			busyDays++
		}
	}

	content.WriteString(fmt.Sprintf("Active Hours: %s/24\n", statsStyle.Render(fmt.Sprintf("%d", busyHours))))
	content.WriteString(fmt.Sprintf("Active Days: %s/7\n", statsStyle.Render(fmt.Sprintf("%d", busyDays))))

	if len(d.TopAuthors) > 0 {
		topContributor := d.TopAuthors[0]
		content.WriteString(fmt.Sprintf("Top Contributor: %s (%s commits)\n",
			statsStyle.Render(topContributor.Name),
			statsStyle.Render(fmt.Sprintf("%d", topContributor.Commits))))
	}

	return content.String()
}

func loadActivityData() tea.Msg {
	data, err := gatherActivityData()
	if err != nil {
		return errMsg{err}
	}
	return dataLoadedMsg{data}
}

func gatherActivityData() (ActivityData, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return ActivityData{}, fmt.Errorf("failed to open repository: %w", err)
	}

	ref, err := repo.Head()
	if err != nil {
		return ActivityData{}, fmt.Errorf("failed to get HEAD: %w", err)
	}

	cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return ActivityData{}, fmt.Errorf("failed to get log: %w", err)
	}

	data := ActivityData{
		CommitsByHour:   make(map[int]int),
		CommitsByDay:    make(map[int]int),
		CommitsByMonth:  make(map[string]int),
		CommitFrequency: make(map[string]int),
	}

	authorStats := make(map[string]int)
	authorFirstCommit := make(map[string]time.Time)
	authorLastCommit := make(map[string]time.Time)
	commitDates := []time.Time{}
	recentDates := make(map[string]int)

	err = cIter.ForEach(func(c *object.Commit) error {
		data.TotalCommits++

		// Time analysis
		commitTime := c.Author.When
		commitDates = append(commitDates, commitTime)

		hour := commitTime.Hour()
		data.CommitsByHour[hour]++

		day := int(commitTime.Weekday())
		data.CommitsByDay[day]++

		month := commitTime.Format("2006-01")
		data.CommitsByMonth[month]++

		dateStr := commitTime.Format("2006-01-02")
		data.CommitFrequency[dateStr]++

		// Recent activity (last 30 days)
		if time.Since(commitTime) < 30*24*time.Hour {
			recentDates[dateStr]++
		}

		// Author stats with timeline
		authorName := c.Author.Name
		authorStats[authorName]++

		if _, exists := authorFirstCommit[authorName]; !exists {
			authorFirstCommit[authorName] = commitTime
			authorLastCommit[authorName] = commitTime
		} else {
			if commitTime.Before(authorFirstCommit[authorName]) {
				authorFirstCommit[authorName] = commitTime
			}
			if commitTime.After(authorLastCommit[authorName]) {
				authorLastCommit[authorName] = commitTime
			}
		}

		return nil
	})

	if err != nil {
		return ActivityData{}, fmt.Errorf("failed to iterate commits: %w", err)
	}

	// Calculate derived stats
	data.AveragePerDay = calculateAveragePerDay(commitDates)
	data.MostActiveDay = findMostActiveDay(data.CommitsByDay)
	data.MostActiveHour = findMostActiveHour(data.CommitsByHour)
	data.LongestStreak, data.CurrentStreak = calculateStreaks(data.CommitFrequency)
	data.TopAuthors = calculateTopAuthors(authorStats, data.TotalCommits, authorFirstCommit, authorLastCommit)
	data.RecentActivity = formatRecentActivity(recentDates)
	data.MonthlyTrends = calculateMonthlyTrends(data.CommitsByMonth)
	data.WeeklyActivity = calculateWeeklyActivity(data.CommitFrequency, commitDates)
	data.HourlyDistrib = calculateHourlyDistribution(data.CommitsByHour)

	return data, nil
}

func calculateAveragePerDay(commitDates []time.Time) float64 {
	if len(commitDates) == 0 {
		return 0
	}

	sort.Slice(commitDates, func(i, j int) bool {
		return commitDates[i].Before(commitDates[j])
	})

	oldest := commitDates[0]
	newest := commitDates[len(commitDates)-1]
	daysDiff := newest.Sub(oldest).Hours() / 24

	if daysDiff == 0 {
		return float64(len(commitDates))
	}

	return float64(len(commitDates)) / daysDiff
}

func findMostActiveDay(commitsByDay map[int]int) string {
	days := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	maxDay := 0
	maxCount := 0

	for day, count := range commitsByDay {
		if count > maxCount {
			maxCount = count
			maxDay = day
		}
	}

	return days[maxDay]
}

func findMostActiveHour(commitsByHour map[int]int) int {
	maxHour := 0
	maxCount := 0

	for hour, count := range commitsByHour {
		if count > maxCount {
			maxCount = count
			maxHour = hour
		}
	}

	return maxHour
}

func calculateStreaks(commitFrequency map[string]int) (int, int) {
	if len(commitFrequency) == 0 {
		return 0, 0
	}

	// Get all dates and sort them
	var dates []time.Time
	for dateStr := range commitFrequency {
		date, err := time.Parse("2006-01-02", dateStr)
		if err == nil {
			dates = append(dates, date)
		}
	}

	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	currentStreak := 1
	longestStreak := 1

	for i := 1; i < len(dates); i++ {
		daysDiff := dates[i].Sub(dates[i-1]).Hours() / 24
		if daysDiff == 1 {
			currentStreak++
			if currentStreak > longestStreak {
				longestStreak = currentStreak
			}
		} else {
			currentStreak = 1
		}
	}

	// Calculate current streak from today
	currentStreakFromToday := 0
	currentDate := time.Now()

	for {
		dateStr := currentDate.Format("2006-01-02")
		if _, exists := commitFrequency[dateStr]; exists {
			currentStreakFromToday++
			currentDate = currentDate.AddDate(0, 0, -1)
		} else {
			break
		}
	}

	return longestStreak, currentStreakFromToday
}

func calculateTopAuthors(authorStats map[string]int, totalCommits int, firstCommits, lastCommits map[string]time.Time) []AuthorStats {
	var authors []AuthorStats

	for name, commits := range authorStats {
		percentage := float64(commits) / float64(totalCommits) * 100

		firstCommit := "N/A"
		lastCommit := "N/A"
		avgPerWeek := 0.0

		if first, exists := firstCommits[name]; exists {
			firstCommit = first.Format("2006-01-02")
		}
		if last, exists := lastCommits[name]; exists {
			lastCommit = last.Format("2006-01-02")

			// Calculate average per week
			if first, exists := firstCommits[name]; exists {
				weeksDiff := last.Sub(first).Hours() / (24 * 7)
				if weeksDiff > 0 {
					avgPerWeek = float64(commits) / weeksDiff
				} else {
					avgPerWeek = float64(commits) // If same day, treat as one week
				}
			}
		}

		authors = append(authors, AuthorStats{
			Name:        name,
			Commits:     commits,
			Percentage:  percentage,
			FirstCommit: firstCommit,
			LastCommit:  lastCommit,
			AvgPerWeek:  avgPerWeek,
		})
	}

	sort.Slice(authors, func(i, j int) bool {
		return authors[i].Commits > authors[j].Commits
	})

	return authors
}

func calculateMonthlyTrends(commitsByMonth map[string]int) []MonthlyTrend {
	var trends []MonthlyTrend
	var months []string

	for month := range commitsByMonth {
		months = append(months, month)
	}

	sort.Strings(months)

	for i, month := range months {
		count := commitsByMonth[month]
		change := 0.0

		if i > 0 {
			prevCount := commitsByMonth[months[i-1]]
			if prevCount > 0 {
				change = (float64(count) - float64(prevCount)) / float64(prevCount) * 100
			}
		}

		trends = append(trends, MonthlyTrend{
			Month:  month,
			Count:  count,
			Change: change,
		})
	}

	// Return last 12 months
	if len(trends) > 12 {
		trends = trends[len(trends)-12:]
	}

	return trends
}

func calculateWeeklyActivity(commitFrequency map[string]int, commitDates []time.Time) []WeeklyActivity {
	var activity []WeeklyActivity

	// Group by week (simplified - could be more sophisticated)
	weeklyData := make(map[string]int)

	for _, date := range commitDates {
		year, week := date.ISOWeek()
		weekKey := fmt.Sprintf("%d-W%02d", year, week)
		weeklyData[weekKey]++
	}

	for week, count := range weeklyData {
		activity = append(activity, WeeklyActivity{
			Week:    week,
			Count:   count,
			Authors: []string{}, // Could be enhanced to track authors per week
		})
	}

	sort.Slice(activity, func(i, j int) bool {
		return activity[i].Week > activity[j].Week
	})

	return activity
}

func calculateHourlyDistribution(commitsByHour map[int]int) []HourlyActivity {
	var distribution []HourlyActivity

	// Find average
	total := 0
	count := 0
	for _, commits := range commitsByHour {
		if commits > 0 {
			total += commits
			count++
		}
	}

	average := 0.0
	if count > 0 {
		average = float64(total) / float64(count)
	}

	for hour := 0; hour < 24; hour++ {
		commits := commitsByHour[hour]
		isPeak := float64(commits) > average*1.5

		distribution = append(distribution, HourlyActivity{
			Hour:  hour,
			Count: commits,
			Peak:  isPeak,
		})
	}

	return distribution
}

func formatRecentActivity(recentDates map[string]int) []CommitActivity {
	var activity []CommitActivity

	for dateStr, count := range recentDates {
		activity = append(activity, CommitActivity{
			Date:  dateStr,
			Count: count,
		})
	}

	sort.Slice(activity, func(i, j int) bool {
		return activity[i].Date > activity[j].Date
	})

	// Limit to 10 most recent
	if len(activity) > 10 {
		activity = activity[:10]
	}

	return activity
}

// RunActivityDashboard starts the repository activity dashboard TUI
func RunActivityDashboard() error {
	m := model{
		loading:   true,
		tuiHelper: gitservice.NewResponsiveTUIHelper(),
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
