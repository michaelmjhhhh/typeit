package ui

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/charmbracelet/lipgloss"
	"github.com/michaelmjhhhh/typeit/internal/domain"
)

func (m *Model) typingView() string {
	t := m.Typing
	if t == nil {
		return "Loading challenge…"
	}
	completed := 0
	for _, s := range m.Stages {
		if !s.Skipped && !s.Failed {
			completed++
		}
	}
	metrics := domain.Calculate(t.Position+t.Mistakes, t.Mistakes, m.elapsed().Milliseconds(), false)
	// The original live accuracy uses cursor progress, unlike final stage accuracy.
	metrics.Accuracy = 0
	if t.Position > 0 {
		metrics.Accuracy = float64(max(0, t.Position-t.Mistakes)) / float64(t.Position) * 100
	}
	header := fmt.Sprintf(" %s:%d-%d [%s] [%s] · Stage %d/3", m.Challenge.Path, m.Challenge.StartLine, m.Challenge.EndLine, m.Challenge.Language, m.Challenge.Difficulty, completed+1)
	footer := fmt.Sprintf("%d/%d characters · %d skips remaining · Esc: pause", t.Position, len(t.Text), 3-m.skips())
	switch {
	case m.Paused:
		footer = m.heading("PAUSED") + " · Esc/any key: resume · S: skip · Q: end session"
	case m.Phase == "ready":
		footer = m.heading("Press SPACE to start") + " · Esc: options"
	case m.Phase == "countdown":
		elapsed := m.Now.Sub(m.CountdownAt)
		label := "GO!"
		if elapsed < 1800*time.Millisecond {
			label = fmt.Sprintf("%d", 3-int(elapsed/(600*time.Millisecond)))
		}
		footer = m.heading(label)
	}
	width := m.Width - 4
	percent := 0
	if len(t.Text) > 0 {
		percent = t.Position * 100 / len(t.Text)
	}
	metricsLine := fmt.Sprintf(" WPM: %.0f | CPM: %.0f | Accuracy: %.0f%% | Mistakes: %d | Streak: %d | Time: %.0fs | Skips: %d", metrics.WPM, metrics.CPM, metrics.Accuracy, t.Mistakes, m.Streak, m.elapsed().Seconds(), 3-m.skips())
	if m.Height < 18 {
		return header + "\n" + m.codeView() + "\n" + footer
	}
	return m.panel("Challenge", header, width, 3) + "\n" + m.panel("Code", m.codeView(), width, m.Height-13) + "\n" + m.panel("Metrics", metricsLine, width, 3) + "\n" + m.panel("Progress", centerLines(fmt.Sprintf("%d%%", percent), width-2), width, 3) + "\n" + footer
}
func (m *Model) codeView() string {
	t := m.Typing
	pos := t.DisplayCursor()
	currentLine := 0
	for i, r := range t.Display {
		if i >= pos {
			break
		}
		if r == '\n' {
			currentLine++
		}
	}
	height := max(1, m.Height-15)
	var lines []string
	number := func(n int) string { return m.style("text_secondary").Render(fmt.Sprintf("%4d │ ", n)) }
	for i, line := range m.Challenge.PreContext {
		lines = append(lines, number(m.Challenge.StartLine-len(m.Challenge.PreContext)+i)+m.style("text_secondary").Render(line))
	}
	var line strings.Builder
	lineNo := 0
	column := 0
	flush := func() {
		lines = append(lines, number(m.Challenge.StartLine+lineNo)+line.String())
		line.Reset()
		lineNo++
		column = 0
	}
	for i, r := range t.Display {
		if r == '\n' {
			flush()
			continue
		}
		style := m.style("typing_untyped_text")
		if domain.InComment(t.DisplayMap[i], t.Ranges) {
			style = m.style("text_secondary")
		} else if i < pos {
			style = m.style("typing_typed_text")
		}
		if i == pos && !t.Completed() {
			style = m.cursorStyle()
		}
		text := string(r)
		if unicode.IsControl(r) {
			text = "�"
		}
		if column < m.Width-12 {
			line.WriteString(style.Render(text))
		}
		column += lipgloss.Width(text)
	}
	flush()
	for i, line := range m.Challenge.PostContext {
		lines = append(lines, number(m.Challenge.EndLine+i+1)+m.style("text_secondary").Render(line))
	}
	start := min(max(0, currentLine+len(m.Challenge.PreContext)-height/2), max(0, len(lines)-height))
	return strings.Join(lines[start:min(len(lines), start+height)], "\n")
}

func (m *Model) cursorStyle() lipgloss.Style {
	bg := m.color("typing_cursor_bg")
	if m.Typing.Wrong {
		bg = m.color("typing_mistake_bg")
	}
	return m.style("typing_cursor_fg").Background(bg)
}
