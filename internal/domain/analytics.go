package domain

import (
	"sort"
	"time"
)

type GroupStats struct {
	Name                                                 string
	Count, Keys, Mistakes, Completed, Attempted, Skipped int
	DurationMS                                           int64
	CPM, WPM, Accuracy, Score, BestCPM, BestAccuracy     float64
}
type DailyStats struct {
	Date          string
	Sessions      int
	CPM, Accuracy float64
}
type Analytics struct {
	Sessions                                      int
	CPM, Accuracy, BestCPM, Hours, AverageMinutes float64
	Mistakes, Streak                              int
	Repositories, Languages                       []GroupStats
	Days                                          []DailyStats
}

func Analyze(sessions []SessionResult, now time.Time) Analytics {
	a := Analytics{}
	repos := map[string]*GroupStats{}
	languages := map[string]*GroupStats{}
	daily := map[string]*DailyStats{}
	var duration int64
	for _, s := range sessions {
		when, err := time.Parse(time.RFC3339Nano, s.StartedAt)
		if err != nil {
			when, err = time.Parse("2006-01-02 15:04:05", s.StartedAt)
		}
		if err != nil || when.Before(now.AddDate(0, 0, -90)) {
			continue
		}
		a.Sessions++
		a.CPM += s.CPM
		a.Accuracy += s.Accuracy
		a.BestCPM = max(a.BestCPM, s.CPM)
		a.Mistakes += s.Mistakes + s.PartialMistakes
		duration += s.DurationMS
		date := when.Format("2006-01-02")
		if daily[date] == nil {
			daily[date] = &DailyStats{Date: date}
		}
		d := daily[date]
		d.Sessions++
		d.CPM += s.CPM
		d.Accuracy += s.Accuracy
		key := s.Repository.Owner + "/" + s.Repository.Name
		if repos[key] == nil {
			repos[key] = &GroupStats{Name: key}
		}
		g := repos[key]
		addMetrics(g, s.Metrics)
		g.Completed += s.Completed
		g.Attempted += s.Attempted
		g.Skipped += s.Skipped
		for _, st := range s.Stages {
			if st.Skipped || st.Failed || st.Challenge.Language == "" {
				continue
			}
			key = st.Challenge.Language
			if languages[key] == nil {
				languages[key] = &GroupStats{Name: key}
			}
			g = languages[key]
			addMetrics(g, st.Metrics)
			g.Completed++
			g.Attempted++
		}
	}
	if a.Sessions > 0 {
		a.CPM /= float64(a.Sessions)
		a.Accuracy /= float64(a.Sessions)
		a.AverageMinutes = float64(duration) / 60000 / float64(a.Sessions)
	}
	a.Hours = float64(duration) / 3600000
	for _, d := range daily {
		d.CPM /= float64(d.Sessions)
		d.Accuracy /= float64(d.Sessions)
		a.Days = append(a.Days, *d)
	}
	sort.Slice(a.Days, func(i, j int) bool { return a.Days[i].Date < a.Days[j].Date })
	for day := now; ; day = day.AddDate(0, 0, -1) {
		if daily[day.Format("2006-01-02")] == nil {
			break
		}
		a.Streak++
	}
	a.Repositories = finishGroups(repos)
	a.Languages = finishGroups(languages)
	return a
}
func addMetrics(g *GroupStats, m Metrics) {
	g.Count++
	g.CPM += m.CPM
	g.WPM += m.WPM
	g.Accuracy += m.Accuracy
	g.Score += m.Score
	g.Keys += m.Keystrokes
	g.Mistakes += m.Mistakes
	g.DurationMS += m.DurationMS
	g.BestCPM = max(g.BestCPM, m.CPM)
	g.BestAccuracy = max(g.BestAccuracy, m.Accuracy)
}
func finishGroups(groups map[string]*GroupStats) []GroupStats {
	out := []GroupStats{}
	for _, g := range groups {
		n := float64(g.Count)
		g.CPM /= n
		g.WPM /= n
		g.Accuracy /= n
		g.Score /= n
		out = append(out, *g)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count == out[j].Count {
			return out[i].Name < out[j].Name
		}
		return out[i].Count > out[j].Count
	})
	return out
}
