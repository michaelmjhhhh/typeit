package gittype_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/michaelmjhhhh/typeit/internal/infra"
	"github.com/michaelmjhhhh/typeit/internal/ui"
)

func TestCharmThemeCatalog(t *testing.T) {
	themes, err := infra.Themes(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"charm", "dracula", "catppuccin", "base16", "default", "custom"}
	var ids []string
	for _, theme := range themes {
		ids = append(ids, theme.ID)
		for _, palette := range []map[string]infra.RGB{theme.Dark, theme.Light} {
			for key := range themes[0].Dark {
				if _, ok := palette[key]; !ok {
					t.Errorf("%s missing %s", theme.ID, key)
				}
			}
			if palette["typing_untyped_text"] == palette["typing_typed_text"] || palette["typing_cursor_fg"] == palette["typing_cursor_bg"] {
				t.Errorf("%s has indistinguishable typing feedback", theme.ID)
			}
		}
	}
	if !reflect.DeepEqual(ids, want) {
		t.Fatalf("theme IDs: %v", ids)
	}
	if themes[2].Dark["background"].Hex() != "#1e1e2e" || themes[2].Light["background"].Hex() != "#eff1f5" {
		t.Fatal("Catppuccin must use Mocha and Latte")
	}
	if !reflect.DeepEqual(themes[1].Dark, themes[1].Light) {
		t.Fatal("Dracula must retain its upstream dark palette")
	}
}

func TestCustomThemesSurvivePresetReplacement(t *testing.T) {
	dir := t.TempDir()
	data := []byte(`{"dark":{"title":"#123456"},"light":{"title":"#654321"}}`)
	path := filepath.Join(dir, "custom-theme.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "themes"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "themes", "charm.json"), []byte(`{"id":"charm","name":"Personal Charm","dark":{"title":"#abcdef"},"light":{"title":"#fedcba"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	// A user-provided preset can retain an otherwise removed built-in ID.
	if err := os.WriteFile(filepath.Join(dir, "themes", "aurora.json"), []byte(`{"id":"aurora","name":"Personal Aurora","dark":{"title":"#123456"},"light":{"title":"#654321"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	themes, err := infra.Themes(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, theme := range themes {
		if _, ok := theme.Dark["status_error"]; !ok {
			t.Fatalf("%s lost built-in defaults after overriding Charm", theme.ID)
		}
		if theme.ID == "custom" || theme.ID == "aurora" {
			if theme.Dark["title"].Hex() != "#123456" || theme.Light["title"].Hex() != "#654321" || theme.Dark["status_error"] != themes[0].Dark["status_error"] {
				t.Fatalf("custom colors or fallback lost: %s", theme.ID)
			}
		}
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != string(data) {
		t.Fatal("existing custom file was changed", err)
	}
	m := ui.New(ui.Options{Themes: themes, Config: infra.Config{Theme: infra.ThemeConfig{ID: "aurora", Mode: "Light"}}})
	if m.Options.Themes[m.ThemeIndex].ID != "aurora" {
		t.Fatal("user theme was replaced by fallback")
	}
}

func TestRemovedThemeFallsBackAndSettingsSave(t *testing.T) {
	m := testModel(t)
	options := m.Options
	options.Config.Theme.ID = "aurora"
	m = ui.New(options)
	m.Screen = "title"
	if m.Options.Config.Theme.ID != "charm" || m.Options.Themes[m.ThemeIndex].ID != "charm" {
		t.Fatal("removed theme did not fall back to Charm")
	}
	send(m, "s")
	send(m, "right")
	send(m, "down")
	send(m, "esc")
	if m.Options.Themes[m.ThemeIndex].ID != "charm" {
		t.Fatal("cancel failed to restore Charm")
	}
	send(m, "s")
	send(m, "right")
	send(m, "down")
	cmd := send(m, " ")
	if cmd == nil {
		t.Fatal("missing save command")
	}
	m.Update(cmd())
	saved, err := infra.LoadConfig(options.DataDir)
	if err != nil || saved.Theme.ID != "dracula" {
		t.Fatal("replacement preset did not persist", err)
	}
	options.Config = saved
	reopened := ui.New(options)
	if reopened.Options.Themes[reopened.ThemeIndex].ID != "dracula" {
		t.Fatal("saved preset did not reopen")
	}
}
