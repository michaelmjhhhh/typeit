package gittype_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/michaelmjhhhh/typeit/internal/domain"
	"github.com/michaelmjhhhh/typeit/internal/infra"
)

func TestOriginalTypingSnapshots(t *testing.T) {
	files, err := filepath.Glob("../upstream/*typing_core_common__*.snap")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) < 100 {
		t.Fatalf("only %d snapshots", len(files))
	}
	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			b, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			text := string(b)
			_, text, ok := strings.Cut(text, "# text_original\n")
			if !ok {
				t.Fatal("missing original")
			}
			source, rest, ok := strings.Cut(text, "\n\n# text_to_type\n")
			if !ok {
				t.Fatal("missing typing section")
			}
			want, display, ok := strings.Cut(rest, "\n\n# text_to_display\n")
			if !ok {
				t.Fatal("missing display section")
			}
			name := strings.Split(filepath.Base(file), "typing_core_common__")[1]
			lang := strings.Split(name, "_")[0]
			challenges, err := infra.Extract("sample", lang, source)
			if err != nil {
				t.Fatal(err)
			}
			for _, c := range challenges {
				if c.Difficulty == domain.Zen {
					core := domain.NewTyping(c)
					if got := string(core.Display); got != strings.TrimSuffix(display, "\n") {
						t.Fatalf("display mismatch\nwant %q\n got %q", strings.TrimSuffix(display, "\n"), got)
					}
					if got := string(core.Text); got != want {
						t.Fatalf("typing mismatch\nwant %q\n got %q", want, got)
					}
					return
				}
			}
			t.Fatal("missing zen challenge")
		})
	}
}
func TestOriginalExtractionSnapshots(t *testing.T) {
	files, err := filepath.Glob("../upstream/*extractor__*.snap")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) < 100 {
		t.Fatalf("only %d snapshots", len(files))
	}
	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			b, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			start := strings.Index(string(b), "{\n")
			if start < 0 {
				t.Fatal("missing json")
			}
			var snapshot struct {
				Chunks []struct {
					Type     string         `json:"chunk_type"`
					Code     string         `json:"content"`
					Language string         `json:"language"`
					Start    int            `json:"start_line"`
					End      int            `json:"end_line"`
					Comments []domain.Range `json:"comment_ranges"`
				} `json:"chunks"`
			}
			if err = json.Unmarshal(b[start:], &snapshot); err != nil {
				t.Fatal(err)
			}
			source, language := "", ""
			for _, c := range snapshot.Chunks {
				if c.Type == "File" {
					source = c.Code
					language = c.Language
					break
				}
			}
			if source == "" {
				t.Skip("snapshot has no original file")
			}
			got, err := infra.Extract("sample", language, source)
			if err != nil {
				t.Fatal(err)
			}
			actual := map[string]domain.Challenge{}
			count := 0
			for _, c := range got {
				if c.Difficulty == domain.Wild {
					count++
					actual[fmt.Sprintf("%d:%d:%s", c.StartLine, c.EndLine, c.Code)] = c
				}
			}
			if count != len(snapshot.Chunks) {
				t.Errorf("chunk count got %d want %d", count, len(snapshot.Chunks))
			}
			for _, c := range snapshot.Chunks {
				v, ok := actual[fmt.Sprintf("%d:%d:%s", c.Start, c.End, c.Code)]
				if !ok {
					t.Errorf("missing %s chunk at %d-%d: %.100q", c.Type, c.Start, c.End, c.Code)
					continue
				}
				if !slices.Equal(v.Comments, c.Comments) {
					t.Errorf("comment ranges at %d: got %v want %v", c.Start, v.Comments, c.Comments)
				}
				if v.StartLine != c.Start || v.EndLine != c.End {
					t.Errorf("source lines: got %d-%d want %d-%d", v.StartLine, v.EndLine, c.Start, c.End)
				}
			}
		})
	}
}
