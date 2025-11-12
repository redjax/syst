package healthService

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/redjax/syst/internal/utils/terminal"
)

type HealthReport struct {
	OverallScore    int
	Issues          []HealthIssue
	LargeFiles      []LargeFile
	RepositoryStats RepositoryStats
	SecurityIssues  []SecurityIssue
	BestPractices   []BestPracticeCheck
	GitIgnoreStatus GitIgnoreAnalysis
	CommitHealth    CommitHealthAnalysis
}

type HealthIssue struct {
	Severity    string // "high", "medium", "low"
	Category    string
	Title       string
	Description string
	Suggestion  string
}

type LargeFile struct {
	Path string
	Size int64
	Type string
}

type RepositoryStats struct {
	TotalFiles      int
	TotalSize       int64
	BinaryFiles     int
	TextFiles       int
	AverageFileSize int64
	OldestFile      time.Time
	NewestFile      time.Time
}

type SecurityIssue struct {
	Type        string
	File        string
	Description string
	Risk        string
}

type BestPracticeCheck struct {
	Name        string
	Status      string // "pass", "fail", "warning"
	Description string
	Suggestion  string
}

type GitIgnoreAnalysis struct {
	Exists          bool
	MissingPatterns []string
	UnusedPatterns  []string
	RecommendedAdds []string
}

type CommitHealthAnalysis struct {
	AverageMessageLength int
	LargeCommits         []LargeCommit
	FrequentAuthors      []AuthorStats
	CommitPatterns       map[string]int
}

type LargeCommit struct {
	Hash         string
	Size         int
	FilesChanged int
	Date         time.Time
	Message      string
}

type AuthorStats struct {
	Name       string
	Commits    int
	Percentage float64
}

type model struct {
	report    HealthReport
	err       error
	loading   bool
	tuiHelper *terminal.ResponsiveTUIHelper
	sections  []string
	selected  int
}

type reportLoadedMsg struct {
	report HealthReport
}

type errMsg struct {
	err error
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#FF5F87")).
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

	goodStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAA00"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87"))

	criticalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)
)

func (m model) Init() tea.Cmd {
	return loadHealthReport
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.tuiHelper.HandleWindowSizeMsg(msg)
		return m, nil

	case reportLoadedMsg:
		m.report = msg.report
		m.loading = false
		m.sections = []string{
			"Overview",
			"Issues",
			"Large Files",
			"Security",
			"Best Practices",
			"Git Ignore",
			"Commit Health",
		}
		return m, nil

	case errMsg:
		m.err = msg.err
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c", "esc"))):
			return m, tea.Quit
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			if m.selected > 0 {
				m.selected--
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			if m.selected < len(m.sections)-1 {
				m.selected++
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.loading {
		return "\n  Analyzing repository health...\n"
	}

	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("\n  Error: %v\n", m.err))
	}

	var sections []string

	// Title with overall score
	score := m.report.OverallScore
	scoreColor := goodStyle
	if score < 70 {
		scoreColor = warningStyle
	}
	if score < 50 {
		scoreColor = errorStyle
	}

	title := titleStyle.Render(fmt.Sprintf("🏥 Repository Health - Score: %s",
		scoreColor.Render(fmt.Sprintf("%d/100", score))))
	sections = append(sections, title)

	// Navigation menu
	var menuItems []string
	for i, section := range m.sections {
		style := lipgloss.NewStyle()
		if i == m.selected {
			style = style.Foreground(lipgloss.Color("#01FAC6")).Bold(true)
		}
		menuItems = append(menuItems, style.Render(fmt.Sprintf("%d. %s", i+1, section)))
	}
	menu := strings.Join(menuItems, " | ")
	sections = append(sections, menu)

	// Selected section content
	content := m.renderSelectedSection()
	sections = append(sections, sectionStyle.Render(content))

	// Instructions
	instructions := helpStyle.Render("↑/↓: navigate sections • q: quit")
	sections = append(sections, instructions)

	return strings.Join(sections, "\n")
}

