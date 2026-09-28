package infra

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
)

func sqliteURI(path string) string {
	return "file:" + (&url.URL{Path: filepath.ToSlash(path)}).EscapedPath()
}

// Copy a consistent SQLite snapshot, including committed WAL data. Keep the
// original database as a backup and never replace an existing Typeit database.
func copyLegacyDatabase(source, target string) error {
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		return err
	}
	if _, err := os.Stat(source); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	stage, err := os.MkdirTemp(filepath.Dir(target), ".typeit-migration-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	db, err := sql.Open("sqlite3", sqliteURI(source)+"?mode=ro&_busy_timeout=5000")
	if err != nil {
		return err
	}
	defer db.Close()
	snapshot := filepath.Join(stage, "typeit.db")
	if _, err = db.Exec("VACUUM INTO ?", snapshot); err != nil {
		return fmt.Errorf("migrate GitType database: %w", err)
	}
	// Hard-link publication is atomic and fails if another process got here first.
	if err = os.Link(snapshot, target); os.IsExist(err) {
		return nil
	}
	return err
}

// Migrate personal history and settings on first use of the default directory.
// Downloaded repositories and disposable caches stay in the original directory.
func migrateLegacyData(source, target string) error {
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		return err
	}
	if _, err := os.Stat(source); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	stage, err := os.MkdirTemp(filepath.Dir(target), ".typeit-migration-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	if err = copyLegacyDatabase(filepath.Join(source, "gittype.db"), filepath.Join(stage, "typeit.db")); err != nil {
		return err
	}
	themes, err := filepath.Glob(filepath.Join(source, "themes", "*.json"))
	if err != nil {
		return err
	}
	files := append([]string{filepath.Join(source, "config.json"), filepath.Join(source, "custom-theme.json")}, themes...)
	for _, file := range files {
		data, err := os.ReadFile(file)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(source, file)
		if err != nil {
			return err
		}
		dest := filepath.Join(stage, rel)
		if err = os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
			return err
		}
		if err = os.WriteFile(dest, data, 0600); err != nil {
			return err
		}
	}
	if err = os.Rename(stage, target); err != nil {
		if info, statErr := os.Stat(target); statErr == nil && info.IsDir() {
			return nil // Another instance initialized the directory first.
		}
	}
	return err
}
