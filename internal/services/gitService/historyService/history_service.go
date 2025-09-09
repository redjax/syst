package historyService

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/redjax/syst/internal/utils/terminal"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type ViewMode int

const (
	TimelineView ViewMode = iota
	FrequencyView
	TagsView
	MergesView
)

type HistoryAnalysis struct {
	Timeline      []TimelineCommit
	FrequencyData FrequencyData
	Tags          []TagInfo
	Merges        []MergeCommit
	OverallStats  OverallHistoryStats
}

type TimelineCommit struct {
	Hash        string
	ShortHash   string
	Message     string
	Author      string
	Email       string
	Date        time.Time
	ParentCount int
	IsMerge     bool
	Branch      string
	Files       []string
	Additions   int
	Deletions   int
}

type FrequencyData struct {
	CommitsByDate    map[string]int // date -> count
	CommitsByMonth   map[string]int // month -> count
	CommitsByWeekday map[int]int    // weekday -> count
	CommitsByHour    map[int]int    // hour -> count
	CommitsByAuthor  map[string]int // author -> count
	HeatmapData      [][]int        // week x day grid for heatmap
	HeatmapWeeks     []string       // week labels
	MaxCommitsPerDay int
	TotalDays        int
	CommitStreak     StreakInfo
}

type StreakInfo struct {
	Current    int
	Longest    int
	CurrentEnd time.Time
	LongestEnd time.Time
}

type TagInfo struct {
	Name         string
	Hash         string
	Date         time.Time
	Tagger       string
	Message      string
	CommitsSince int
	Type         string // "annotated" or "lightweight"
}

type MergeCommit struct {
	Hash         string
	ShortHash    string
	Message      string
	Author       string
	Date         time.Time
	ParentHashes []string
	BranchMerged string
	FilesChanged int
	Additions    int
	Deletions    int
}

type OverallHistoryStats struct {
	TotalCommits     int
	FirstCommit      time.Time
	LastCommit       time.Time
	ActiveDays       int
	TotalAuthors     int
	AveragePerDay    float64
	MostActiveDay    string
	MostActiveAuthor string
	TotalTags        int
	TotalMerges      int
}

type model struct {
	analysis     HistoryAnalysis
	currentView  ViewMode
	timelineList list.Model
	tagsList     list.Model
	mergesList   list.Model
	loading      bool
	err          error
	tuiHelper *terminal.ResponsiveTUIHelper
	sections     []string
}

type timelineItem struct {
	commit TimelineCommit
}

func (i timelineItem) FilterValue() string { return i.commit.Message }
func (i timelineItem) Title() string {
	prefix := "📝"
	if i.commit.IsMerge {
		prefix = "🔀"
	}
	return fmt.Sprintf("%s %s %s", prefix, i.commit.ShortHash, i.commit.Message)
}
func (i timelineItem) Description() string {
	return fmt.Sprintf("%s • %s • %d files",
		i.commit.Author, i.commit.Date.Format("2006-01-02 15:04"), len(i.commit.Files))
}

type tagItem struct {
	tag TagInfo
}

func (i tagItem) FilterValue() string { return i.tag.Name }
func (i tagItem) Title() string {
	prefix := "🏷️"
	if i.tag.Type == "annotated" {
		prefix = "📋"
	}
	return fmt.Sprintf("%s %s", prefix, i.tag.Name)
}
func (i tagItem) Description() string {
	return fmt.Sprintf("%s • %s • %d commits since",
		i.tag.Tagger, i.tag.Date.Format("2006-01-02"), i.tag.CommitsSince)
}

type mergeItem struct {
	merge MergeCommit
}

func (i mergeItem) FilterValue() string { return i.merge.Message }
func (i mergeItem) Title() string {
	return fmt.Sprintf("🔀 %s %s", i.merge.ShortHash, i.merge.Message)
}
func (i mergeItem) Description() string {
	return fmt.Sprintf("%s • %s • %d files • +%d -%d",
		i.merge.Author, i.merge.Date.Format("2006-01-02 15:04"),
		i.merge.FilesChanged, i.merge.Additions, i.merge.Deletions)
}