func (m model) renderSelectedSection() string {
	switch m.sections[m.selected] {
	case "Overview":
		return m.renderOverview()
	case "Issues":
		return m.renderIssues()
	case "Large Files":
		return m.renderLargeFiles()
	case "Security":
		return m.renderSecurity()
	case "Best Practices":
		return m.renderBestPractices()
	case "Git Ignore":
		return m.renderGitIgnore()
	case "Commit Health":
		return m.renderCommitHealth()
	default:
		return "Section not implemented"
	}
}

func (m model) renderOverview() string {
	var content strings.Builder
	stats := m.report.RepositoryStats

	content.WriteString(headerStyle.Render("📊 Repository Overview"))
	content.WriteString("\n\n")

	content.WriteString(fmt.Sprintf("Total Files: %s\n",
		goodStyle.Render(fmt.Sprintf("%d", stats.TotalFiles))))
	content.WriteString(fmt.Sprintf("Repository Size: %s\n",
		goodStyle.Render(formatBytes(stats.TotalSize))))
	content.WriteString(fmt.Sprintf("Binary Files: %s\n",
		goodStyle.Render(fmt.Sprintf("%d", stats.BinaryFiles))))
	content.WriteString(fmt.Sprintf("Text Files: %s\n",
		goodStyle.Render(fmt.Sprintf("%d", stats.TextFiles))))
	content.WriteString(fmt.Sprintf("Average File Size: %s\n",
		goodStyle.Render(formatBytes(stats.AverageFileSize))))

	// Issue summary
	content.WriteString("\n")
	content.WriteString(headerStyle.Render("🚨 Issue Summary"))
	content.WriteString("\n\n")

	highIssues := 0
	mediumIssues := 0
	lowIssues := 0

	for _, issue := range m.report.Issues {
		switch issue.Severity {
		case "high":
			highIssues++
		case "medium":
			mediumIssues++
		case "low":
			lowIssues++
		}
	}

	content.WriteString(fmt.Sprintf("High Priority: %s\n",
		criticalStyle.Render(fmt.Sprintf("%d", highIssues))))
	content.WriteString(fmt.Sprintf("Medium Priority: %s\n",
		warningStyle.Render(fmt.Sprintf("%d", mediumIssues))))
	content.WriteString(fmt.Sprintf("Low Priority: %s\n",
		goodStyle.Render(fmt.Sprintf("%d", lowIssues))))

	return content.String()
}

func (m model) renderIssues() string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("🚨 Health Issues"))
	content.WriteString("\n\n")

	if len(m.report.Issues) == 0 {
		content.WriteString(goodStyle.Render("✓ No issues found! Your repository is healthy."))
		return content.String()
	}

	for _, issue := range m.report.Issues {
		var style lipgloss.Style
		var icon string

		switch issue.Severity {
		case "high":
			style = criticalStyle
			icon = "🔴"
		case "medium":
			style = warningStyle
			icon = "🟡"
		case "low":
			style = goodStyle
			icon = "🟢"
		}

		content.WriteString(fmt.Sprintf("%s %s [%s]\n",
			icon, style.Render(issue.Title), issue.Category))
		content.WriteString(fmt.Sprintf("   %s\n", issue.Description))
		if issue.Suggestion != "" {
			content.WriteString(fmt.Sprintf("   💡 %s\n", issue.Suggestion))
		}
		content.WriteString("\n")
	}

	return content.String()
}

func (m model) renderLargeFiles() string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("📦 Large Files"))
	content.WriteString("\n\n")

	if len(m.report.LargeFiles) == 0 {
		content.WriteString(goodStyle.Render("✓ No unusually large files detected."))
		return content.String()
	}

	content.WriteString("Files larger than 1MB:\n\n")

	for _, file := range m.report.LargeFiles {
		sizeStyle := goodStyle
		if file.Size > 10*1024*1024 { // 10MB
			sizeStyle = criticalStyle
		} else if file.Size > 5*1024*1024 { // 5MB
			sizeStyle = warningStyle
		}

		content.WriteString(fmt.Sprintf("%s %s (%s)\n",
			sizeStyle.Render(formatBytes(file.Size)),
			file.Path,
			file.Type))
	}

	return content.String()
}

