package infra

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Theme ThemeConfig `json:"theme"`
}
type ThemeConfig struct {
	ID   string `json:"current_theme_id"`
	Mode string `json:"current_color_mode"`
}

func DataDir() (string, error) {
	if dir := os.Getenv("TYPEIT_DATA_DIR"); dir != "" {
		return filepath.Abs(dir)
	}
	if dir := os.Getenv("GITTYPE_DATA_DIR"); dir != "" {
		return filepath.Abs(dir)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".typeit")
	if err := migrateLegacyData(filepath.Join(home, ".gittype"), dir); err != nil {
		return "", err
	}
	return dir, nil
}
func LoadConfig(dir string) (Config, error) {
	c := Config{Theme: ThemeConfig{ID: "charm", Mode: "Dark"}}
	b, err := os.ReadFile(filepath.Join(dir, "config.json"))
	if os.IsNotExist(err) {
		return c, nil
	}
	if err != nil {
		return c, err
	}
	if err = json.Unmarshal(b, &c); err != nil {
		return c, fmt.Errorf("read config: %w", err)
	}
	return c, nil
}
func SaveConfig(dir string, c Config) error { return WriteJSON(filepath.Join(dir, "config.json"), c) }
func WriteJSON(path string, value any) error {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return AtomicWrite(path, append(b, '\n'))
}
func AtomicWrite(path string, b []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".typeit-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}
