package ui

import (
	"fmt"
	"strings"

	"github.com/michaelmjhhhh/typeit/internal/domain"
	"github.com/michaelmjhhhh/typeit/internal/infra"
)

func (m *Model) settingsView() string {
	lines := []string{}
	section := "Color mode  ·  Theme"
	if m.SettingsSection == 0 {
		section = "[Color mode]  ·  Theme"
		for _, mode := range []string{"Dark", "Light"} {
			line := "  " + mode
			if mode == m.Options.Config.Theme.Mode {
				line = m.style("key_action").Render("▸ " + mode)
			}
			lines = append(lines, line)
		}
		return m.heading("SETTINGS") + "\n\n" + section + "\n\n" + strings.Join(lines, "\n") + "\n\nTheme: " + m.Options.Themes[m.ThemeIndex].Name + "\n\n←/→: section   ↑/↓: choose   Space: save   Esc: cancel"
	}
	section = "Color mode  ·  [Theme]"
	for i, t := range m.Options.Themes {
		line := t.Name + " — " + t.Description
		if i == m.ThemeIndex {
			line = m.style("key_action").Render("▸ " + line)
		} else {
			line = "  " + line
		}
		lines = append(lines, line)
	}
	return m.heading("SETTINGS · "+m.Options.Config.Theme.Mode) + "\n\n" + section + "\n\n" + m.scrollLines(lines, m.ThemeIndex) + "\n\n←/→: section   ↑/↓: choose   Space: save   Esc: cancel"
}
func (m *Model) selectionView() string {
	var lines []string
	title := "CACHED REPOSITORIES"
	switch m.Screen {
	case "repos", "repo-list":
		for _, r := range m.Repositories {
			state := "cached"
			if r.Path == "" {
				state = "history only"
			}
			lines = append(lines, r.Owner+"/"+r.Name+" · "+state)
		}
	case "languages":
		title = "TRENDING · SELECT LANGUAGE"
		lines = append(lines, "All languages")
		for _, l := range infra.Languages() {
			lines = append(lines, l.Display)
		}
	case "trending":
		title = "TRENDING · " + m.Options.Period
		for _, r := range m.Trending {
			lines = append(lines, fmt.Sprintf("%-32s ★ %s  %s\n    %s", r.Name, r.Stars, r.Language, r.Description))
		}
	}
	for i, line := range lines {
		if i == m.Selected {
			lines[i] = m.style("key_action").Render("▸ " + line)
		} else {
			lines[i] = "  " + line
		}
	}
	if len(lines) == 0 {
		lines = append(lines, "No repositories available.")
	}
	return m.heading(title) + "\n\n" + m.scrollLines(lines, m.Selected) + "\n\n↑/↓: select   Enter: play   Esc: back"
}
func (m *Model) helpView() string {
	names := []string{"Controls", "Scoring", "Ranks", "About / CLI"}
	var lines []string
	switch m.Tab {
	case 0:
		lines = strings.Split("Title: ←/→ or H/L select Easy, Normal, Hard, Wild or Zen.\nSpace selects the challenge; Space again starts the countdown.\nType the code exactly. Enter advances a newline.\nComments, blank lines and leading indentation are skipped.\nA wrong key counts as a mistake and leaves the cursor in place.\nEsc pauses the timer. Any key resumes without typing.\nWhile paused: S skips (up to three); Q ends the session.\nComplete three stages to finish a session.\nCtrl+C exits from any screen.", "\n")
	case 1:
		lines = strings.Split("WPM = CPM / 5\nStage CPM = correct keystrokes / elapsed minutes\nAccuracy = correct keystrokes / all keystrokes × 100\nBase score = CPM × accuracy / 100 × 10\nConsistency bonus: up to 70% of the base score\nTime bonus: 20 points per second below chars / 10\nMistake penalty: 5 points per mistake\nFinal score = max(0, (base + bonuses − penalty) × 2 + 100)\nPaused time and countdown time never affect your score.\nSkipped and failed stages are recorded as partial effort.", "\n")
	case 2:
		for _, r := range domain.Ranks {
			lines = append(lines, fmt.Sprintf("%-24s %-13s %d–%d", r.Name, r.Tier, r.Min, r.Max))
		}
	case 3:
		lines = strings.Split("Typeit · Go + Bubble Tea + Lip Gloss\nBased on GitType by unhappychoice · MIT license\nhttps://github.com/unhappychoice/gittype\n\ntypeit [path] [--langs rust,python]\ntypeit --repo owner/repo\ntypeit history | stats | export\ntypeit cache stats | list | clear\ntypeit repo list | play | clear [--force]\ntypeit trending [language] [repo] [--period daily]\n\nCustom themes: ~/.typeit/themes/*.json\nUse TYPEIT_DATA_DIR to choose an isolated data directory.", "\n")
	}
	return m.heading("HELP · "+names[m.Tab]) + "\n\n" + m.textPage(lines) + "\n\nTab/←/→: section   ↑/↓: scroll   Esc: back"
}