func (m model) renderSecurity() string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("🔒 Security Analysis"))
	content.WriteString("\n\n")

	if len(m.report.SecurityIssues) == 0 {
		content.WriteString(goodStyle.Render("✓ No obvious security issues detected."))
		return content.String()
	}

	for _, issue := range m.report.SecurityIssues {
		var style lipgloss.Style
		switch issue.Risk {
		case "high":
			style = criticalStyle
		case "medium":
			style = warningStyle
		default:
			style = goodStyle
		}

		content.WriteString(fmt.Sprintf("%s %s\n",
			style.Render(issue.Type), issue.File))
		content.WriteString(fmt.Sprintf("   %s\n\n", issue.Description))
	}

	return content.String()
}

func (m model) renderBestPractices() string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("✨ Best Practices"))
	content.WriteString("\n\n")

	for _, check := range m.report.BestPractices {
		var style lipgloss.Style
		var icon string

		switch check.Status {
		case "pass":
			style = goodStyle
			icon = "✓"
		case "warning":
			style = warningStyle
			icon = "⚠"
		case "fail":
			style = criticalStyle
			icon = "✗"
		}

		content.WriteString(fmt.Sprintf("%s %s\n",
			icon, style.Render(check.Name)))
		content.WriteString(fmt.Sprintf("   %s\n", check.Description))
		if check.Suggestion != "" {
			content.WriteString(fmt.Sprintf("   💡 %s\n", check.Suggestion))
		}
		content.WriteString("\n")
	}

	return content.String()
}

func (m model) renderGitIgnore() string {
	var content strings.Builder
	gi := m.report.GitIgnoreStatus

	content.WriteString(headerStyle.Render("📋 .gitignore Analysis"))
	content.WriteString("\n\n")

	if gi.Exists {
		content.WriteString(goodStyle.Render("✓ .gitignore file exists"))
	} else {
		content.WriteString(criticalStyle.Render("✗ .gitignore file missing"))
	}
	content.WriteString("\n\n")

	if len(gi.RecommendedAdds) > 0 {
		content.WriteString("Recommended additions:\n")
		for _, pattern := range gi.RecommendedAdds {
			content.WriteString(fmt.Sprintf("  + %s\n", warningStyle.Render(pattern)))
		}
		content.WriteString("\n")
	}

	if len(gi.MissingPatterns) > 0 {
		content.WriteString("Common missing patterns:\n")
		for _, pattern := range gi.MissingPatterns {
			content.WriteString(fmt.Sprintf("  ! %s\n", criticalStyle.Render(pattern)))
		}
	}

	return content.String()
}

func (m model) renderCommitHealth() string {
	var content strings.Builder
	ch := m.report.CommitHealth

	content.WriteString(headerStyle.Render("📝 Commit Health"))
	content.WriteString("\n\n")

	content.WriteString(fmt.Sprintf("Average Message Length: %s characters\n",
		goodStyle.Render(fmt.Sprintf("%d", ch.AverageMessageLength))))

	if len(ch.LargeCommits) > 0 {
		content.WriteString("\nLarge commits (>100 files):\n")
		for _, commit := range ch.LargeCommits {
			content.WriteString(fmt.Sprintf("%s %d files - %s\n",
				warningStyle.Render(commit.Hash[:8]),
				commit.FilesChanged,
				commit.Message))
		}
	}

	if len(ch.FrequentAuthors) > 0 {
		content.WriteString("\nTop contributors:\n")
		for i, author := range ch.FrequentAuthors {
			if i >= 5 { // Top 5
				break
			}
			content.WriteString(fmt.Sprintf("%s: %s commits (%.1f%%)\n",
				author.Name,
				goodStyle.Render(fmt.Sprintf("%d", author.Commits)),
				author.Percentage))
		}
	}

	return content.String()
}

func loadHealthReport() tea.Msg {
	report, err := analyzeRepositoryHealth()
	if err != nil {
		return errMsg{err}
	}
	return reportLoadedMsg{report}
}

