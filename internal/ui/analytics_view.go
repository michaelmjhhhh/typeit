package ui

import (
	"fmt"
	"strings"

	"github.com/michaelmjhhhh/typeit/internal/domain"
)

func (m *Model) analyticsView() string {
	a := domain.Analyze(m.History, m.Now)
	tabs := []string{"Overview", "Trends", "Languages", "Repositories"}
	title := m.heading("ANALYTICS · " + tabs[m.Tab] + " · Last 90 days")
	if a.Sessions == 0 {
		return title + "\n\nNo sessions yet.\n\nTab: next view   R: refresh   Esc: back"
	}
	var lines []string
	switch m.Tab {
	case 0:
		lines = []string{fmt.Sprintf("Sessions: %d  │  Avg CPM: %.1f  │  Best CPM: %.1f  │  Accuracy: %.1f%%", a.Sessions, a.CPM, a.BestCPM, a.Accuracy), fmt.Sprintf("Practice: %.1fh  │  Avg session: %.1fm  │  Mistakes: %d  │  Streak: %d days", a.Hours, a.AverageMinutes, a.Mistakes, a.Streak), "", "RECENT ACTIVITY"}
		for _, d := range a.Days[max(0, len(a.Days)-7):] {
			lines = append(lines, fmt.Sprintf("%s  %s %d sessions", d.Date, m.style("metrics_score").Render(strings.Repeat("█", min(40, d.Sessions))), d.Sessions))
		}
		lines = append(lines, "", "TOP REPOSITORIES")
		for _, g := range a.Repositories[:min(3, len(a.Repositories))] {
			lines = append(lines, fmt.Sprintf("%-30s %.1f CPM · %d sessions", g.Name, g.CPM, g.Count))
		}
		lines = append(lines, "", "TOP LANGUAGES")
		for _, g := range a.Languages[:min(3, len(a.Languages))] {
			lines = append(lines, fmt.Sprintf("%-20s %.1f CPM · %d stages", g.Name, g.CPM, g.Count))
		}
	case 1:
		lines = []string{"DATE          CPM    ACCURACY   SESSIONS   SPEED"}
		for _, d := range a.Days {
			lines = append(lines, fmt.Sprintf("%s %7.1f %9.1f%% %8d   %s", d.Date, d.CPM, d.Accuracy, d.Sessions, m.style("metrics_cpm_wpm").Render(strings.Repeat("█", min(30, int(d.CPM/25))))))
		}
	case 2, 3:
		groups := a.Languages
		if m.Tab == 3 {
			groups = a.Repositories
		}
		if len(groups) == 0 {
			lines = []string{"No completed stages available."}
			break
		}
		selected := min(m.Scroll, len(groups)-1)
		for i, g := range groups {
			line := fmt.Sprintf("%-28s %5d runs %7.1f CPM %6.1f%%", g.Name, g.Count, g.CPM, g.Accuracy)
			if i == selected {
				line = m.style("key_action").Render("▸ " + line)
			} else {
				line = "  " + line
			}
			lines = append(lines, line)
		}
		g := groups[selected]
		lines = append(lines, "", m.heading(g.Name), fmt.Sprintf("WPM %.1f · CPM %.1f · Accuracy %.1f%% · Average score %.0f", g.WPM, g.CPM, g.Accuracy, g.Score), fmt.Sprintf("Best CPM %.1f · Best accuracy %.1f%%", g.BestCPM, g.BestAccuracy), fmt.Sprintf("Keystrokes %d · Mistakes %d · Practice %.1fm", g.Keys, g.Mistakes, float64(g.DurationMS)/60000), fmt.Sprintf("Stages completed %d · Attempted %d · Skipped %d", g.Completed, g.Attempted, g.Skipped))
	}
	return title + "\n\n" + m.textPage(lines) + "\n\nTab/←/→: view   ↑/↓: select/scroll   R: refresh   Esc: back"
}
