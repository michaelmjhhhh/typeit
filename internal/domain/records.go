package domain

import "time"

type BestRecords struct {
	Today, Weekly, AllTime          float64
	HasToday, HasWeekly, HasAllTime bool
}

func Records(sessions []SessionResult, now time.Time) BestRecords {
	r := BestRecords{}
	for _, s := range sessions {
		r.AllTime = max(r.AllTime, s.Score)
		r.HasAllTime = true
		t, err := time.Parse(time.RFC3339Nano, s.StartedAt)
		if err != nil {
			t, _ = time.Parse("2006-01-02 15:04:05", s.StartedAt)
		}
		if t.After(now.AddDate(0, 0, -7)) {
			r.Weekly = max(r.Weekly, s.Score)
			r.HasWeekly = true
		}
		if t.Format("2006-01-02") == now.Format("2006-01-02") {
			r.Today = max(r.Today, s.Score)
			r.HasToday = true
		}
	}
	return r
}
func (r BestRecords) Status(score float64) string {
	if r.HasAllTime && score >= r.AllTime {
		return "ALL TIME"
	}
	if r.HasWeekly && score >= r.Weekly {
		return "WEEKLY"
	}
	if !r.HasToday || score >= r.Today {
		return "TODAY'S"
	}
	return ""
}
