package gittype_test

import (
	"math"
	"testing"
	"time"

	"github.com/michaelmjhhhh/typeit/internal/domain"
)

func TestTypingUnicodeCommentsIndentation(t *testing.T) {
	code := "  // 説明\n\tfunc café() {\n    println(\"你好🌍\") // note\n\n  }\n"
	chars := []rune(code)
	var comments []domain.Range
	for i := 0; i < len(chars)-1; i++ {
		if chars[i] == '/' && chars[i+1] == '/' {
			end := i
			for end < len(chars) && chars[end] != '\n' {
				end++
			}
			comments = append(comments, domain.Range{i, end})
		}
	}
	core := domain.NewTyping(domain.Challenge{Code: code, Comments: comments})
	expected := "func café() {\nprintln(\"你好🌍\")\n}"
	if string(core.Text) != expected {
		t.Fatalf("text: %q", string(core.Text))
	}
	if core.Input('X') || core.Position != 0 || core.Mistakes != 1 || !core.Wrong {
		t.Fatal("mistake changed progress")
	}
	for _, r := range expected {
		if !core.Input(r) {
			t.Fatalf("rejected %q at %d", r, core.Position)
		}
	}
	if !core.Completed() || core.Wrong {
		t.Fatal("did not complete")
	}
	if core.Input('x') || core.Mistakes != 1 {
		t.Fatal("completed input changed stats")
	}
}
func TestScoreGolden(t *testing.T) {
	cases := []struct {
		cpm, accuracy float64
		mistakes      int
		seconds       float64
		chars         int
		want          float64
	}{
		{300, 100, 0, 20, 100, 10300}, {300, 90, 10, 20, 100, 8100}, {300, 70, 30, 20, 100, 4000}, {600, 100, 0, 5, 100, 20700}, {0, 0, 0, 0, 0, 100},
	}
	for _, c := range cases {
		got := domain.Score(c.cpm, c.accuracy, c.mistakes, c.seconds, c.chars)
		if math.Abs(got-c.want) > 1e-6 {
			t.Errorf("%+v: got %f", c, got)
		}
	}
}
func TestAllRankBoundaries(t *testing.T) {
	if len(domain.Ranks) != 63 {
		t.Fatalf("got %d ranks", len(domain.Ranks))
	}
	for i, r := range domain.Ranks {
		for _, s := range []float64{float64(r.Min), float64(r.Max), float64(r.Max) + 0.99} {
			if got := domain.RankFor(s); got.Name != r.Name {
				t.Errorf("score %f: %s != %s", s, got.Name, r.Name)
			}
		}
		if i > 0 && r.Min != domain.Ranks[i-1].Max+1 {
			t.Fatal("rank gap")
		}
	}
}
func TestSessionExcludesPartialEffort(t *testing.T) {
	valid := domain.StageResult{Metrics: domain.Calculate(100, 10, 20000, false)}
	skipped := domain.StageResult{Metrics: domain.Calculate(50, 20, 30000, false), Skipped: true}
	s := domain.Summarize([]domain.StageResult{valid, skipped})
	if s.Completed != 1 || s.Skipped != 1 || s.PartialKeys != 50 || s.Keystrokes != 100 || s.DurationMS != 50000 || s.CPM != 300 || s.Accuracy != 90 {
		t.Fatalf("%+v", s)
	}
	if s.Score != domain.Score(300, 90, 10, 20, 100) {
		t.Fatal("partial effort affected score")
	}
}
func TestDifficultyGeneration(t *testing.T) {
	c := domain.Challenge{Path: "x.go", Code: "func greet() {\n    println(\"hello, wonderful world\");\n}\n", StartLine: 4, EndLine: 6, Language: "go"}
	chunks := domain.Generate(c, false)
	seen := map[domain.Difficulty]bool{}
	for _, v := range chunks {
		seen[v.Difficulty] = true
		if v.StartLine != 4 {
			t.Fatal("source line lost")
		}
		if v.ID == "" {
			t.Fatal("empty ID")
		}
	}
	if !seen[domain.Easy] || !seen[domain.Wild] || seen[domain.Zen] || seen[domain.Hard] {
		t.Fatalf("%v", seen)
	}
	for _, v := range domain.Generate(c, true) {
		seen[v.Difficulty] = true
	}
	if !seen[domain.Zen] {
		t.Fatal("file has no zen challenge")
	}
}
func TestTotalSumsSessionScoresAndPartialEffort(t *testing.T) {
	a := domain.SessionResult{Metrics: domain.Metrics{Keystrokes: 100, Mistakes: 5, DurationMS: 30000, Score: 1000}, Completed: 1, PartialKeys: 20, PartialMistakes: 10, Successful: true}
	b := a
	b.Score = 2000
	total := domain.Total([]domain.SessionResult{a, b})
	if total.Score != 3000 || total.Keystrokes != 240 || total.Mistakes != 30 || total.CPM != 240 || total.SessionsCompleted != 2 {
		t.Fatalf("%+v", total)
	}
}
func TestAnalyticsGroupsAndDateWindow(t *testing.T) {
	now := time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC)
	metric := domain.Metrics{CPM: 300, WPM: 60, Accuracy: 95, Keystrokes: 100, Mistakes: 5, DurationMS: 20000, Score: 1000}
	recent := domain.SessionResult{StartedAt: now.Format(time.RFC3339), Metrics: metric, Repository: domain.Repository{Owner: "owner", Name: "repo"}, Stages: []domain.StageResult{{Metrics: metric, Challenge: domain.Challenge{Language: "go"}}, {Metrics: metric, Challenge: domain.Challenge{Language: "rust"}, Skipped: true}}}
	old := recent
	old.StartedAt = now.AddDate(0, 0, -100).Format(time.RFC3339)
	a := domain.Analyze([]domain.SessionResult{recent, old}, now)
	if a.Sessions != 1 || len(a.Languages) != 1 || a.Languages[0].Name != "go" || a.CPM != 300 || a.Streak != 1 {
		t.Fatalf("%+v", a)
	}
}

func TestSubsecondSessionScoreFloor(t *testing.T) {
	stage := domain.Calculate(100, 0, 0, false)
	session := domain.Calculate(100, 0, 0, true)
	if stage.Score != 2040500 || session.Score != 2040496 {
		t.Fatalf("duration floor: stage=%v session=%v", stage.Score, session.Score)
	}
}
