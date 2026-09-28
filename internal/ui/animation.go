package ui

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/michaelmjhhhh/typeit/assets"
)

func animationLines(rank string) []string {
	b, _ := assets.Files.ReadFile("messages.json")
	var all map[string][]string
	_ = json.Unmarshal(b, &all)
	lines := all[rank]
	if len(lines) == 0 {
		return []string{"> analyzing performance data...", "> calculating skill results...", "> determining rank classification...", "> rank assignment complete."}
	}
	return lines
}
func (m *Model) animationView() string {
	elapsed := m.Now.Sub(m.AnimationAt)
	var rendered []string
	for _, line := range animationLines(m.Summary.RankName) {
		chars := []rune(line)
		duration := time.Duration(len(chars)) * 40 * time.Millisecond
		if elapsed < duration {
			n := min(len(chars), max(0, int(elapsed/(40*time.Millisecond))+1))
			rendered = append(rendered, m.style("status_info").Render(string(chars[:n])+"█"))
			break
		}
		rendered = append(rendered, m.style("status_success").Render(line))
		elapsed -= duration + 500*time.Millisecond
		if elapsed < 0 {
			break
		}
	}
	return m.heading("ANALYZING PERFORMANCE") + "\n\n" + strings.Join(rendered, "\n") + "\n\n[S] Skip animation"
}
func (m *Model) animationDone() bool {
	duration := 3 * time.Second
	for _, line := range animationLines(m.Summary.RankName) {
		duration += time.Duration(len([]rune(line)))*40*time.Millisecond + 500*time.Millisecond
	}
	return m.Now.Sub(m.AnimationAt) >= duration
}