type dataLoadedMsg struct {
	analysis HistoryAnalysis
}

type errMsg struct {
	err error
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#9B59B6")).
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
	return loadHistoryData
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.tuiHelper.HandleWindowSizeMsg(msg)
		m.timelineList.SetWidth(m.tuiHelper.GetWidth())
		m.timelineList.SetHeight(m.tuiHelper.GetHeight() - 12)
		m.tagsList.SetWidth(m.tuiHelper.GetWidth())
		m.tagsList.SetHeight(m.tuiHelper.GetHeight() - 12)
		m.mergesList.SetWidth(m.tuiHelper.GetWidth())
		m.mergesList.SetHeight(m.tuiHelper.GetHeight() - 12)
		return m, nil

	case dataLoadedMsg:
		m.analysis = msg.analysis
		m.loading = false
		m.sections = []string{
			"Timeline",
			"Frequency",
			"Tags",
			"Merges",
		}
		m.updateListItems()
		return m, nil

	case errMsg:
		m.err = msg.err
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c", "esc"))):
			return m, tea.Quit
		case key.Matches(msg, key.NewBinding(key.WithKeys("1"))):
			m.currentView = TimelineView
			m.updateListItems()
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("2"))):
			m.currentView = FrequencyView
			m.updateListItems()
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("3"))):
			m.currentView = TagsView
			m.updateListItems()
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("4"))):
			m.currentView = MergesView
			m.updateListItems()
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("left", "h"))):
			if m.currentView > 0 {
				m.currentView--
				m.updateListItems()
			}
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("right", "l"))):
			if int(m.currentView) < len(m.sections)-1 {
				m.currentView++
				m.updateListItems()
			}
			return m, nil
		default:
			// Pass to the appropriate list
			var cmd tea.Cmd
			switch m.currentView {
			case TimelineView:
				m.timelineList, cmd = m.timelineList.Update(msg)
			case TagsView:
				m.tagsList, cmd = m.tagsList.Update(msg)
			case MergesView:
				m.mergesList, cmd = m.mergesList.Update(msg)
			}
			return m, cmd
		}
	}

	return m, nil
}

func (m *model) updateListItems() {
	switch m.currentView {
	case TimelineView:
		var items []list.Item
		for _, commit := range m.analysis.Timeline {
			items = append(items, timelineItem{commit: commit})
		}
		m.timelineList.SetItems(items)
	case TagsView:
		var items []list.Item
		for _, tag := range m.analysis.Tags {
			items = append(items, tagItem{tag: tag})
		}
		m.tagsList.SetItems(items)
	case MergesView:
		var items []list.Item
		for _, merge := range m.analysis.Merges {
			items = append(items, mergeItem{merge: merge})
		}
		m.mergesList.SetItems(items)
	}
}

func (m model) View() string {
	if m.loading {
		return "\n  Analyzing repository history...\n"
	}

	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("\n  Error: %v\n", m.err))
	}

	var sections []string

	// Title
	title := titleStyle.Render("📈 Git History Explorer")
	sections = append(sections, title)

	// Navigation tabs
	tabs := m.renderTabs()
	sections = append(sections, tabs)

	// Content based on current view
	content := m.renderCurrentView()
	sections = append(sections, sectionStyle.Render(content))

	// Instructions
	help := helpStyle.Render("1-4: sections • ←/→: navigate • ↑/↓: scroll • q: quit")
	sections = append(sections, help)

	return strings.Join(sections, "\n")
}

func (m model) renderTabs() string {
	var tabs []string

	for i, section := range m.sections {
		style := lipgloss.NewStyle().Padding(0, 1)
		if ViewMode(i) == m.currentView {
			style = style.Foreground(lipgloss.Color("#01FAC6")).Bold(true).
				Background(lipgloss.Color("#874BFD"))
		}
		tabs = append(tabs, style.Render(fmt.Sprintf("%d. %s", i+1, section)))
	}

	return strings.Join(tabs, " ")
}

