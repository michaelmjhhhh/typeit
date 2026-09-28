package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/michaelmjhhhh/typeit/internal/domain"
	"github.com/michaelmjhhhh/typeit/internal/infra"
)

func (m *Model) key(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()
	if key == "ctrl+c" {
		return tea.Quit
	}
	if m.Screen == "typing" {
		return m.typingKey(msg)
	}
	if m.Busy && (key != "esc" || m.Screen == "summary" || m.Screen == "failure" || m.Screen == "animation" || m.Screen == "total") {
		return nil
	}
	switch m.Screen {
	case "loading":
		if key == "esc" {
			return tea.Quit
		}
	case "error":
		if key == "esc" || key == "q" {
			return tea.Quit
		}
		if key == "r" {
			m.Error = ""
			return m.load(m.Options.Path, m.Options.Remote)
		}
	case "title":
		switch strings.ToLower(key) {
		case "v":
			m.ReturnScreen = "title"
			m.Screen = "version"
		case "left", "h":
			m.Difficulty = (m.Difficulty + 4) % 5
			m.Error = ""
		case "right", "l":
			m.Difficulty = (m.Difficulty + 1) % 5
			m.Error = ""
		case " ", "enter":
			m.startSession()
		case "r":
			m.Screen = "history"
			return m.loadHistory()
		case "a":
			m.Screen = "stats"
			return m.loadHistory()
		case "s":
			m.ReturnScreen = "title"
			m.OriginalConfig = m.Options.Config
			m.OriginalTheme = m.ThemeIndex
			m.SettingsSection = 0
			m.Screen = "settings"
		case "i", "?":
			m.ReturnScreen = "title"
			m.Screen = "help"
			m.Tab = 0
			m.Scroll = 0
		case "esc", "q":
			return tea.Quit
		}
	case "stage":
		if key == " " || key == "enter" {
			if domain.Summarize(m.Stages).Completed >= 3 {
				return m.finishSession()
			}
			m.nextChallenge()
		}
		if key == "esc" {
			return m.saveSession(true)
		}
	case "animation":
		if key == "s" || key == "S" {
			m.Screen = "summary"
		}
	case "summary", "failure":
		switch strings.ToLower(key) {
		case "r":
			m.startSession()
		case "d":
			m.ReturnScreen = m.Screen
			m.Screen = "details"
			m.Scroll = 0
		case "s":
			m.ReturnScreen = m.Screen
			m.Screen = "share"
		case "t":
			m.Screen = "title"
		case "esc", "q":
			m.Screen = "total"
		}
	case "total":
		if key == "esc" || key == "q" || key == "enter" {
			return tea.Quit
		}
		if key == "s" {
			total := domain.Total(m.Totals)
			m.Summary.Metrics = total.Metrics
			m.Summary.PartialMistakes = 0
			m.ReturnScreen = "total"
			m.Screen = "share"
		}
	case "history":
		rows := m.historyRows()
		switch key {
		case "r":
			return m.loadHistory()
		case "s":
			m.SortBy = (m.SortBy + 1) % 4
			if m.SortBy == 0 {
				m.SortDescending = !m.SortDescending
			}
			m.Selected = 0
		case "f":
			m.DateFilter = (m.DateFilter + 1) % 4
			m.Selected = 0
		case "up", "k":
			m.Selected = max(0, m.Selected-1)
		case "down", "j":
			m.Selected = min(max(0, len(rows)-1), m.Selected+1)
		case "enter", " ":
			if len(rows) > 0 {
				m.Summary = rows[m.Selected]
				m.ReturnScreen = "history"
				m.Busy = true
				db := m.Options.Database
				id := m.Summary.ID
				return func() tea.Msg { s, e := db.Stages(id); return stagesMsg{s, e} }
			}
		case "esc":
			return m.backHome()
		}
	case "stats", "help":
		switch key {
		case "r":
			if m.Screen == "stats" {
				return m.loadHistory()
			}
		case "tab", "right", "l":
			m.Tab = (m.Tab + 1) % 4
			m.Scroll = 0
		case "shift+tab", "left", "h":
			m.Tab = (m.Tab + 3) % 4
			m.Scroll = 0
		case "down", "j":
			m.Scroll++
		case "up", "k":
			m.Scroll = max(0, m.Scroll-1)
		case "esc":
			return m.backHome()
		}
	case "details":
		switch key {
		case "down", "j":
			m.Scroll++
		case "up", "k":
			m.Scroll = max(0, m.Scroll-1)
		case "esc":
			m.Screen = m.ReturnScreen
		}
	case "settings":
		return m.settingsKey(key)
	case "repo-list":
		if key == "esc" {
			return m.backHome()
		}
	case "repos":
		switch key {
		case "up", "k":
			m.Selected = max(0, m.Selected-1)
		case "down", "j":
			m.Selected = min(max(0, len(m.Repositories)-1), m.Selected+1)
		case "enter", " ":
			if len(m.Repositories) > 0 {
				return m.load(m.Repositories[m.Selected].Path, "")
			}
		case "esc":
			return m.backHome()
		}
	case "languages":
		count := len(infra.Languages()) + 1
		switch key {
		case "up", "k":
			m.Selected = max(0, m.Selected-1)
		case "down", "j":
			m.Selected = min(count-1, m.Selected+1)
		case "enter", " ":
			lang := ""
			if m.Selected > 0 {
				lang = infra.Languages()[m.Selected-1].Name
			}
			m.Screen = "trending"
			m.Selected = 0
			return m.loadTrending(lang)
		case "esc":
			return m.backHome()
		}
	case "trending":
		switch key {
		case "up", "k":
			m.Selected = max(0, m.Selected-1)
		case "down", "j":
			m.Selected = min(max(0, len(m.Trending)-1), m.Selected+1)
		case "enter", " ":
			if len(m.Trending) > 0 {
				return m.load("", m.Trending[m.Selected].Name)
			}
		case "esc":
			m.Screen = "languages"
			m.Selected = 0
			m.Error = ""
		}
	case "version":
		if key == "esc" {
			m.Screen = m.ReturnScreen
		}
		if key == "enter" {
			return func() tea.Msg {
				return statusMsg{Err: infra.OpenURL("https://github.com/unhappychoice/gittype/releases")}
			}
		}
	case "share":
		if key == "esc" {
			m.Screen = m.ReturnScreen
		}
		if key == "1" || key == "2" || key == "3" || key == "4" {
			link := ShareURL(m.Summary, int(key[0]-'1'))
			return func() tea.Msg { err := infra.OpenURL(link); return statusMsg{Text: link, Err: err} }
		}
	}
	return nil
}
func (m *Model) backHome() tea.Cmd {
	m.Error = ""
	m.Selected = 0
	m.Scroll = 0
	if len(m.Challenges) == 0 {
		return tea.Quit
	}
	m.Screen = "title"
	return nil
}
func (m *Model) typingKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()
	if key == "esc" {
		if m.Paused {
			m.resume()
		} else {
			m.Paused = true
			m.PausedAt = m.Now
		}
		return nil
	}
	if m.Paused {
		switch strings.ToLower(key) {
		case "s":
			if m.skips() < 3 {
				return m.finishStage(true, false)
			}
			m.Error = "No skips remaining"
			return nil
		case "q":
			return m.finishStage(false, true)
		default:
			m.resume()
			return nil
		}
	}
	if m.Phase == "ready" {
		if key == " " {
			m.Phase = "countdown"
			m.CountdownAt = m.Now
		}
		return nil
	}
	if m.Phase != "active" {
		return nil
	}
	var chars []rune
	switch msg.Type {
	case tea.KeyEnter:
		chars = []rune{'\n'}
	case tea.KeyTab:
		chars = []rune{'\t'}
	case tea.KeySpace:
		chars = []rune{' '}
	case tea.KeyRunes:
		if msg.Paste {
			return nil
		}
		chars = msg.Runes
	}
	for _, r := range chars {
		if m.Typing.Input(r) {
			m.Streak++
		} else {
			if m.Streak > 0 {
				m.Streaks = append(m.Streaks, m.Streak)
			}
			m.Streak = 0
		}
		if m.Typing.Completed() {
			return m.finishStage(false, false)
		}
	}
	return nil
}
func (m *Model) resume() {
	d := m.Now.Sub(m.PausedAt)
	if m.Phase == "countdown" {
		m.CountdownAt = m.CountdownAt.Add(d)
	} else if m.Phase == "active" {
		m.PausedFor += d
	}
	m.Paused = false
	m.Error = ""
}
func (m *Model) skips() int {
	n := 0
	for _, s := range m.Stages {
		if s.Skipped {
			n++
		}
	}
	return n
}
func (m *Model) settingsKey(key string) tea.Cmd {
	switch key {
	case "left", "right", "h", "l", "tab":
		m.SettingsSection = 1 - m.SettingsSection
	case "up", "k":
		if m.SettingsSection == 0 {
			m.Options.Config.Theme.Mode = "Dark"
		} else {
			m.ThemeIndex = max(0, m.ThemeIndex-1)
		}
	case "down", "j":
		if m.SettingsSection == 0 {
			m.Options.Config.Theme.Mode = "Light"
		} else {
			m.ThemeIndex = min(len(m.Options.Themes)-1, m.ThemeIndex+1)
		}
	case "esc":
		m.Options.Config = m.OriginalConfig
		m.ThemeIndex = m.OriginalTheme
		m.Screen = m.ReturnScreen
		return nil
	case " ":
		m.Options.Config.Theme.ID = m.Options.Themes[m.ThemeIndex].ID
		dir := m.Options.DataDir
		c := m.Options.Config
		m.Screen = m.ReturnScreen
		return func() tea.Msg { return statusMsg{Err: infra.SaveConfig(dir, c)} }
	}
	return nil
}
