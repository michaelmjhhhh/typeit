package gittype_test

import (
	"strings"
	"testing"
	"time"

	"github.com/michaelmjhhhh/typeit/internal/domain"
	"github.com/michaelmjhhhh/typeit/internal/infra"
)

func TestDatabaseRoundTrip(t *testing.T) {
	db, err := infra.OpenDatabase(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	c := domain.Generate(domain.Challenge{Code: languageSamples["go"], Path: "main.go", Language: "go", StartLine: 1, EndLine: 5}, true)[0]
	s := domain.Summarize([]domain.StageResult{{Challenge: c, Metrics: domain.Calculate(100, 2, 30000, false), Streaks: []int{10, 88}}})
	s.Repository = domain.Repository{Owner: "local", Name: "test"}
	s.StartedAt = time.Now().UTC().Format(time.RFC3339Nano)
	s.Difficulty = domain.Easy
	s.Mode = "Normal"
	if err = db.Save(&s); err != nil {
		t.Fatal(err)
	}
	history, err := db.History()
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 1 || history[0].Score != s.Score || history[0].ID != s.ID {
		t.Fatalf("%+v", history)
	}
	stages, err := db.Stages(s.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(stages) != 1 || stages[0].Challenge.Code != c.Code || len(stages[0].Streaks) != 2 {
		t.Fatalf("%+v", stages)
	}
}
func TestThemeConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	themes, err := infra.Themes(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(themes) != 16 {
		t.Fatalf("%d themes", len(themes))
	}
	c, err := infra.LoadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	c.Theme.ID = "aurora"
	c.Theme.Mode = "Light"
	if err = infra.SaveConfig(dir, c); err != nil {
		t.Fatal(err)
	}
	got, err := infra.LoadConfig(dir)
	if err != nil || got != c {
		t.Fatalf("%+v %v", got, err)
	}
}
func TestRustDatabaseCompatibility(t *testing.T) {
	db, err := infra.OpenDatabase(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := domain.Summarize([]domain.StageResult{{Metrics: domain.Calculate(100, 5, 10000, false), Challenge: domain.Challenge{ID: "compat", Code: "main()", Language: "go"}}})
	s.Repository = domain.Repository{Owner: "local", Name: "compat"}
	s.StartedAt = time.Now().UTC().Format(time.RFC3339Nano)
	if err = db.Save(&s); err != nil {
		t.Fatal(err)
	}
	var stamp string
	var tier, pos, total int
	var best float64
	if err = db.DB.QueryRow("SELECT CAST(s.started_at AS TEXT),sr.rank_position,sr.position,sr.total,sr.best_stage_wpm FROM sessions s JOIN session_results sr ON sr.session_id=s.id WHERE s.id=?", s.ID).Scan(&stamp, &tier, &pos, &total, &best); err != nil {
		t.Fatal(err)
	}
	if _, err = time.Parse("2006-01-02 15:04:05", stamp); err != nil {
		t.Fatal("Rust cannot read timestamp", err)
	}
	if tier < 1 || pos < 1 || total != 63 || best <= 0 {
		t.Fatal("missing rank metadata")
	}
}
func TestOriginalArtworkCoverage(t *testing.T) {
	for _, r := range domain.Ranks {
		if infra.Artwork("ranks.json", r.Name) == "" {
			t.Errorf("no artwork for %s", r.Name)
		}
	}
	if len(strings.Split(infra.Artwork("logo.json", ""), "\n")) != 6 {
		t.Fatal("logo does not match original six lines")
	}
}
