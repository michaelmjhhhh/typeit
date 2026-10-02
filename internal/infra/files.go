package infra

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/michaelmjhhhh/typeit/internal/domain"
)

type Language struct {
	Name       string   `json:"name"`
	Display    string   `json:"display"`
	Extensions []string `json:"extensions"`
	Aliases    []string `json:"aliases"`
}

//go:embed languages.json
var languageJSON []byte

//go:embed excludes.json
var excludeJSON []byte

func Languages() []Language {
	var langs []Language
	_ = json.Unmarshal(languageJSON, &langs)
	return langs
}
func ResolveLanguage(name string) (Language, error) {
	for _, l := range Languages() {
		for _, a := range append([]string{l.Name, l.Display}, l.Aliases...) {
			if strings.EqualFold(name, a) {
				return l, nil
			}
		}
	}
	return Language{}, fmt.Errorf("unsupported language %q", name)
}
func FileLanguage(path string) string {
	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	for _, l := range Languages() {
		for _, e := range l.Extensions {
			if e == ext {
				return l.Name
			}
		}
	}
	return ""
}
func Scan(ctx context.Context, root string, filter []string) ([]domain.Challenge, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	allowed := map[string]bool{}
	for _, name := range filter {
		l, err := ResolveLanguage(name)
		if err != nil {
			return nil, err
		}
		allowed[l.Name] = true
	}
	var excludes []string
	if err = json.Unmarshal(excludeJSON, &excludes); err != nil {
		return nil, err
	}
	var patterns []gitignore.Pattern
	for _, p := range excludes {
		patterns = append(patterns, gitignore.ParsePattern(p, nil))
	}
	base, sources := ignoreSources(ctx, root)
	for _, file := range sources {
		loaded, err := readIgnore(file)
		if err != nil {
			return nil, err
		}
		patterns = append(patterns, loaded...)
	}
	matchers := map[string][]gitignore.Pattern{}
	var result []domain.Challenge
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		parts := strings.Split(filepath.ToSlash(rel), "/")
		active := append([]gitignore.Pattern{}, patterns...)
		dir := filepath.Dir(rel)
		if entry.IsDir() {
			dir = rel
		}
		prefix := "."
		active = append(active, matchers[prefix]...)
		if dir != "." {
			for _, component := range strings.Split(filepath.ToSlash(dir), "/") {
				prefix = filepath.Join(prefix, component)
				active = append(active, matchers[prefix]...)
			}
		}
		if rel != "." && (parts[0] == ".git" || gitignore.NewMatcher(active).Match(ignoreDomain(base, path), entry.IsDir())) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			var local []gitignore.Pattern
			for _, name := range []string{".gitignore", ".ignore", ".gittypeignore"} {
				loaded, err := readIgnore(ignoreFile{filepath.Join(path, name), ignoreDomain(base, path)})
				if err != nil {
					return err
				}
				local = append(local, loaded...)
			}
			matchers[rel] = local
			return nil
		}
		lang := FileLanguage(path)
		if lang == "" || (len(allowed) > 0 && !allowed[lang]) {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Size() > 1024*1024 {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !utf8.Valid(b) || strings.IndexByte(string(b), 0) >= 0 {
			return nil
		}
		challenges, err := Extract(filepath.ToSlash(rel), lang, string(b))
		if err != nil {
			return fmt.Errorf("%s: %w", rel, err)
		}
		result = append(result, challenges...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no supported source code found in %s", root)
	}
	return result, nil
}
