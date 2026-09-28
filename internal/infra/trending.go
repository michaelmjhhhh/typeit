package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

type TrendingRepo struct {
	Name        string `json:"repo_name"`
	Language    string `json:"primary_language"`
	Description string `json:"description"`
	Stars       string `json:"stars"`
	Forks       string `json:"forks"`
	Score       string `json:"total_score"`
}
type TrendingCache struct {
	Fetched      time.Time      `json:"fetched"`
	Repositories []TrendingRepo `json:"repositories"`
}

func Trending(ctx context.Context, dir, language, period string) ([]TrendingRepo, error) {
	periods := map[string]string{"daily": "past_24_hours", "weekly": "past_week", "monthly": "past_month"}
	p, ok := periods[period]
	if !ok {
		return nil, fmt.Errorf("invalid period %q: use daily, weekly or monthly", period)
	}
	canonical := ""
	display := ""
	if language != "" {
		l, err := ResolveLanguage(language)
		if err != nil {
			return nil, err
		}
		canonical = l.Name
		display = l.Display
	}
	cache := filepath.Join(dir, "trending", canonical+"-"+period+".json")
	var old TrendingCache
	if b, err := os.ReadFile(cache); err == nil {
		if json.Unmarshal(b, &old) == nil && time.Since(old.Fetched) < 24*time.Hour {
			return old.Repositories, nil
		}
	}
	endpoint := "https://api.ossinsight.io/v1/trends/repos/?period=" + p
	if display != "" {
		endpoint += "&language=" + url.QueryEscape(display)
	}
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gittype")
	req.Header.Set("Accept", "application/json")
	response, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("trending service: %s", response.Status)
	}
	var body struct {
		Data struct {
			Rows []TrendingRepo `json:"rows"`
		} `json:"data"`
	}
	if err = json.NewDecoder(io.LimitReader(response.Body, 8<<20)).Decode(&body); err != nil {
		return nil, err
	}
	if err = WriteJSON(cache, TrendingCache{Fetched: time.Now(), Repositories: body.Data.Rows}); err != nil {
		return nil, err
	}
	return body.Data.Rows, nil
}
