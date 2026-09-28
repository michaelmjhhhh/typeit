package ui

import (
	"fmt"
	"strings"
)

func (m *Model) historyView() string {
	lines := []string{}
	for i, s := range m.historyRows() {
		date := s.StartedAt
		if len(date) > 19 {
			date = date[:19]
		}
		line := fmt.Sprintf("%s  %-18s %7.0f pts  %5.1f WPM  %5.1f%%  %s", date, s.Repository.Name, s.Score, s.WPM, s.Accuracy, s.Difficulty)
		if i == m.Selected {
			line = m.style("key_action").Render("▸ " + line)
		} else {
			line = "  " + line
		}
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		lines = append(lines, "No sessions yet. Complete a typing session to create a record.")
	}
	return m.heading("SESSION HISTORY · "+[]string{"All time", "Last 7 days", "Last 30 days", "Last 90 days"}[m.DateFilter]+" · Sort: "+[]string{"Date", "Score", "Repository", "Duration"}[m.SortBy]) + "\n\n" + m.scrollLines(lines, m.Selected) + "\n\n↑/↓: select   Enter: details   S: sort   F: filter   R: refresh   Esc: back"
}
func (m *Model) detailsView() string {
	s := m.Summary
	lines := []string{fmt.Sprintf("%s/%s · %s · %s", s.Repository.Owner, s.Repository.Name, s.Difficulty, s.StartedAt), m.metricsView(s.Metrics), ""}
	for i, stage := range s.Stages {
		state := "complete"
		if stage.Skipped {
			state = "skipped"
		}
		if stage.Failed {
			state = "failed"
		}
		lines = append(lines, fmt.Sprintf("Stage %d · %s · %s:%d-%d", i+1, state, stage.Challenge.Path, stage.Challenge.StartLine, stage.Challenge.EndLine), m.metricsView(stage.Metrics), "")
	}
	return m.heading("SESSION DETAILS") + "\n\n" + m.textPage(strings.Split(strings.Join(lines, "\n"), "\n")) + "\n\n↑/↓: scroll   Esc: back"
}
