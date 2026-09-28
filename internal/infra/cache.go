package infra

import (
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/michaelmjhhhh/typeit/internal/domain"
)

type CacheEntry struct {
	Key         string             `json:"key"`
	Path        string             `json:"path"`
	Fingerprint string             `json:"fingerprint"`
	Challenges  []domain.Challenge `json:"challenges"`
}
type CacheInfo struct {
	Key, Path string
	Size      int64
	Count     int
}

func fingerprint(ctx context.Context, root string) (string, error) {
	h := sha256.New()
	_, files := ignoreSources(ctx, root)
	for _, file := range files {
		b, err := os.ReadFile(file.Path)
		if err != nil && !os.IsNotExist(err) {
			return "", err
		}
		fmt.Fprintf(h, "%s\x00%s\x00", file.Path, b)
	}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if d.IsDir() {
			if path != root && (d.Name() == ".git" || d.Name() == "node_modules" || d.Name() == "target") {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		if FileLanguage(path) == "" && d.Name() != ".gitignore" && d.Name() != ".gittypeignore" && d.Name() != ".ignore" {
			return nil
		}
		info, e := d.Info()
		if e != nil {
			return e
		}
		if info.Size() > 1024*1024 {
			return nil
		}
		b, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		rel, _ := filepath.Rel(root, path)
		fmt.Fprintf(h, "%s\x00", rel)
		h.Write(b)
		h.Write([]byte{0})
		return nil
	})
	return hex.EncodeToString(h.Sum(nil)), err
}
func LoadChallenges(ctx context.Context, dir, path string, langs []string) ([]domain.Challenge, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	canonical := append([]string{}, langs...)
	sort.Strings(canonical)
	key := fmt.Sprintf("%x", sha256.Sum256([]byte("go-v2:"+abs+":"+strings.Join(canonical, ","))))
	file := filepath.Join(dir, "challenge-cache", key+".json.gz")
	fp, err := fingerprint(ctx, abs)
	if err != nil {
		return nil, err
	}
	if cached, e := readCache(file); e == nil && cached.Fingerprint == fp && len(cached.Challenges) > 0 {
		return cached.Challenges, nil
	}
	challenges, err := Scan(ctx, abs, langs, nil)
	if err != nil {
		return nil, err
	}
	if err = writeCache(file, CacheEntry{Key: key, Path: abs, Fingerprint: fp, Challenges: challenges}); err != nil {
		return nil, err
	}
	return challenges, nil
}
func readCache(path string) (CacheEntry, error) {
	var e CacheEntry
	f, err := os.Open(path)
	if err != nil {
		return e, err
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		return e, err
	}
	defer z.Close()
	err = json.NewDecoder(z).Decode(&e)
	return e, err
}
func writeCache(path string, e CacheEntry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".cache-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	z := gzip.NewWriter(f)
	if err = json.NewEncoder(z).Encode(e); err != nil {
		z.Close()
		f.Close()
		return err
	}
	if err = z.Close(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}
func CacheStats(dir string) ([]CacheInfo, error) {
	files, err := filepath.Glob(filepath.Join(dir, "challenge-cache", "*.json.gz"))
	if err != nil {
		return nil, err
	}
	out := []CacheInfo{}
	for _, path := range files {
		entry, e := readCache(path)
		if e != nil {
			return nil, e
		}
		info, e := os.Stat(path)
		if e != nil {
			return nil, e
		}
		out = append(out, CacheInfo{Key: entry.Key, Path: entry.Path, Size: info.Size(), Count: len(entry.Challenges)})
	}
	return out, nil
}
