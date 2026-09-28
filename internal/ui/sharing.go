package ui

import (
	"fmt"
	"net/url"

	"github.com/michaelmjhhhh/typeit/internal/domain"
)

func ShareText(s domain.SessionResult) string {
	return fmt.Sprintf("Achieved %q with %.0fpts on [%s/%s] in gittype! CPM: %.0f, Mistakes: %d 🚀\n\nType your own code! https://github.com/unhappychoice/gittype\n\n#gittype #typing #coding", s.RankName, s.Score, s.Repository.Owner, s.Repository.Name, s.CPM, s.Mistakes+s.PartialMistakes)
}
func ShareURL(s domain.SessionResult, platform int) string {
	text := url.QueryEscape(ShareText(s))
	switch platform {
	case 1:
		return "https://www.reddit.com/submit?title=" + url.QueryEscape(fmt.Sprintf("Achieved %s rank with %.0f points in gittype!", s.RankName, s.Score)) + "&selftext=true&text=" + text
	case 2:
		return "https://www.linkedin.com/feed/?shareActive=true&mini=true&text=" + text
	case 3:
		return "https://www.facebook.com/sharer/sharer.php?u=" + url.QueryEscape("https://github.com/unhappychoice/gittype") + "&quote=" + text
	default:
		return "https://x.com/intent/tweet?text=" + text
	}
}
func (m *Model) shareView() string {
	return m.heading("SHARE RESULT") + "\n\n" + ShareText(m.Summary) + "\n\n1: X   2: Reddit   3: LinkedIn   4: Facebook\nOpens a draft in your browser. Esc: back\n\n" + m.Status
}
