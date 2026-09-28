package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type VersionInfo struct {
	Latest    string    `json:"latest_version"`
	Current   string    `json:"current_version"`
	Available bool      `json:"update_available"`
	Checked   time.Time `json:"last_checked"`
}

func NewerVersion(latest, current string) bool {
	parse := func(s string) ([]int, bool) {
		var out []int
		for _, p := range strings.Split(s, ".") {
			v, err := strconv.Atoi(p)
			if err != nil || v < 0 {
				return nil, false
			}
			out = append(out, v)
		}
		return out, true
	}
	a, ok := parse(latest)
	if !ok {
		return false
	}
	b, ok := parse(current)
	if !ok {
		return false
	}
	for i := 0; i < min(len(a), len(b)); i++ {
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}
	return len(a) > len(b)
}
func CheckVersion(ctx context.Context, dir, current string) (VersionInfo, error) {
	path := filepath.Join(dir, "version_cache.json")
	var cached VersionInfo
	b, err := os.ReadFile(path)
	if err == nil && json.Unmarshal(b, &cached) == nil && cached.Current == current && time.Since(cached.Checked) < 24*time.Hour {
		return cached, nil
	}
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/repos/unhappychoice/gittype/releases/latest", nil)
	if err != nil {
		return cached, err
	}
	req.Header.Set("User-Agent", "gittype")
	req.Header.Set("Accept", "application/vnd.github+json")
	response, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	if err != nil {
		if cached.Latest != "" {
			return cached, nil
		}
		return cached, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return cached, fmt.Errorf("version check: %s", response.Status)
	}
	var body struct {
		Tag string `json:"tag_name"`
	}
	if err = json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&body); err != nil {
		return cached, err
	}
	latest := strings.TrimPrefix(body.Tag, "v")
	result := VersionInfo{Latest: latest, Current: current, Available: NewerVersion(latest, current), Checked: time.Now().UTC()}
	return result, WriteJSON(path, result)
}