func analyzeRepositoryHealth() (HealthReport, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return HealthReport{}, fmt.Errorf("failed to open repository: %w", err)
	}

	report := HealthReport{
		Issues:         []HealthIssue{},
		LargeFiles:     []LargeFile{},
		SecurityIssues: []SecurityIssue{},
		BestPractices:  []BestPracticeCheck{},
	}

	// Analyze repository stats
	report.RepositoryStats = analyzeRepositoryStats(repo)

	// Check for large files
	report.LargeFiles = findLargeFiles()

	// Analyze gitignore
	report.GitIgnoreStatus = analyzeGitIgnore()

	// Analyze commit health
	report.CommitHealth = analyzeCommitHealth(repo)

	// Run best practice checks
	report.BestPractices = runBestPracticeChecks(repo)

	// Check for security issues
	report.SecurityIssues = checkSecurityIssues()

	// Generate issues based on analysis
	report.Issues = generateHealthIssues(report)

	// Calculate overall score
	report.OverallScore = calculateHealthScore(report)

	return report, nil
}

func analyzeRepositoryStats(repo *git.Repository) RepositoryStats {
	stats := RepositoryStats{}

	// Get the HEAD commit
	ref, err := repo.Head()
	if err != nil {
		return stats
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return stats
	}

	tree, err := commit.Tree()
	if err != nil {
		return stats
	}

	var totalSize int64
	var fileCount int
	var binaryCount int

	// Walk the git tree (only tracked files)
	err = tree.Files().ForEach(func(file *object.File) error {
		fileCount++
		totalSize += file.Size

		// Simple binary file detection based on file extension
		if isBinaryFile(file.Name) {
			binaryCount++
		}

		return nil
	})

	if err != nil {
		return stats
	}

	stats.TotalFiles = fileCount
	stats.TotalSize = totalSize
	stats.BinaryFiles = binaryCount
	stats.TextFiles = fileCount - binaryCount
	if fileCount > 0 {
		stats.AverageFileSize = totalSize / int64(fileCount)
	}

	return stats
}

func findLargeFiles() []LargeFile {
	var largeFiles []LargeFile
	const threshold = 1024 * 1024 // 1MB

	// Open repository to get tracked files only
	repo, err := git.PlainOpen(".")
	if err != nil {
		return largeFiles
	}

	ref, err := repo.Head()
	if err != nil {
		return largeFiles
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return largeFiles
	}

	tree, err := commit.Tree()
	if err != nil {
		return largeFiles
	}

	// Walk only tracked files in git tree
	err = tree.Files().ForEach(func(file *object.File) error {
		if file.Size > threshold {
			fileType := "text"
			if isBinaryFile(file.Name) {
				fileType = "binary"
			}

			largeFiles = append(largeFiles, LargeFile{
				Path: file.Name,
				Size: file.Size,
				Type: fileType,
			})
		}

		return nil
	})

	if err != nil {
		return largeFiles
	}

	// Sort by size descending
	sort.Slice(largeFiles, func(i, j int) bool {
		return largeFiles[i].Size > largeFiles[j].Size
	})

	return largeFiles
}

func analyzeGitIgnore() GitIgnoreAnalysis {
	analysis := GitIgnoreAnalysis{}

	// Check if .gitignore exists
	if _, err := os.Stat(".gitignore"); err == nil {
		analysis.Exists = true
	}

	// Recommend common patterns based on what we find in the repo
	repo, _ := git.PlainOpen(".")
	var foundFiles []string

	if repo != nil {
		if ref, err := repo.Head(); err == nil {
			if commit, err := repo.CommitObject(ref.Hash()); err == nil {
				if tree, err := commit.Tree(); err == nil {
					// #nosec G104 - ForEach callback errors are handled by returning nil in all cases
					tree.Files().ForEach(func(file *object.File) error {
						foundFiles = append(foundFiles, file.Name)
						return nil
					})
				}
			}
		}
	}

	// Check for common patterns that should be ignored
	potentialIssues := map[string]string{
		"node_modules": "node_modules/",
		".log":         "*.log",
		".env":         ".env*",
		"dist":         "dist/",
		"build":        "build/",
		".tmp":         "*.tmp",
		".swp":         "*.swp",
		".DS_Store":    ".DS_Store",
		"Thumbs.db":    "Thumbs.db",
		"__pycache__":  "__pycache__/",
		".pyc":         "*.pyc",
		"target":       "target/",
		".class":       "*.class",
		"bin":          "bin/",
		"obj":          "obj/",
		".exe":         "*.exe",
	}

	var recommendedAdds []string

	// Check which problematic files/patterns exist in tracked files
	for pattern, gitignorePattern := range potentialIssues {
		for _, file := range foundFiles {
			if strings.Contains(strings.ToLower(file), pattern) {
				recommendedAdds = append(recommendedAdds, gitignorePattern)
				break
			}
		}
	}

	if !analysis.Exists {
		// If no .gitignore, recommend basic patterns plus any we detected
		basicPatterns := []string{
			"*.log",
			".env*",
			".DS_Store",
			"Thumbs.db",
			"*.tmp",
			"*.swp",
		}
		analysis.RecommendedAdds = append(basicPatterns, recommendedAdds...)
	} else {
		// Check which patterns are missing from existing .gitignore
		content, err := os.ReadFile(".gitignore")
		if err == nil {
			gitignoreContent := string(content)
			for _, pattern := range recommendedAdds {
				if !strings.Contains(gitignoreContent, pattern) {
					analysis.MissingPatterns = append(analysis.MissingPatterns, pattern)
				}
			}
		}
	}

	// Remove duplicates
	analysis.RecommendedAdds = removeDuplicates(analysis.RecommendedAdds)
	analysis.MissingPatterns = removeDuplicates(analysis.MissingPatterns)

	return analysis
}

func removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

func analyzeCommitHealth(repo *git.Repository) CommitHealthAnalysis {
	analysis := CommitHealthAnalysis{
		CommitPatterns: make(map[string]int),
	}

	ref, err := repo.Head()
	if err != nil {
		return analysis
	}

	cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return analysis
	}

	var totalMessageLength int
	var commitCount int
	authorStats := make(map[string]int)

	err = cIter.ForEach(func(c *object.Commit) error {
		commitCount++
		totalMessageLength += len(c.Message)
		authorStats[c.Author.Name]++

		// Check for large commits (simplified)
		stats, err := c.Stats()
		if err == nil && len(stats) > 100 {
			analysis.LargeCommits = append(analysis.LargeCommits, LargeCommit{
				Hash:         c.Hash.String(),
				FilesChanged: len(stats),
				Date:         c.Author.When,
				Message:      strings.Split(c.Message, "\n")[0],
			})
		}

		return nil
	})

	if commitCount > 0 {
		analysis.AverageMessageLength = totalMessageLength / commitCount
	}

	// Convert author stats
	var authors []AuthorStats
	for name, commits := range authorStats {
		percentage := float64(commits) / float64(commitCount) * 100
		authors = append(authors, AuthorStats{
			Name:       name,
			Commits:    commits,
			Percentage: percentage,
		})
	}

	sort.Slice(authors, func(i, j int) bool {
		return authors[i].Commits > authors[j].Commits
	})

	analysis.FrequentAuthors = authors

	return analysis
}

func runBestPracticeChecks(repo *git.Repository) []BestPracticeCheck {
	var checks []BestPracticeCheck

	// Check for README
	readme := BestPracticeCheck{
		Name:        "README file",
		Description: "Repository should have a README file",
	}
	if _, err := os.Stat("README.md"); err == nil {
		readme.Status = "pass"
	} else if _, err := os.Stat("README.txt"); err == nil {
		readme.Status = "pass"
	} else {
		readme.Status = "fail"
		readme.Suggestion = "Add a README.md file to document your project"
	}
	checks = append(checks, readme)

	// Check for .gitignore
	gitignore := BestPracticeCheck{
		Name:        ".gitignore file",
		Description: "Repository should have a .gitignore file",
	}
	if _, err := os.Stat(".gitignore"); err == nil {
		gitignore.Status = "pass"
	} else {
		gitignore.Status = "fail"
		gitignore.Suggestion = "Add a .gitignore file to exclude unnecessary files"
	}
	checks = append(checks, gitignore)

	// Check for license
	license := BestPracticeCheck{
		Name:        "License file",
		Description: "Repository should have a license file",
	}
	if _, err := os.Stat("LICENSE"); err == nil {
		license.Status = "pass"
	} else if _, err := os.Stat("LICENSE.txt"); err == nil {
		license.Status = "pass"
	} else {
		license.Status = "warning"
		license.Suggestion = "Consider adding a LICENSE file"
	}
	checks = append(checks, license)

	return checks
}

