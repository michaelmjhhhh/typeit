package ui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/michaelmjhhhh/typeit/assets"
)

func (m *Model) panel(title, body string, width, height int) string {
	width = max(4, width)
	height = max(3, height)
	border := m.style("border")
	inner := width - 2
	top := border.Render("┌" + ansi.Truncate(title, inner, "") + strings.Repeat("─", max(0, inner-ansi.StringWidth(title))) + "┐")
	lines := strings.Split(body, "\n")
	rows := []string{top}
	for i := 0; i < height-2; i++ {
		line := ""
		if i < len(lines) {
			line = ansi.Truncate(lines[i], inner, "")
		}
		rows = append(rows, border.Render("│")+line+strings.Repeat(" ", max(0, inner-ansi.StringWidth(line)))+border.Render("│"))
	}
	rows = append(rows, border.Render("└"+strings.Repeat("─", inner)+"┘"))
	return strings.Join(rows, "\n")
}
func centerLines(body string, width int) string {
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		lines[i] = lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(ansi.Truncate(line, width, "…"))
	}
	return strings.Join(lines, "\n")
}
func scoreArt(score float64) string {
	b, err := assets.Files.ReadFile("digits.json")
	if err != nil {
		return ""
	}
	var digits [][]string
	if json.Unmarshal(b, &digits) != nil {
		return ""
	}
	rows := make([]string, 4)
	for _, r := range fmt.Sprintf("%.0f", score) {
		if r < '0' || r > '9' {
			continue
		}
		for i, line := range digits[int(r-'0')] {
			rows[i] += line + " "
		}
	}
	return strings.Join(rows, "\n")
}
