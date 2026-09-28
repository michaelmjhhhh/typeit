package ui

import (
	"context"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/michaelmjhhhh/typeit/internal/domain"
	"github.com/michaelmjhhhh/typeit/internal/infra"
)

type Options struct {
	Clock                                                     func() time.Time
	Version                                                   string
	Context                                                   context.Context
	DataDir, Path, Remote, Screen, Language, Period, RepoName string
	Languages                                                 []string
	Database                                                  *infra.Database
	Config                                                    infra.Config
	Themes                                                    []infra.Theme
}
type Model struct {
	AnimationAt                                      time.Time
	Best                                             domain.BestRecords
	VersionInfo                                      infra.VersionInfo
	SettingsSection                                  int
	OriginalConfig                                   infra.Config
	OriginalTheme                                    int
	SortBy, DateFilter                               int
	SortDescending                                   bool
	Options                                          Options
	Screen                                           string
	Width, Height                                    int
	Difficulty                                       int
	Challenges                                       []domain.Challenge
	Repository                                       domain.Repository
	Typing                                           *domain.Typing
	Challenge                                        domain.Challenge
	Stages                                           []domain.StageResult
	Summary                                          domain.SessionResult
	Totals                                           []domain.SessionResult
	History                                          []domain.SessionResult
	Repositories                                     []domain.Repository
	Trending                                         []infra.TrendingRepo
	Selected, Tab, Scroll, ThemeIndex                int
	Phase                                            string
	Paused                                           bool
	Now, CountdownAt, StartedAt, PausedAt, SessionAt time.Time
	PausedFor                                        time.Duration
	Streak                                           int
	Streaks                                          []int
	Error, Status, ReturnScreen                      string
	Busy                                             bool
	Spinner                                          spinner.Model
}
type versionMsg struct{ Info infra.VersionInfo }
type tickMsg time.Time
type loadedMsg struct {
	Challenges []domain.Challenge
	Repository domain.Repository
	Err        error
}
type historyMsg struct {
	Sessions []domain.SessionResult
	Err      error
}
type stagesMsg struct {
	Stages []domain.StageResult
	Err    error
}
type reposMsg struct {
	Repos []domain.Repository
	Err   error
}
type trendingMsg struct {
	Repos []infra.TrendingRepo
	Err   error
}
type savedMsg struct {
	Best    domain.BestRecords
	Summary domain.SessionResult
	Err     error
}
type statusMsg struct {
	Text string
	Err  error
}

func New(options Options) *Model {
	if options.Clock == nil {
		options.Clock = time.Now
	}
	if options.Context == nil {
		options.Context = context.Background()
	}
	if options.Path == "" {
		options.Path = "."
	}
	if options.Period == "" {
		options.Period = "daily"
	}
	s := spinner.New()
	s.Spinner = spinner.Dot
	m := &Model{Options: options, Width: 80, Height: 24, Difficulty: 1, Screen: "loading", Spinner: s, Now: options.Clock(), DateFilter: 2, SortDescending: true}
	for i, t := range options.Themes {
		if t.ID == options.Config.Theme.ID {
			m.ThemeIndex = i
		}
	}
	return m
}
func (m *Model) Init() tea.Cmd {
	var cmd tea.Cmd
	switch m.Options.Screen {
	case "history", "stats":
		m.Screen = m.Options.Screen
		cmd = m.loadHistory()
	case "repos", "repo-list":
		m.Screen = m.Options.Screen
		cmd = m.loadRepos()
	case "trending":
		if m.Options.Language == "" {
			m.Screen = "languages"
		} else {
			m.Screen = "trending"
			cmd = m.loadTrending(m.Options.Language)
		}
	default:
		cmd = m.load(m.Options.Path, m.Options.Remote)
	}
	return tea.Batch(cmd, m.Spinner.Tick, tick(), m.versionCheck())
}
func tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}
func (m *Model) load(path, remote string) tea.Cmd {
	m.Screen = "loading"
	m.Busy = true
	m.Status = "Scanning source code and extracting challenges…"
	if remote != "" {
		m.Status = "Cloning " + remote + "…"
	}
	o := m.Options
	return func() tea.Msg {
		if remote != "" {
			p, err := infra.Clone(o.Context, o.DataDir, remote)
			if err != nil {
				return loadedMsg{Err: err}
			}
			path = p
		}
		c, err := infra.LoadChallenges(o.Context, o.DataDir, path, o.Languages)
		return loadedMsg{Challenges: c, Repository: infra.GitInfo(o.Context, path), Err: err}
	}
}
func (m *Model) loadHistory() tea.Cmd {
	db := m.Options.Database
	m.Busy = true
	return func() tea.Msg {
		s, e := db.History()
		if e == nil {
			for i := range s {
				s[i].Stages, e = db.Stages(s[i].ID)
				if e != nil {
					break
				}
			}
		}
		return historyMsg{s, e}
	}
}
func (m *Model) loadRepos() tea.Cmd {
	dir, db, catalog := m.Options.DataDir, m.Options.Database, m.Screen == "repo-list"
	m.Busy = true
	return func() tea.Msg {
		r, e := infra.CachedRepositories(dir)
		if e != nil || !catalog {
			return reposMsg{r, e}
		}
		stored, e := db.Repositories()
		if e != nil {
			return reposMsg{Err: e}
		}
		seen := map[string]bool{}
		for _, repo := range r {
			seen[repo.Owner+"/"+repo.Name] = true
		}
		for _, repo := range stored {
			if !seen[repo.Owner+"/"+repo.Name] {
				r = append(r, repo)
			}
		}
		return reposMsg{r, nil}
	}
}

