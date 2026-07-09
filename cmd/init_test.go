/*
Package cmd
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package cmd

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestCreateDirectoryCreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "Monkey")

	if err := CreateDirectory(target); err != nil {
		t.Fatalf("CreateDirectory returned error: %v", err)
	}

	assertDirExists(t, target)
}

func TestCreateDirectoryFailsIfDirectoryExists(t *testing.T) {
	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "Monkey")

	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	err := CreateDirectory(target)
	if err == nil {
		t.Fatal("expected error when directory already exists, got nil")
	}
}

func TestCreateStoryDirectoryCreatesDirectoryInCurrentWorkingDirectory(t *testing.T) {
	tempDir := t.TempDir()

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatal(err)
		}
	})

	path, err := CreateStoryDirectory("Monkey")
	if err != nil {
		t.Fatalf("CreateStoryDirectory returned error: %v", err)
	}

	expected := filepath.Join(tempDir, "Monkey")

	if canonicalPath(t, path) != canonicalPath(t, expected) {
		t.Fatalf("expected path %q, got %q", expected, path)
	}

	assertDirExists(t, expected)
}

func TestGetNameUsesProvidedName(t *testing.T) {
	name, err := GetName(InitOptions{
		Name: "Monkey",
	})
	if err != nil {
		t.Fatalf("GetName returned error: %v", err)
	}

	if name != "Monkey" {
		t.Fatalf("expected name %q, got %q", "Monkey", name)
	}
}

func TestInitialiseDBCreatesStoryDB(t *testing.T) {
	tempDir := t.TempDir()

	if err := InitialiseDB(tempDir); err != nil {
		t.Fatalf("InitialiseDB returned error: %v", err)
	}

	dbPath := filepath.Join(tempDir, ".story.db")
	assertFileExists(t, dbPath)
}

func TestInitialiseDBCreatesTables(t *testing.T) {
	tempDir := t.TempDir()

	if err := InitialiseDB(tempDir); err != nil {
		t.Fatalf("InitialiseDB returned error: %v", err)
	}

	db, err := sql.Open("sqlite", filepath.Join(tempDir, ".story.db"))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Fatal(closeErr)
		}
	})

	assertTableExists(t, db, "chapters")
	assertTableExists(t, db, "scenes")
}

func TestCreateShortStoryCreatesProject(t *testing.T) {
	tempDir := t.TempDir()

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatal(err)
		}
	})

	path, err := CreateShortStory(InitOptions{
		Name:     "Monkey",
		WorkType: "short",
		Defaults: true,
		Vim:      false,
	})
	if err != nil {
		t.Fatalf("CreateShortStory returned error: %v", err)
	}

	expectedRoot := filepath.Join(tempDir, "Monkey")

	if canonicalPath(t, path) != canonicalPath(t, expectedRoot) {
		t.Fatalf("expected path %q, got %q", expectedRoot, path)
	}

	assertDirExists(t, expectedRoot)
	assertFileExists(t, filepath.Join(expectedRoot, ".story.toml"))
	assertFileExists(t, filepath.Join(expectedRoot, ".story.db"))
}

func TestCreateShortStoryInitialisesDatabase(t *testing.T) {
	tempDir := t.TempDir()

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatal(err)
		}
	})

	path, err := CreateShortStory(InitOptions{
		Name:     "Monkey",
		WorkType: "short",
		Defaults: true,
		Vim:      false,
	})
	if err != nil {
		t.Fatalf("CreateShortStory returned error: %v", err)
	}

	db, err := sql.Open("sqlite", filepath.Join(path, ".story.db"))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Fatal(closeErr)
		}
	})

	assertTableExists(t, db, "chapters")
	assertTableExists(t, db, "scenes")
}

func TestInitWorkRejectsInvalidWorkType(t *testing.T) {
	_, err := InitWork(InitOptions{
		Name:     "Monkey",
		WorkType: "epic-poem",
		Defaults: true,
		Vim:      false,
	})

	if err == nil {
		t.Fatal("expected error for invalid work type, got nil")
	}
}

func canonicalPath(t *testing.T, path string) string {
	t.Helper()

	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatalf("could not resolve path %q: %v", path, err)
	}

	return resolved
}

func assertDirExists(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected directory %q to exist: %v", path, err)
	}

	if !info.IsDir() {
		t.Fatalf("expected %q to be a directory", path)
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected file %q to exist: %v", path, err)
	}

	if info.IsDir() {
		t.Fatalf("expected %q to be a file, got directory", path)
	}
}

func assertTableExists(t *testing.T, db *sql.DB, tableName string) {
	t.Helper()

	var name string
	err := db.QueryRow(`
		SELECT name
		FROM sqlite_master
		WHERE type = 'table'
		  AND name = ?
	`, tableName).Scan(&name)
	if err != nil {
		t.Fatalf("expected table %q to exist: %v", tableName, err)
	}

	if name != tableName {
		t.Fatalf("expected table name %q, got %q", tableName, name)
	}
}