func (m model) renderCurrentView() string {
	switch m.currentView {
	case TimelineView:
		return m.renderTimelineView()
	case FrequencyView:
		return m.renderFrequencyView()
	case TagsView:
		return m.renderTagsView()
	case MergesView:
		return m.renderMergesView()
	default:
		return "Unknown view"
	}
}

func (m model) renderTimelineView() string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("📅 Commit Timeline"))
	content.WriteString("\n")
	content.WriteString("Chronological view of repository commits")
	content.WriteString("\n\n")

	// Overall stats
	stats := m.analysis.OverallStats
	content.WriteString(fmt.Sprintf("📊 %s total commits from %s to %s\n",
		statsStyle.Render(fmt.Sprintf("%d", stats.TotalCommits)),
		stats.FirstCommit.Format("2006-01-02"),
		stats.LastCommit.Format("2006-01-02")))
	content.WriteString(fmt.Sprintf("👥 %s authors • 📈 %.1f commits/day average\n\n",
		statsStyle.Render(fmt.Sprintf("%d", stats.TotalAuthors)),
		stats.AveragePerDay))

	if len(m.timelineList.Items()) == 0 {
		content.WriteString("No commits to display")
		return content.String()
	}

	content.WriteString(m.timelineList.View())
	return content.String()
}

func (m model) renderFrequencyView() string {
	var content strings.Builder
	freq := m.analysis.FrequencyData

	content.WriteString(headerStyle.Render("📊 Commit Frequency Analysis"))
	content.WriteString("\n\n")

	// Activity summary
	content.WriteString(fmt.Sprintf("📅 Active on %s out of %s days\n",
		statsStyle.Render(fmt.Sprintf("%d", freq.TotalDays)),
		statsStyle.Render(fmt.Sprintf("%d", int(time.Since(m.analysis.OverallStats.FirstCommit).Hours()/24)))))
	content.WriteString(fmt.Sprintf("🔥 Current streak: %s days (longest: %s)\n",
		highlightStyle.Render(fmt.Sprintf("%d", freq.CommitStreak.Current)),
		statsStyle.Render(fmt.Sprintf("%d", freq.CommitStreak.Longest))))
	content.WriteString(fmt.Sprintf("📈 Max commits per day: %s\n\n",
		statsStyle.Render(fmt.Sprintf("%d", freq.MaxCommitsPerDay))))

	// Weekday pattern
	content.WriteString(headerStyle.Render("📅 Weekly Pattern"))
	content.WriteString("\n")
	days := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	maxDaily := 0
	for _, count := range freq.CommitsByWeekday {
		if count > maxDaily {
			maxDaily = count
		}
	}

	for i, day := range days {
		count := freq.CommitsByWeekday[i]
		if maxDaily > 0 {
			bars := strings.Repeat("█", (count*15)/maxDaily+1)
			content.WriteString(fmt.Sprintf("%s %s %d\n", day, bars, count))
		}
	}

	// Hourly pattern
	content.WriteString("\n")
	content.WriteString(headerStyle.Render("🕒 Hourly Pattern"))
	content.WriteString("\n")

	maxHourly := 0
	peakHour := 0
	for hour, count := range freq.CommitsByHour {
		if count > maxHourly {
			maxHourly = count
			peakHour = hour
		}
	}

	// Show a condensed hourly view
	content.WriteString("Hours: ")
	for hour := 0; hour < 24; hour += 4 {
		count := freq.CommitsByHour[hour]
		intensity := "·"
		if count > 0 {
			if count > maxHourly/2 {
				intensity = "█"
			} else {
				intensity = "▓"
			}
		}
		content.WriteString(fmt.Sprintf("%02d%s ", hour, intensity))
	}
	content.WriteString(fmt.Sprintf("\nPeak: %02d:00 (%d commits)\n",
		peakHour, maxHourly))

	return content.String()
}

func (m model) renderTagsView() string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("🏷️ Tags & Releases"))
	content.WriteString("\n")
	content.WriteString("Repository tags and release history")
	content.WriteString("\n\n")

	if len(m.analysis.Tags) == 0 {
		content.WriteString("No tags found in repository")
		return content.String()
	}

	content.WriteString(fmt.Sprintf("📋 %s total tags found\n\n",
		statsStyle.Render(fmt.Sprintf("%d", len(m.analysis.Tags)))))

	content.WriteString(m.tagsList.View())
	return content.String()
}

