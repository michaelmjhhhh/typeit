package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/michaelmjhhhh/typeit/internal/domain"
	"github.com/michaelmjhhhh/typeit/internal/infra"
)

func (m *Model) color(name string) lipgloss.Color {
	if len(m.Options.Themes) == 0 {
		return lipgloss.Color("7")
	}
	theme := m.Options.Themes[m.ThemeIndex]
	palette := theme.Dark
	if m.Options.Config.Theme.Mode == "Light" {
		palette = theme.Light
	}
	return lipgloss.Color(palette[name].Hex())
}
func (m *Model) style(name string) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(m.color(name))
}
func (m *Model) View() string {
	if m.Width < 25 || m.Height < 8 {
		return "Typeit\nPlease enlarge the terminal.\nCtrl+C to exit."
	}
	var body string
	switch m.Screen {
	case "loading":
		body = m.Spinner.View() + " " + m.Status + "\n\nEsc: cancel"
	case "error":
		body = "Unable to load repository\n\nR: retry   Esc: exit"
	case "title":
		body = m.titleView()
	case "typing":
		body = m.typingView()
	case "stage":
		body = m.stageView()
	case "animation":
		body = m.animationView()
	case "summary", "failure":
		body = m.summaryView()
	case "total":
		body = m.totalView()
	case "history":
		body = m.historyView()
	case "details":
		body = m.detailsView()
	case "stats":
		body = m.analyticsView()
	case "settings":
		body = m.settingsView()
	case "help":
		body = m.helpView()
	case "repos", "repo-list", "languages", "trending":
		body = m.selectionView()
	case "version":
		body = m.heading("VERSION CHECK") + fmt.Sprintf("\n\nCurrent: %s\nLatest release: %s\n\nEnter: release page   Esc: back", m.Options.Version, m.VersionInfo.Latest)
	case "share":
		body = m.shareView()
	}
	if m.Error != "" {
		body += "\n\n" + m.style("status_error").Render(m.Error)
	}
	if m.Busy && m.Screen != "loading" {
		body += "\n" + m.Spinner.View() + " Loading…"
	}
	w, h := max(1, m.Width-4), max(1, m.Height-2)
	if m.Screen == "animation" || m.Screen == "title" || m.Screen == "stage" || m.Screen == "summary" || m.Screen == "failure" || m.Screen == "total" {
		body = centerLines(body, w)
	}
	lines := strings.Split(body, "\n")
	if len(lines) > h {
		lines = lines[:h]
	}
	for i, line := range lines {
		lines[i] = ansi.Truncate(line, w, "…")
	}
	body = strings.Join(lines, "\n")
	return m.renderFrame(body)
}

func (m *Model) renderFrame(body string) string {
	foreground, background := m.color("text"), m.color("background")
	body = lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, body)
	// Lip Gloss v1's nested styles end with a full ANSI reset. Restore the
	// frame colors after each reset so text and padding never inherit the
	// terminal's defaults. Explicit backgrounds (such as the cursor) still win.
	profile := lipgloss.ColorProfile()
	const reset = "\x1b[0m"
	colors := strings.TrimSuffix(profile.String("").
		Foreground(profile.Color(string(foreground))).
		Background(profile.Color(string(background))).String(), reset)
	body = strings.ReplaceAll(body, reset, reset+colors)
	return lipgloss.NewStyle().Foreground(foreground).Background(background).
		Width(m.Width).Height(m.Height).Render(body)
}

