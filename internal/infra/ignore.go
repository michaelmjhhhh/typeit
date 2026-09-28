package infra

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

type ignoreFile struct {
	Path   string
	Domain []string
}

func ignoreSources(ctx context.Context, root string) (string, []ignoreFile) {
	git := func(args ...string) string {
		b, err := exec.CommandContext(ctx, "git", append([]string{"-C", root}, args...)...).Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(b))
	}
	base := git("rev-parse", "--show-toplevel")
	if base == "" {
		base = root
	}
	files := []ignoreFile{}
	global := git("config", "--path", "--get", "core.excludesfile")
	if global == "" {
		config := os.Getenv("XDG_CONFIG_HOME")
		if config == "" {
			if home, err := os.UserHomeDir(); err == nil {
				config = filepath.Join(home, ".config")
			}
		}
		if config != "" {
			global = filepath.Join(config, "git", "ignore")
		}
	}
	if global != "" {
		files = append(files, ignoreFile{Path: global})
	}
	exclude := git("rev-parse", "--git-path", "info/exclude")
	if exclude != "" {
		if !filepath.IsAbs(exclude) {
			exclude = filepath.Join(root, exclude)
		}
		files = append(files, ignoreFile{Path: exclude})
	}
	relative, _ := filepath.Rel(base, root)
	dir := base
	if relative != "." {
		for _, component := range strings.Split(relative, string(filepath.Separator)) {
			domain := ignoreDomain(base, dir)
			for _, name := range []string{".gitignore", ".ignore", ".gittypeignore"} {
				files = append(files, ignoreFile{filepath.Join(dir, name), domain})
			}
			dir = filepath.Join(dir, component)
		}
	}
	return base, files
}
func ignoreDomain(base, path string) []string {
	rel, _ := filepath.Rel(base, path)
	if rel == "." {
		return nil
	}
	return strings.Split(filepath.ToSlash(rel), "/")
}
func readIgnore(file ignoreFile) ([]gitignore.Pattern, error) {
	b, err := os.ReadFile(file.Path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var patterns []gitignore.Pattern
	for _, line := range strings.Split(strings.TrimPrefix(string(b), "\ufeff"), "\n") {
		line = strings.TrimSuffix(line, "\r")
		if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "#") {
			patterns = append(patterns, gitignore.ParsePattern(line, file.Domain))
		}
	}
	return patterns, nil
}
