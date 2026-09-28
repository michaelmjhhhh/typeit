package gittype_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/michaelmjhhhh/typeit/internal/infra"
)

func migrationHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("TYPEIT_DATA_DIR", "")
	t.Setenv("GITTYPE_DATA_DIR", "")
	return home
}

func legacyDatabase(t *testing.T, dir string) *sql.DB {
	t.Helper()
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite3", filepath.Join(dir, "gittype.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if _, err = db.Exec("PRAGMA journal_mode=WAL; PRAGMA wal_autocheckpoint=0; CREATE TABLE migration_fixture(value TEXT); INSERT INTO migration_fixture VALUES('saved history');"); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestLegacyDataMigration(t *testing.T) {
	home := migrationHome(t)
	legacy := filepath.Join(home, ".gittype")
	source := legacyDatabase(t, legacy)
	// Keep the source open so committed records still live in its WAL.
	for _, name := range []string{"config.json", "custom-theme.json", "themes/personal.json"} {
		path := filepath.Join(legacy, name)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("{}\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	dir, err := infra.DataDir()
	if err != nil || dir != filepath.Join(home, ".typeit") {
		t.Fatalf("directory %q: %v", dir, err)
	}
	for _, name := range []string{"config.json", "custom-theme.json", "themes/personal.json"} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil || string(data) != "{}\n" {
			t.Fatalf("settings %s: %q %v", name, data, err)
		}
	}
	db, err := infra.OpenDatabase(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var value string
	if err = db.DB.QueryRow("SELECT value FROM migration_fixture").Scan(&value); err != nil || value != "saved history" {
		t.Fatalf("WAL history not migrated: %q %v", value, err)
	}
	if _, err = source.Exec("INSERT INTO migration_fixture VALUES('later legacy session')"); err != nil {
		t.Fatal("original database must remain usable", err)
	}
	if _, err = infra.DataDir(); err != nil {
		t.Fatal(err)
	}
	var count int
	if err = db.DB.QueryRow("SELECT COUNT(*) FROM migration_fixture").Scan(&count); err != nil || count != 1 {
		t.Fatalf("migration ran again: %d %v", count, err)
	}
}

func TestDataDirectoryOverrides(t *testing.T) {
	home := migrationHome(t)
	old := filepath.Join(home, "legacy-override")
	current := filepath.Join(home, "current-override")
	t.Setenv("GITTYPE_DATA_DIR", old)
	if dir, err := infra.DataDir(); err != nil || dir != old {
		t.Fatalf("legacy override: %q %v", dir, err)
	}
	t.Setenv("TYPEIT_DATA_DIR", current)
	if dir, err := infra.DataDir(); err != nil || dir != current {
		t.Fatalf("Typeit override must win: %q %v", dir, err)
	}
}

func TestCustomDirectoryDatabaseMigration(t *testing.T) {
	dir := t.TempDir()
	legacyDatabase(t, dir)
	for i := 0; i < 2; i++ {
		db, err := infra.OpenDatabase(dir)
		if err != nil {
			t.Fatal(err)
		}
		var value string
		err = db.DB.QueryRow("SELECT value FROM migration_fixture").Scan(&value)
		db.Close()
		if err != nil || value != "saved history" {
			t.Fatalf("history missing: %q %v", value, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "gittype.db")); err != nil {
		t.Fatal("original database removed", err)
	}
}

func TestMigrationFailureDoesNotCreateEmptyHistory(t *testing.T) {
	home := migrationHome(t)
	old := filepath.Join(home, ".gittype")
	if err := os.Mkdir(old, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(old, "gittype.db"), []byte("corrupt database"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := infra.DataDir(); err == nil {
		t.Fatal("corrupt database must fail migration")
	}
	if _, err := os.Stat(filepath.Join(home, ".typeit")); !os.IsNotExist(err) {
		t.Fatalf("partial migration published: %v", err)
	}
}
