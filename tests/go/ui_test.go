package gittype_test

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/michaelmjhhhh/typeit/internal/domain"
	"github.com/michaelmjhhhh/typeit/internal/infra"
	"github.com/michaelmjhhhh/typeit/internal/ui"
)

func testModel(t *testing.T) *ui.Model {
	t.Helper()
	dir := t.TempDir()
	db, err := infra.OpenDatabase(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	themes, err := infra.Themes(dir)
	if err != nil {
		t.Fatal(err)
	}
	config, err := infra.LoadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	m := ui.New(ui.Options{Context: context.Background(), DataDir: dir, Database: db, Themes: themes, Config: config})
	m.Options.Clock = func() time.Time { return m.Now }
	m.Screen = "title"
	m.Difficulty = 0
	m.Challenges = []domain.Challenge{{ID: "challenge", Code: "你好()", Path: "main.go", StartLine: 1, EndLine: 1, Language: "go", Difficulty: domain.Easy}}
	m.Repository = domain.Repository{Owner: "local", Name: "test"}
	return m
}
func send(m *ui.Model, key string) tea.Cmd {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	switch key {
	case "down":
		msg.Type = tea.KeyDown
	case "up":
		msg.Type = tea.KeyUp
	case "left":
		msg.Type = tea.KeyLeft
	case "right":
		msg.Type = tea.KeyRight
	case "tab":
		msg.Type = tea.KeyTab
	case "esc":
		msg.Type = tea.KeyEsc
	case "enter":
		msg.Type = tea.KeyEnter
	case " ":
		msg.Type = tea.KeySpace
	}
	_, cmd := m.Update(msg)
	return cmd
}
func TestThreeStageSessionAndSave(t *testing.T) {
	m := testModel(t)
	send(m, " ")
	if m.Screen != "typing" || m.Phase != "ready" {
		t.Fatal("challenge preview missing")
	}
	for stage := 0; stage < 3; stage++ {
		send(m, " ")
		if m.Phase != "countdown" {
			t.Fatal("countdown missing")
		}
		m.Phase = "active"
		m.StartedAt = m.Now
		m.Now = m.Now.Add(5 * time.Second)
		var cmd tea.Cmd
		for _, r := range "你好()" {
			cmd = send(m, string(r))
		}
		if stage < 2 {
			if m.Screen != "stage" {
				t.Fatal("stage result missing")
			}
			send(m, " ")
		} else {
			if m.Screen != "stage" {
				t.Fatal("final stage summary missing")
			}
			cmd = send(m, " ")
			if m.Screen != "animation" || cmd == nil {
				t.Fatal("session did not save")
			}
			m.Update(cmd())
			send(m, "s")
			if m.Screen != "summary" {
				t.Fatal("animation skip failed")
			}
		}
	}
	if m.Summary.Completed != 3 || m.Summary.ID == 0 || len(m.Totals) != 1 {
		t.Fatalf("%+v", m.Summary)
	}
	h, err := m.Options.Database.History()
	if err != nil || len(h) != 1 {
		t.Fatalf("%v %v", h, err)
	}
}
func TestPauseSkipAndLiteralKeys(t *testing.T) {
	m := testModel(t)
	m.Challenges[0].Code = "sq"
	send(m, " ")
	m.Phase = "active"
	m.StartedAt = m.Now
	m.Now = m.Now.Add(time.Second)
	send(m, "esc")
	m.Now = m.Now.Add(time.Minute)
	send(m, "esc")
	if m.Paused || m.PausedFor != time.Minute {
		t.Fatal("pause duration not excluded")
	}
	send(m, "s")
	if m.Typing.Position != 1 {
		t.Fatal("s skipped instead of typing")
	}
	send(m, "esc")
	send(m, "s")
	if len(m.Stages) != 1 || !m.Stages[0].Skipped || m.Screen != "typing" {
		t.Fatal("skip behavior incorrect")
	}
	for i := 0; i < 2; i++ {
		send(m, "esc")
		send(m, "s")
	}
	send(m, "esc")
	send(m, "s")
	if len(m.Stages) != 3 || !strings.Contains(m.Error, "No skips") {
		t.Fatal("skip limit not enforced")
	}
}
func TestViewsFitTerminal(t *testing.T) {
	m := testModel(t)
	send(m, " ")
	m.Summary = domain.Summarize(nil)
	m.Stages = []domain.StageResult{{Challenge: m.Challenge}}
	screens := []string{"title", "typing", "stage", "summary", "failure", "history", "details", "stats", "settings", "help", "repos", "languages", "trending", "share", "total", "error", "loading"}
	for _, screen := range screens {
		for _, size := range [][2]int{{80, 24}, {120, 40}, {32, 10}, {10, 4}} {
			m.Screen = screen
			m.Width = size[0]
			m.Height = size[1]
			view := m.View()
			if strings.TrimSpace(view) == "" {
				t.Fatalf("blank %s", screen)
			}
			for _, line := range strings.Split(view, "\n") {
				if m.Width >= 25 && ansi.StringWidth(line) > m.Width {
					t.Fatalf("%s overflow: %d > %d", screen, ansi.StringWidth(line), m.Width)
				}
			}
		}
	}
}
func TestSettingsCancelAndSave(t *testing.T) {
	m := testModel(t)
	original := m.Options.Config
	send(m, "s")
	send(m, "down")
	if m.Options.Config.Theme.Mode != "Light" {
		t.Fatal("preview missing")
	}
	send(m, "esc")
	if m.Options.Config != original {
		t.Fatal("cancel saved preview")
	}
	send(m, "s")
	send(m, "down")
	cmd := send(m, " ")
	if cmd == nil {
		t.Fatal("save command missing")
	}
	m.Update(cmd())
	saved, err := infra.LoadConfig(m.Options.DataDir)
	if err != nil || saved.Theme.Mode != "Light" {
		t.Fatal("did not save", err)
	}
}
func TestInputClockExcludesCountdown(t *testing.T) {
	m := testModel(t)
	send(m, " ")
	send(m, " ")
	before := m.Now
	m.Now = m.Now.Add(10 * time.Second)
	send(m, "esc")
	m.Now = m.Now.Add(20 * time.Second)
	send(m, "esc")
	if m.CountdownAt.Sub(before) != 20*time.Second {
		t.Fatal("countdown pause duration lost")
	}
	m.Phase = "active"
	m.StartedAt = m.Now
	m.Now = m.Now.Add(2 * time.Second)
	for _, r := range "你好()" {
		send(m, string(r))
	}
	if m.Stages[0].DurationMS != 2000 {
		t.Fatalf("countdown included: %d", m.Stages[0].DurationMS)
	}
}
func TestCoalescedNavigationKeys(t *testing.T) {
	m := testModel(t)
	m.Difficulty = 1
	send(m, "lll")
	if m.Difficulty != 4 {
		t.Fatalf("coalesced navigation lost: %d", m.Difficulty)
	}
}

func TestAbortFromStageResult(t *testing.T) {
	m := testModel(t)
	send(m, " ")
	m.Phase = "active"
	m.StartedAt = m.Now
	m.Now = m.Now.Add(5 * time.Second)
	send(m, "你好()")
	cmd := send(m, "esc")
	if cmd == nil || m.Screen != "failure" || m.Summary.Successful {
		t.Fatal("abort did not open failure summary")
	}
	m.Update(cmd())
	if m.Summary.Successful || m.Summary.ID == 0 || m.Totals[0].Successful {
		t.Fatal("saving discarded the aborted status")
	}
}
