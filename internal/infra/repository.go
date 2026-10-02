package infra

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/michaelmjhhhh/typeit/internal/domain"
)

func GitInfo(ctx context.Context, path string) domain.Repository {
	path, _ = filepath.Abs(path)
	r := domain.Repository{Owner: "local", Name: filepath.Base(path), Path: path}
	run := func(args ...string) string {
		b, err := exec.CommandContext(ctx, "git", append([]string{"-C", path}, args...)...).Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(b))
	}
	r.URL = run("remote", "get-url", "origin")
	if parsed, err := ParseRepoRef(r.URL); err == nil {
		r.Owner = parsed.Owner
		r.Name = parsed.Name
	}
	r.Branch = run("branch", "--show-current")
	r.Commit = run("rev-parse", "HEAD")
	r.Dirty = run("status", "--porcelain") != ""
	return r
}
func Clone(ctx context.Context, dir, ref string) (string, error) {
	parsed, err := ParseRepoRef(ref)
	owner, name := parsed.Owner, parsed.Name
	if err != nil {
		return "", err
	}
	origin := parsed.Origin
	if runtime.GOOS == "windows" {
		// Colons in host:port are not valid Windows directory characters.
		// ParseRepoRef rejects underscores in hosts, so this is reversible.
		origin = strings.ReplaceAll(origin, ":", "_")
	}
	parent := filepath.Join(dir, "repos", origin, owner)
	dest := filepath.Join(parent, name)
	if info, err := os.Stat(filepath.Join(dest, ".git")); err == nil && info.IsDir() {
		return dest, nil
	}
	if err = os.MkdirAll(parent, 0755); err != nil {
		return "", err
	}
	temp, err := os.MkdirTemp(parent, ".clone-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(temp)
	remote := parsed.URL
	command := exec.CommandContext(ctx, "git", "clone", "--", remote, temp)
	command.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	if b, err := command.CombinedOutput(); err != nil {
		return "", fmt.Errorf("clone %s/%s: %w: %s", owner, name, err, strings.TrimSpace(string(b)))
	}
	if err = os.Rename(temp, dest); err != nil {
		return "", err
	}
	return dest, nil
}
func CachedRepositories(dir string) ([]domain.Repository, error) {
	repos := []domain.Repository{}
	root := filepath.Join(dir, "repos")
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if os.IsNotExist(err) && path == root {
			return nil
		}
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			return nil
		}
		if strings.HasPrefix(entry.Name(), ".") {
			return filepath.SkipDir
		}
		if _, err := os.Stat(filepath.Join(path, ".git", "HEAD")); err == nil {
			rel, e := filepath.Rel(root, path)
			if e != nil {
				return e
			}
			parts := strings.Split(filepath.ToSlash(rel), "/")
			if len(parts) >= 3 {
				parts[0] = strings.ReplaceAll(parts[0], "_", ":")
				repos = append(repos, domain.Repository{Owner: parts[1], Name: strings.Join(parts[2:], "/"), Path: path, URL: "https://" + strings.Join(parts, "/")})
			}
			return filepath.SkipDir
		}
		return nil
	})
	return repos, err
}
