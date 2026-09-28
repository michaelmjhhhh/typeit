package infra

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/michaelmjhhhh/typeit/assets"
)

type RGB struct {
	Name string `json:"-"`
	R    int    `json:"r"`
	G    int    `json:"g"`
	B    int    `json:"b"`
}

func (c RGB) MarshalJSON() ([]byte, error) {
	if c.Name != "" {
		return json.Marshal(c.Name)
	}
	type plain RGB
	return json.Marshal(plain(c))
}
func (c *RGB) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '"' {
		return json.Unmarshal(b, &c.Name)
	}
	type plain RGB
	return json.Unmarshal(b, (*plain)(c))
}
func (c RGB) Hex() string {
	if c.Name != "" {
		names := map[string]string{"black": "0", "red": "1", "green": "2", "yellow": "3", "blue": "4", "magenta": "5", "cyan": "6", "gray": "7", "light_gray": "7", "dark_gray": "8", "light_red": "9", "light_green": "10", "light_yellow": "11", "light_blue": "12", "light_magenta": "13", "light_cyan": "14", "white": "15", "reset": ""}
		if value, ok := names[c.Name]; ok {
			return value
		}
		return c.Name
	}
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

type Theme struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Dark        map[string]RGB `json:"dark"`
	Light       map[string]RGB `json:"light"`
}

func Themes(dir string) ([]Theme, error) {
	entries, err := assets.Files.ReadDir("themes")
	if err != nil {
		return nil, err
	}
	byID := map[string]Theme{}
	for _, entry := range entries {
		b, err := assets.Files.ReadFile("themes/" + entry.Name())
		if err != nil {
			return nil, err
		}
		var t Theme
		if err = json.Unmarshal(b, &t); err != nil {
			return nil, err
		}
		byID[t.ID] = t
	}
	customPath := filepath.Join(dir, "custom-theme.json")
	if _, e := os.Stat(customPath); os.IsNotExist(e) {
		base := byID["default"]
		if e = WriteJSON(customPath, map[string]any{"dark": base.Dark, "light": base.Light}); e != nil {
			return nil, e
		}
	}
	if b, e := os.ReadFile(customPath); e == nil {
		var t Theme
		if e = json.Unmarshal(b, &t); e != nil {
			return nil, e
		}
		t.ID = "custom"
		t.Name = "Custom"
		t.Description = "Your personal custom theme — edit ~/.gittype/custom-theme.json"
		byID[t.ID] = t
	}
	files, err := filepath.Glob(filepath.Join(dir, "themes", "*.json"))
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		b, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		var t Theme
		if err = json.Unmarshal(b, &t); err != nil {
			return nil, fmt.Errorf("theme %s: %w", file, err)
		}
		if t.ID == "" || t.Name == "" || len(t.Dark) == 0 || len(t.Light) == 0 {
			return nil, fmt.Errorf("invalid theme %s", file)
		}
		byID[t.ID] = t
	}
	for id, t := range byID {
		base := byID["default"]
		if t.Dark == nil {
			t.Dark = map[string]RGB{}
		}
		if t.Light == nil {
			t.Light = map[string]RGB{}
		}
		for k, v := range base.Dark {
			if _, ok := t.Dark[k]; !ok {
				t.Dark[k] = v
			}
		}
		for k, v := range base.Light {
			if _, ok := t.Light[k]; !ok {
				t.Light[k] = v
			}
		}
		byID[id] = t
	}
	out := make([]Theme, 0, len(byID))
	for _, t := range byID {
		out = append(out, t)
	}
	order := []string{"default", "original", "ascii", "aurora", "blood_oath", "cyber_void", "eclipse", "glacier", "inferno", "neon_abyss", "oblivion", "runic", "spectral", "starforge", "venom", "custom"}
	indices := map[string]int{}
	for i, id := range order {
		indices[id] = i
	}
	sort.Slice(out, func(i, j int) bool {
		a, ok := indices[out[i].ID]
		if !ok {
			a = 100
		}
		b, ok := indices[out[j].ID]
		if !ok {
			b = 100
		}
		if a == b {
			return out[i].Name < out[j].Name
		}
		return a < b
	})
	return out, nil
}
func Artwork(file, key string) string {
	b, err := assets.Files.ReadFile(file)
	if err != nil {
		return ""
	}
	var lines []string
	if key == "" {
		_ = json.Unmarshal(b, &lines)
	} else {
		var all map[string][]string
		_ = json.Unmarshal(b, &all)
		lines = all[key]
	}
	s := ""
	for i, line := range lines {
		if i > 0 {
			s += "\n"
		}
		s += line
	}
	return s
}