func (m model) renderMergesView() string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("🔀 Merge Commits"))
	content.WriteString("\n")
	content.WriteString("Analysis of merge commits and branch integration")
	content.WriteString("\n\n")

	if len(m.analysis.Merges) == 0 {
		content.WriteString("No merge commits found")
		return content.String()
	}

	content.WriteString(fmt.Sprintf("🔀 %s merge commits found\n",
		statsStyle.Render(fmt.Sprintf("%d", len(m.analysis.Merges)))))

	// Calculate merge stats
	totalFiles := 0
	totalAdditions := 0
	totalDeletions := 0
	for _, merge := range m.analysis.Merges {
		totalFiles += merge.FilesChanged
		totalAdditions += merge.Additions
		totalDeletions += merge.Deletions
	}

	content.WriteString(fmt.Sprintf("📁 %s files changed • +%s -%s lines\n\n",
		statsStyle.Render(fmt.Sprintf("%d", totalFiles)),
		statsStyle.Render(fmt.Sprintf("%d", totalAdditions)),
		statsStyle.Render(fmt.Sprintf("%d", totalDeletions))))

	content.WriteString(m.mergesList.View())
	return content.String()
}

func loadHistoryData() tea.Msg {
	analysis, err := analyzeHistory()
	if err != nil {
		return errMsg{err}
	}
	return dataLoadedMsg{analysis}
}

func analyzeHistory() (HistoryAnalysis, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return HistoryAnalysis{}, fmt.Errorf("failed to open repository: %w", err)
	}

	ref, err := repo.Head()
	if err != nil {
		return HistoryAnalysis{}, fmt.Errorf("failed to get HEAD: %w", err)
	}

	analysis := HistoryAnalysis{}

	// Analyze commits for timeline and frequency
	err = analyzeCommits(repo, ref.Hash(), &analysis)
	if err != nil {
		return HistoryAnalysis{}, fmt.Errorf("failed to analyze commits: %w", err)
	}

	// Analyze tags
	err = analyzeTags(repo, &analysis)
	if err != nil {
		return HistoryAnalysis{}, fmt.Errorf("failed to analyze tags: %w", err)
	}

	// Calculate overall stats
	calculateOverallStats(&analysis)

	return analysis, nil
}

