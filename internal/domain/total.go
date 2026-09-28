package domain

type TotalResult struct {
	Metrics
	SessionsAttempted, SessionsCompleted, StagesAttempted, StagesCompleted, StagesSkipped int
}

func Total(sessions []SessionResult) TotalResult {
	total := TotalResult{SessionsAttempted: len(sessions)}
	for _, s := range sessions {
		if s.Successful {
			total.SessionsCompleted++
		}
		total.StagesAttempted += s.Attempted
		total.StagesCompleted += s.Completed
		total.StagesSkipped += s.Skipped
		total.Keystrokes += s.Keystrokes + s.PartialKeys
		total.Mistakes += s.Mistakes + s.PartialMistakes
		total.DurationMS += s.DurationMS
		if s.Completed > 0 {
			total.Score += s.Score
		}
	}
	if total.DurationMS > 0 {
		total.CPM = float64(total.Keystrokes) * 60000 / float64(total.DurationMS)
		total.WPM = total.CPM / 5
	}
	if total.Keystrokes > 0 {
		total.Accuracy = float64(max(0, total.Keystrokes-total.Mistakes)) / float64(total.Keystrokes) * 100
	}
	r := RankFor(total.Score)
	total.RankName = r.Name
	total.TierName = r.Tier
	return total
}
func RankPositions(score float64) (tierPosition, tierTotal, position, total int) {
	rank := RankFor(score)
	total = len(Ranks)
	for i, r := range Ranks {
		if r.Tier == rank.Tier {
			tierTotal++
			if r.Min >= rank.Min {
				tierPosition++
			}
		}
		if r.Name == rank.Name {
			position = total - i
		}
	}
	return
}