func (m *Model) heading(s string) string { return m.style("title").Bold(true).Render(s) }
func (m *Model) titleView() string {
	counts := [5]int{}
	for _, c := range m.Challenges {
		for i, d := range domain.Difficulties {
			if c.Difficulty == d {
				counts[i]++
			}
		}
	}
	length := []string{"~100 characters", "~200 characters", "~500 characters", "Full chunks", "Entire files"}
	subtitle := []string{"Short code snippets", "Medium functions", "Long functions or classes", "Unpredictable length chunks", "Complete files as challenges"}
	body := m.style("metrics_score").Render(infra.Artwork("logo.json", "")) + "\n\n" + m.heading("Code Typing Challenge") + fmt.Sprintf("\n\nDifficulty: ← %s →\n%d challenges available\n%s\n%s\n\n", domain.Difficulties[m.Difficulty], counts[m.Difficulty], length[m.Difficulty], subtitle[m.Difficulty]) + "[←→/HL] Change Difficulty\n[R] Records  [A] Analytics  [S] Settings  [I/?] Help\n[SPACE] Start  [ESC] Quit\n\n" + m.repoLabel()
	if m.VersionInfo.Available {
		body += "\n[V] Update available: " + m.VersionInfo.Latest
	}
	return body
}
func (m *Model) repoLabel() string {
	s := m.Repository.Owner + "/" + m.Repository.Name
	if m.Repository.Branch != "" {
		s += " · " + m.Repository.Branch
	}
	if len(m.Repository.Commit) >= 7 {
		s += " @ " + m.Repository.Commit[:7]
	}
	if m.Repository.Dirty {
		s += " (modified)"
	}
	return m.style("status_info").Render(s)
}
func (m *Model) metricsView(v domain.Metrics) string {
	return fmt.Sprintf("%s  %s  %s  %s\n%s  %s", m.style("metrics_score").Render(fmt.Sprintf("Score %.0f", v.Score)), m.style("metrics_cpm_wpm").Render(fmt.Sprintf("WPM %.1f · CPM %.0f", v.WPM, v.CPM)), m.style("metrics_accuracy").Render(fmt.Sprintf("Accuracy %.1f%%", v.Accuracy)), m.style("metrics_duration").Render(fmt.Sprintf("Time %.1fs", float64(v.DurationMS)/1000)), fmt.Sprintf("Keystrokes %d · Mistakes %d", v.Keystrokes, v.Mistakes), v.RankName)
}
func (m *Model) stageView() string {
	s := m.Stages[len(m.Stages)-1]
	return m.heading("STAGE COMPLETE") + "\n\n" + m.style("metrics_score").Render(scoreArt(s.Score)) + "\n\n" + m.metricsView(s.Metrics) + "\n\n[SPACE] Continue   [ESC] Quit"
}
func (m *Model) summaryView() string {
	title := "SESSION COMPLETE"
	if m.Screen == "failure" {
		title = "SESSION FAILED"
	}
	s := m.Summary
	badge := ""
	if label := m.Best.Status(s.Score); label != "" {
		badge = "\n" + m.style("status_success").Render("★ "+label+" BEST")
	}
	tp, tt, p, total := domain.RankPositions(s.Score)
	rankDetail := fmt.Sprintf("\n%s tier · %d/%d (overall %d/%d)", s.TierName, tp, tt, p, total)
	return m.heading(title) + badge + rankDetail + "\n\n" + m.style("metrics_score").Render(infra.Artwork("ranks.json", s.RankName)) + "\n\n" + m.metricsView(s.Metrics) + fmt.Sprintf("\n\nStages %d/3 · Skipped %d · Partial effort %d keystrokes", s.Completed, s.Skipped, s.PartialKeys) + "\n\nD: details   R: retry   S: share   T: title   Esc: total summary"
}
func (m *Model) totalView() string {
	s := domain.Total(m.Totals)
	return m.heading("TOTAL SUMMARY") + fmt.Sprintf("\n\n%d sessions · %d completed stages\n\n", s.SessionsAttempted, s.StagesCompleted) + m.metricsView(s.Metrics) + "\n\nS: share   Enter/Esc: exit"
}
func (m *Model) scrollLines(lines []string, selected int) string {
	limit := max(1, m.Height-10)
	start := min(max(0, selected-limit/2), max(0, len(lines)-limit))
	end := min(len(lines), start+limit)
	return strings.Join(lines[start:end], "\n")
}
func (m *Model) textPage(lines []string) string {
	start := min(m.Scroll, max(0, len(lines)-max(1, m.Height-9)))
	return strings.Join(lines[start:min(len(lines), start+max(1, m.Height-9))], "\n")
}