func analyzeCommits(repo *git.Repository, fromHash plumbing.Hash, analysis *HistoryAnalysis) error {
	cIter, err := repo.Log(&git.LogOptions{From: fromHash})
	if err != nil {
		return err
	}

	var timeline []TimelineCommit
	var merges []MergeCommit
	frequencyData := FrequencyData{
		CommitsByDate:    make(map[string]int),
		CommitsByMonth:   make(map[string]int),
		CommitsByWeekday: make(map[int]int),
		CommitsByHour:    make(map[int]int),
		CommitsByAuthor:  make(map[string]int),
	}

	var commitDates []time.Time
	activeDaysSet := make(map[string]bool)

	err = cIter.ForEach(func(c *object.Commit) error {
		// Timeline data
		timelineCommit := TimelineCommit{
			Hash:        c.Hash.String(),
			ShortHash:   c.Hash.String()[:8],
			Message:     strings.Split(c.Message, "\n")[0],
			Author:      c.Author.Name,
			Email:       c.Author.Email,
			Date:        c.Author.When,
			ParentCount: c.NumParents(),
			IsMerge:     c.NumParents() > 1,
		}

		// Get file stats
		if stats, err := c.Stats(); err == nil {
			for _, stat := range stats {
				timelineCommit.Files = append(timelineCommit.Files, stat.Name)
				timelineCommit.Additions += stat.Addition
				timelineCommit.Deletions += stat.Deletion
			}
		}

		timeline = append(timeline, timelineCommit)

		// Merge analysis
		if timelineCommit.IsMerge {
			parents := c.ParentHashes
			var parentHashes []string
			for _, p := range parents {
				parentHashes = append(parentHashes, p.String()[:8])
			}

			merge := MergeCommit{
				Hash:         timelineCommit.Hash,
				ShortHash:    timelineCommit.ShortHash,
				Message:      timelineCommit.Message,
				Author:       timelineCommit.Author,
				Date:         timelineCommit.Date,
				ParentHashes: parentHashes,
				FilesChanged: len(timelineCommit.Files),
				Additions:    timelineCommit.Additions,
				Deletions:    timelineCommit.Deletions,
			}

			// Try to extract branch name from merge message
			if strings.Contains(strings.ToLower(merge.Message), "merge") {
				parts := strings.Fields(merge.Message)
				for i, part := range parts {
					if strings.ToLower(part) == "merge" && i+1 < len(parts) {
						merge.BranchMerged = parts[i+1]
						break
					}
				}
			}

			merges = append(merges, merge)
		}

		// Frequency analysis
		dateStr := timelineCommit.Date.Format("2006-01-02")
		monthStr := timelineCommit.Date.Format("2006-01")

		frequencyData.CommitsByDate[dateStr]++
		frequencyData.CommitsByMonth[monthStr]++
		frequencyData.CommitsByWeekday[int(timelineCommit.Date.Weekday())]++
		frequencyData.CommitsByHour[timelineCommit.Date.Hour()]++
		frequencyData.CommitsByAuthor[timelineCommit.Author]++

		activeDaysSet[dateStr] = true
		commitDates = append(commitDates, timelineCommit.Date)

		return nil
	})

	if err != nil {
		return err
	}

	// Sort timeline by date (newest first)
	sort.Slice(timeline, func(i, j int) bool {
		return timeline[i].Date.After(timeline[j].Date)
	})

	// Sort merges by date (newest first)
	sort.Slice(merges, func(i, j int) bool {
		return merges[i].Date.After(merges[j].Date)
	})

	// Calculate frequency stats
	frequencyData.TotalDays = len(activeDaysSet)

	// Find max commits per day
	for _, count := range frequencyData.CommitsByDate {
		if count > frequencyData.MaxCommitsPerDay {
			frequencyData.MaxCommitsPerDay = count
		}
	}

	// Calculate streaks
	frequencyData.CommitStreak = calculateCommitStreak(commitDates)

	analysis.Timeline = timeline
	analysis.Merges = merges
	analysis.FrequencyData = frequencyData

	return nil
}

func analyzeTags(repo *git.Repository, analysis *HistoryAnalysis) error {
	tagRefs, err := repo.Tags()
	if err != nil {
		return err
	}

	var tags []TagInfo

	err = tagRefs.ForEach(func(ref *plumbing.Reference) error {
		tag := TagInfo{
			Name: strings.TrimPrefix(ref.Name().String(), "refs/tags/"),
			Hash: ref.Hash().String()[:8],
		}

		// Try to get tag object for annotated tags
		tagObj, err := repo.TagObject(ref.Hash())
		if err == nil {
			// Annotated tag
			tag.Type = "annotated"
			tag.Date = tagObj.Tagger.When
			tag.Tagger = tagObj.Tagger.Name
			tag.Message = tagObj.Message
		} else {
			// Lightweight tag - points directly to commit
			tag.Type = "lightweight"
			commit, err := repo.CommitObject(ref.Hash())
			if err == nil {
				tag.Date = commit.Author.When
				tag.Tagger = commit.Author.Name
				tag.Message = commit.Message
			}
		}

		// Calculate commits since this tag
		// This is a simplified version - in a real implementation you might want
		// to count commits from tag to HEAD
		tag.CommitsSince = 0

		tags = append(tags, tag)
		return nil
	})

	if err != nil {
		return err
	}

	// Sort tags by date (newest first)
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Date.After(tags[j].Date)
	})

	analysis.Tags = tags
	return nil
}

