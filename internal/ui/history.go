package ui

import (
	"cmp"
	"slices"
	"strings"
	"time"

	"github.com/michaelmjhhhh/typeit/internal/domain"
)

func (m *Model) historyRows() []domain.SessionResult {
	rows := []domain.SessionResult{}
	days := []int{0, 7, 30, 90}[m.DateFilter]
	for _, s := range m.History {
		when, err := time.Parse(time.RFC3339Nano, s.StartedAt)
		if err != nil {
			when, _ = time.Parse("2006-01-02 15:04:05", s.StartedAt)
		}
		if days > 0 && when.Before(m.Now.AddDate(0, 0, -days)) {
			continue
		}
		rows = append(rows, s)
	}
	slices.SortStableFunc(rows, func(a, b domain.SessionResult) int {
		var result int
		switch m.SortBy {
		case 0:
			result = strings.Compare(a.StartedAt, b.StartedAt)
		case 1:
			result = cmp.Compare(a.Score, b.Score)
		case 2:
			result = strings.Compare(a.Repository.Name, b.Repository.Name)
		case 3:
			result = cmp.Compare(a.DurationMS, b.DurationMS)
		}
		if m.SortDescending {
			return -result
		}
		return result
	})
	return rows
}