func (m *Model) loadTrending(language string) tea.Cmd {
	o := m.Options
	m.Busy = true
	m.Options.Language = language
	return func() tea.Msg {
		r, e := infra.Trending(o.Context, o.DataDir, language, o.Period)
		return trendingMsg{r, e}
	}
}
func (m *Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case versionMsg:
		m.VersionInfo = msg.Info
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	case tickMsg:
		m.Now = time.Time(msg)
		if m.Screen == "animation" && m.animationDone() {
			m.Screen = "summary"
		}
		if m.Screen == "typing" && m.Phase == "countdown" && !m.Paused && m.Now.Sub(m.CountdownAt) >= 2200*time.Millisecond {
			m.Phase = "active"
			m.StartedAt = m.Now
		}
		return m, tick()
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd
	case loadedMsg:
		m.Busy = false
		if msg.Err != nil {
			m.Error = msg.Err.Error()
			m.Screen = "error"
		} else {
			m.Challenges = msg.Challenges
			m.Repository = msg.Repository
			m.Screen = "title"
			m.Error = ""
		}
	case historyMsg:
		m.Busy = false
		if msg.Err != nil {
			m.Error = msg.Err.Error()
		} else {
			m.History = msg.Sessions
			m.Selected = 0
		}
	case stagesMsg:
		m.Busy = false
		if msg.Err != nil {
			m.Error = msg.Err.Error()
		} else {
			m.Summary.Stages = msg.Stages
			m.Scroll = 0
			m.Screen = "details"
		}
	case reposMsg:
		m.Busy = false
		if msg.Err != nil {
			m.Error = msg.Err.Error()
		} else {
			m.Repositories = msg.Repos
		}
	case trendingMsg:
		m.Busy = false
		if msg.Err != nil {
			m.Error = msg.Err.Error()
		} else {
			m.Trending = msg.Repos
			m.Selected = 0
			if m.Options.RepoName != "" {
				for _, r := range msg.Repos {
					if strings.EqualFold(r.Name, m.Options.RepoName) || strings.EqualFold(strings.Split(r.Name, "/")[len(strings.Split(r.Name, "/"))-1], m.Options.RepoName) {
						m.Options.RepoName = ""
						return m, m.load("", r.Name)
					}
				}
				m.Error = "Repository not found in the trending list"
			}
		}
	case savedMsg:
		m.Busy = false
		if msg.Err != nil {
			m.Error = "Save session: " + msg.Err.Error()
		} else {
			m.Summary = msg.Summary
			m.Best = msg.Best
			m.Totals = append(m.Totals, msg.Summary)
		}
	case statusMsg:
		m.Busy = false
		if msg.Err != nil {
			m.Error = msg.Err.Error()
		} else {
			m.Status = msg.Text
		}
	case tea.KeyMsg:
		m.Now = m.Options.Clock()
		if msg.Type == tea.KeyRunes && len(msg.Runes) > 1 && m.Screen != "typing" && !msg.Paste {
			var commands []tea.Cmd
			for _, r := range msg.Runes {
				single := msg
				single.Runes = []rune{r}
				if cmd := m.key(single); cmd != nil {
					commands = append(commands, cmd)
				}
			}
			return m, tea.Sequence(commands...)
		}
		return m, m.key(msg)
	}
	return m, nil
}
func (m *Model) startSession() {
	m.Stages = nil
	m.SessionAt = m.Now
	m.Summary = domain.SessionResult{}
	m.Error = ""
	m.nextChallenge()
}
func (m *Model) nextChallenge() {
	var choices []domain.Challenge
	for _, c := range m.Challenges {
		if c.Difficulty == domain.Difficulties[m.Difficulty] {
			choices = append(choices, c)
		}
	}
	if len(choices) == 0 {
		m.Error = "No challenges available for this difficulty. Please try a different difficulty or repository."
		m.Screen = "title"
		return
	}
	m.Challenge = choices[rand.IntN(len(choices))]
	m.Typing = domain.NewTyping(m.Challenge)
	m.Phase = "ready"
	m.Screen = "typing"
	m.Paused = false
	m.PausedFor = 0
	m.StartedAt = time.Time{}
	m.Streak = 0
	m.Streaks = nil
}
func (m *Model) elapsed() time.Duration {
	if m.StartedAt.IsZero() {
		return 0
	}
	end := m.Now
	if m.Paused {
		end = m.PausedAt
	}
	return max(0, end.Sub(m.StartedAt)-m.PausedFor)
}
func (m *Model) finishStage(skipped, failed bool) tea.Cmd {
	t := m.Typing
	streaks := append([]int{}, m.Streaks...)
	if m.Streak > 0 {
		streaks = append(streaks, m.Streak)
	}
	metrics := domain.Calculate(t.Position+t.Mistakes, t.Mistakes, m.elapsed().Milliseconds(), false)
	if m.StartedAt.IsZero() {
		metrics = domain.Metrics{RankName: "Unranked", TierName: "Beginner"}
	}
	m.Stages = append(m.Stages, domain.StageResult{Metrics: metrics, Challenge: m.Challenge, Streaks: streaks, Skipped: skipped, Failed: failed})
	m.Paused = false
	if failed {
		return m.finishSession()
	}
	if skipped {
		m.nextChallenge()
	} else {
		m.Screen = "stage"
	}
	return nil
}
func (m *Model) finishSession() tea.Cmd {
	return m.saveSession(false)
}
func (m *Model) saveSession(aborted bool) tea.Cmd {
	m.Summary = domain.Summarize(m.Stages)
	if aborted {
		m.Summary.Successful = false
	}
	m.Summary.Repository = m.Repository
	m.Summary.Difficulty = domain.Difficulties[m.Difficulty]
	m.Summary.Mode = string(m.Summary.Difficulty)
	m.Summary.StartedAt = m.SessionAt.UTC().Format(time.RFC3339Nano)
	m.Screen = "animation"
	m.AnimationAt = m.Now
	if !m.Summary.Successful {
		m.Screen = "failure"
	}
	m.Busy = true
	db := m.Options.Database
	summary := m.Summary
	return func() tea.Msg {
		previous, err := db.History()
		if err != nil {
			return savedMsg{Err: err}
		}
		best := domain.Records(previous, time.Now())
		err = db.Save(&summary)
		return savedMsg{Summary: summary, Err: err, Best: best}
	}
}

func (m *Model) versionCheck() tea.Cmd {
	o := m.Options
	if o.Version == "" {
		return nil
	}
	return func() tea.Msg { v, _ := infra.CheckVersion(o.Context, o.DataDir, o.Version); return versionMsg{v} }
}