func calculateCommitStreak(commitDates []time.Time) StreakInfo {
	if len(commitDates) == 0 {
		return StreakInfo{}
	}

	// Sort dates
	sort.Slice(commitDates, func(i, j int) bool {
		return commitDates[i].Before(commitDates[j])
	})

	// Get unique dates
	uniqueDates := make(map[string]bool)
	for _, date := range commitDates {
		uniqueDates[date.Format("2006-01-02")] = true
	}

	// Convert to sorted slice
	var dates []time.Time
	for dateStr := range uniqueDates {
		if date, err := time.Parse("2006-01-02", dateStr); err == nil {
			dates = append(dates, date)
		}
	}
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	if len(dates) == 0 {
		return StreakInfo{}
	}

	var currentStreak, longestStreak int
	var currentEnd, longestEnd time.Time

	// Calculate streaks
	streak := 1
	for i := 1; i < len(dates); i++ {
		if dates[i].Sub(dates[i-1]).Hours() <= 24*1.5 { // Allow 1.5 days gap
			streak++
		} else {
			if streak > longestStreak {
				longestStreak = streak
				longestEnd = dates[i-1]
			}
			streak = 1
		}
	}

	// Check final streak
	if streak > longestStreak {
		longestStreak = streak
		longestEnd = dates[len(dates)-1]
	}

	// Calculate current streak (from today backwards)
	now := time.Now()
	currentStreak = 0
	for i := len(dates) - 1; i >= 0; i-- {
		if now.Sub(dates[i]).Hours() <= 24*float64(currentStreak+1)+12 { // Allow some flexibility
			currentStreak++
			currentEnd = dates[i]
		} else {
			break
		}
	}

	return StreakInfo{
		Current:    currentStreak,
		Longest:    longestStreak,
		CurrentEnd: currentEnd,
		LongestEnd: longestEnd,
	}
}

func calculateOverallStats(analysis *HistoryAnalysis) {
	stats := &analysis.OverallStats

	if len(analysis.Timeline) == 0 {
		return
	}

	stats.TotalCommits = len(analysis.Timeline)
	stats.TotalTags = len(analysis.Tags)
	stats.TotalMerges = len(analysis.Merges)

	// Find date range
	stats.LastCommit = analysis.Timeline[0].Date // Timeline is sorted newest first
	stats.FirstCommit = analysis.Timeline[len(analysis.Timeline)-1].Date

	// Calculate active days and average
	stats.ActiveDays = analysis.FrequencyData.TotalDays
	daysSinceFirst := time.Since(stats.FirstCommit).Hours() / 24
	if daysSinceFirst > 0 {
		stats.AveragePerDay = float64(stats.TotalCommits) / daysSinceFirst
	}

	// Find most active author
	maxCommits := 0
	for author, count := range analysis.FrequencyData.CommitsByAuthor {
		if count > maxCommits {
			maxCommits = count
			stats.MostActiveAuthor = author
		}
	}
	stats.TotalAuthors = len(analysis.FrequencyData.CommitsByAuthor)

	// Find most active day
	maxDayCommits := 0
	for dateStr, count := range analysis.FrequencyData.CommitsByDate {
		if count > maxDayCommits {
			maxDayCommits = count
			stats.MostActiveDay = dateStr
		}
	}
}

// RunHistoryExplorer starts the advanced history explorer TUI
func RunHistoryExplorer() error {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#01FAC6")).
		BorderLeftForeground(lipgloss.Color("#01FAC6"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#DDDDDD"))

	timelineList := list.New([]list.Item{}, delegate, 0, 0)
	timelineList.SetShowStatusBar(false)
	timelineList.SetShowHelp(false)

	tagsList := list.New([]list.Item{}, delegate, 0, 0)
	tagsList.SetShowStatusBar(false)
	tagsList.SetShowHelp(false)

	mergesList := list.New([]list.Item{}, delegate, 0, 0)
	mergesList.SetShowStatusBar(false)
	mergesList.SetShowHelp(false)

	m := model{
		timelineList: timelineList,
		tagsList:     tagsList,
		mergesList:   mergesList,
		currentView:  TimelineView,
		loading:      true,
		tuiHelper: terminal.NewResponsiveTUIHelper(),
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
