package domain

import "math"

type Metrics struct {
	Keystrokes int     `json:"keystrokes"`
	Mistakes   int     `json:"mistakes"`
	DurationMS int64   `json:"duration_ms"`
	WPM        float64 `json:"wpm"`
	CPM        float64 `json:"cpm"`
	Accuracy   float64 `json:"accuracy"`
	Score      float64 `json:"score"`
	RankName   string  `json:"rank_name"`
	TierName   string  `json:"tier_name"`
}
type StageResult struct {
	Metrics
	Challenge Challenge `json:"challenge"`
	Streaks   []int     `json:"consistency_streaks"`
	Skipped   bool      `json:"was_skipped"`
	Failed    bool      `json:"was_failed"`
}
type SessionResult struct {
	ID int64 `json:"id"`
	Metrics
	Repository      Repository    `json:"repository"`
	StartedAt       string        `json:"started_at"`
	Difficulty      Difficulty    `json:"difficulty_level"`
	Mode            string        `json:"game_mode"`
	Stages          []StageResult `json:"stages"`
	Completed       int           `json:"stages_completed"`
	Attempted       int           `json:"stages_attempted"`
	Skipped         int           `json:"stages_skipped"`
	PartialKeys     int           `json:"partial_effort_keystrokes"`
	PartialMistakes int           `json:"partial_effort_mistakes"`
	Successful      bool          `json:"successful"`
}
type Repository struct {
	Owner  string `json:"user_name"`
	Name   string `json:"repository_name"`
	URL    string `json:"remote_url"`
	Path   string `json:"path"`
	Branch string `json:"branch"`
	Commit string `json:"commit_hash"`
	Dirty  bool   `json:"is_dirty"`
}

func Score(cpm, accuracy float64, mistakes int, seconds float64, chars int) float64 {
	base := cpm * (accuracy / 100) * 10
	a := math.Max(0, math.Min(100, accuracy)) / 100
	factor := 0.0
	if a > 0.7 && a < 0.9 {
		t := (a - 0.7) / 0.2
		factor = 0.5 * t * t * (3 - 2*t)
	} else if a >= 0.9 && a < 0.95 {
		factor = 0.5
	} else if a >= 0.95 {
		t := (a - 0.95) / 0.05
		factor = 0.5 + 0.2*t*t*(3-2*t)
	}
	bonus := 0.0
	if chars > 50 {
		bonus = math.Max(0, float64(chars)/10-seconds) * 20
	}
	return math.Max(0, (base+base*factor+bonus-float64(mistakes)*5)*2+100)
}
func Calculate(keys, mistakes int, millis int64, session bool) Metrics {
	m := Metrics{Keystrokes: keys, Mistakes: mistakes, DurationMS: millis}
	seconds := math.Max(0.1, float64(millis)/1000)
	if keys > 0 {
		m.Accuracy = float64(max(0, keys-mistakes)) / float64(keys) * 100
		count := keys - mistakes
		if session {
			count = keys
		}
		m.CPM = float64(count) / seconds * 60
	} else if !session {
		m.CPM = 0.1
	}
	m.WPM = m.CPM / 5
	scoreSeconds := float64(millis) / 1000
	if session {
		scoreSeconds = seconds
	}
	m.Score = Score(m.CPM, m.Accuracy, mistakes, scoreSeconds, keys)
	if session && keys == 0 {
		m.Score = 0
	}
	r := RankFor(m.Score)
	m.RankName = r.Name
	m.TierName = r.Tier
	return m
}
func Summarize(stages []StageResult) SessionResult {
	s := SessionResult{Stages: stages, Successful: true, Attempted: len(stages)}
	keys, errors := 0, 0
	var valid, total int64
	for _, r := range stages {
		total += r.DurationMS
		if r.Failed {
			s.Successful = false
		}
		if r.Skipped {
			s.Skipped++
		}
		if r.Failed || r.Skipped {
			s.PartialKeys += r.Keystrokes
			s.PartialMistakes += r.Mistakes
			continue
		}
		s.Completed++
		keys += r.Keystrokes
		errors += r.Mistakes
		valid += r.DurationMS
	}
	s.Metrics = Calculate(keys, errors, valid, true)
	s.DurationMS = total
	if valid < 1000 {
		s.CPM = 0
		s.WPM = 0
		s.Accuracy = 0
	}
	return s
}