func checkSecurityIssues() []SecurityIssue {
	var issues []SecurityIssue

	// Check for common sensitive files only in tracked files
	repo, err := git.PlainOpen(".")
	if err != nil {
		return issues
	}

	ref, err := repo.Head()
	if err != nil {
		return issues
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return issues
	}

	tree, err := commit.Tree()
	if err != nil {
		return issues
	}

	// Patterns that might indicate sensitive files
	sensitivePatterns := []string{
		".env",
		"config.json",
		"secrets.json",
		"id_rsa",
		"id_dsa",
		".pem",
		".key",
		"password",
		"secret",
		"token",
		"api_key",
	}

	// Check only tracked files in git tree
	// #nosec G104 - ForEach callback errors are handled by returning nil in all cases
	tree.Files().ForEach(func(file *object.File) error {
		fileName := strings.ToLower(file.Name)

		for _, pattern := range sensitivePatterns {
			if strings.Contains(fileName, pattern) {
				risk := "medium"
				if strings.Contains(fileName, "key") || strings.Contains(fileName, "secret") || strings.Contains(fileName, "password") {
					risk = "high"
				}

				issues = append(issues, SecurityIssue{
					Type:        "Sensitive File",
					File:        file.Name,
					Description: "File may contain sensitive information and is tracked in git",
					Risk:        risk,
				})
				break
			}
		}

		return nil
	})

	return issues
}

func generateHealthIssues(report HealthReport) []HealthIssue {
	var issues []HealthIssue

	// Issues from large files
	for _, file := range report.LargeFiles {
		if file.Size > 10*1024*1024 { // 10MB
			issues = append(issues, HealthIssue{
				Severity:    "high",
				Category:    "Performance",
				Title:       fmt.Sprintf("Very large file: %s", file.Path),
				Description: fmt.Sprintf("File is %s, which may impact repository performance", formatBytes(file.Size)),
				Suggestion:  "Consider using Git LFS for large files",
			})
		}
	}

	// Issues from missing best practices
	for _, check := range report.BestPractices {
		if check.Status == "fail" {
			issues = append(issues, HealthIssue{
				Severity:    "medium",
				Category:    "Best Practice",
				Title:       fmt.Sprintf("Missing %s", check.Name),
				Description: check.Description,
				Suggestion:  check.Suggestion,
			})
		}
	}

	// Issues from security
	for _, security := range report.SecurityIssues {
		severity := "low"
		if security.Risk == "high" {
			severity = "high"
		} else if security.Risk == "medium" {
			severity = "medium"
		}

		issues = append(issues, HealthIssue{
			Severity:    severity,
			Category:    "Security",
			Title:       fmt.Sprintf("%s: %s", security.Type, security.File),
			Description: security.Description,
			Suggestion:  "Review and remove sensitive information if present",
		})
	}

	return issues
}

func calculateHealthScore(report HealthReport) int {
	score := 100

	// Deduct points for issues
	for _, issue := range report.Issues {
		switch issue.Severity {
		case "high":
			score -= 15
		case "medium":
			score -= 10
		case "low":
			score -= 5
		}
	}

	// Deduct points for failed best practices
	for _, check := range report.BestPractices {
		if check.Status == "fail" {
			score -= 10
		} else if check.Status == "warning" {
			score -= 5
		}
	}

	if score < 0 {
		score = 0
	}

	return score
}

func isBinaryFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	binaryExts := []string{
		".exe", ".dll", ".so", ".dylib", ".bin", ".jpg", ".jpeg", ".png", ".gif",
		".pdf", ".zip", ".tar", ".gz", ".rar", ".7z", ".mp3", ".mp4", ".avi",
		".mov", ".wmv", ".flv", ".webm", ".ogg", ".wav", ".ico", ".ttf", ".woff",
		".woff2", ".eot", ".otf", ".class", ".jar", ".war", ".ear",
	}

	for _, binaryExt := range binaryExts {
		if ext == binaryExt {
			return true
		}
	}

	return false
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
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// RunHealthCheck starts the repository health check TUI
func RunHealthCheck() error {
	m := model{
		loading:   true,
		tuiHelper: terminal.NewResponsiveTUIHelper(),
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
