/*
Package utilities
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package utilities

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestFolderTomlExistsReturnsTrueForFile(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, ".chapter.toml")

	if err := os.WriteFile(path, []byte("ID = 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if !FolderTomlExists(path) {
		t.Fatal("expected FolderTomlExists to return true for existing file")
	}
}

func TestFolderTomlExistsReturnsFalseForDirectory(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, ".chapter.toml")

	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}

	if FolderTomlExists(path) {
		t.Fatal("expected FolderTomlExists to return false for directory")
	}
}

func TestFindTomlPathFindsTomlInGivenDirectory(t *testing.T) {
	tempDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tempDir, ".story.toml"), []byte("test = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	path, err := FindTomlPath(tempDir, ".story.toml")
	if err != nil {
		t.Fatalf("FindTomlPath returned error: %v", err)
	}

	if canonicalUtilityPath(t, path) != canonicalUtilityPath(t, tempDir) {
		t.Fatalf("expected %q, got %q", tempDir, path)
	}
}

func TestFindTomlPathReturnsErrorWhenMissing(t *testing.T) {
	tempDir := t.TempDir()

	_, err := FindTomlPath(tempDir, ".story.toml")
	if err == nil {
		t.Fatal("expected error for missing toml, got nil")
	}
}

func TestLoadChapterToml(t *testing.T) {
	tempDir := t.TempDir()

	content := []byte(`
ID = 12
PathToRoot = "../.."
Position = 3

[ParentID]
Int64 = 1
Valid = true

[[Scenes]]
ID = 4
Name = "Opening"
Path = "opening.md"
Position = 1
`)

	if err := os.WriteFile(filepath.Join(tempDir, ".chapter.toml"), content, 0o644); err != nil {
		t.Fatal(err)
	}

	chapter, err := LoadChapterToml(tempDir)
	if err != nil {
		t.Fatalf("LoadChapterToml returned error: %v", err)
	}

	if chapter.ID != 12 {
		t.Fatalf("expected ID 12, got %d", chapter.ID)
	}

	if !chapter.ParentID.Valid || chapter.ParentID.Int64 != 1 {
		t.Fatalf("expected ParentID 1 valid, got %+v", chapter.ParentID)
	}

	if chapter.PathToRoot != "../.." {
		t.Fatalf("expected PathToRoot ../.., got %q", chapter.PathToRoot)
	}

	if chapter.Position != 3 {
		t.Fatalf("expected Position 3, got %d", chapter.Position)
	}

	if len(chapter.Scenes) != 1 {
		t.Fatalf("expected 1 scene, got %d", len(chapter.Scenes))
	}

	if chapter.Scenes[0].Name != "Opening" {
		t.Fatalf("expected scene Opening, got %q", chapter.Scenes[0].Name)
	}
}

func TestIsDraftcatProjectReturnsTrueForStoryRoot(t *testing.T) {
	tempDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tempDir, ".story.toml"), []byte("test = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	withWorkingDirectory(t, tempDir, func() {
		ok, err := IsDraftcatProject()
		if err != nil {
			t.Fatalf("IsDraftcatProject returned error: %v", err)
		}

		if !ok {
			t.Fatal("expected IsDraftcatProject to return true")
		}
	})
}

func TestIsDraftcatProjectReturnsTrueForChapterFolder(t *testing.T) {
	tempDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tempDir, ".chapter.toml"), []byte("ID = 1\nPathToRoot = \"..\"\nPosition = 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	withWorkingDirectory(t, tempDir, func() {
		ok, err := IsDraftcatProject()
		if err != nil {
			t.Fatalf("IsDraftcatProject returned error: %v", err)
		}

		if !ok {
			t.Fatal("expected IsDraftcatProject to return true")
		}
	})
}

func TestIsDraftcatProjectReturnsErrorOutsideProject(t *testing.T) {
	tempDir := t.TempDir()

	withWorkingDirectory(t, tempDir, func() {
		ok, err := IsDraftcatProject()
		if err == nil {
			t.Fatal("expected error outside Draftcat project")
		}

		if ok {
			t.Fatal("expected ok false outside Draftcat project")
		}
	})
}

func TestGetRelativeRootPathFromStoryRoot(t *testing.T) {
	tempDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tempDir, ".story.toml"), []byte("test = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	withWorkingDirectory(t, tempDir, func() {
		root, err := GetRelativeRootPath()
		if err != nil {
			t.Fatalf("GetRelativeRootPath returned error: %v", err)
		}

		if canonicalUtilityPath(t, root) != canonicalUtilityPath(t, tempDir) {
			t.Fatalf("expected root %q, got %q", tempDir, root)
		}
	})
}

func TestGetRelativeRootPathFromChapterFolder(t *testing.T) {
	rootDir := t.TempDir()
	chapterDir := filepath.Join(rootDir, "Chapter 1")

	if err := os.MkdirAll(chapterDir, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(rootDir, ".story.toml"), []byte("test = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	chapter := ChapterMetaData{
		ID:         1,
		ParentID:   sql.NullInt64{Valid: false},
		PathToRoot: "..",
		Position:   1,
	}

	if err := writeChapterTomlForTest(filepath.Join(chapterDir, ".chapter.toml"), chapter); err != nil {
		t.Fatal(err)
	}

	withWorkingDirectory(t, chapterDir, func() {
		root, err := GetRelativeRootPath()
		if err != nil {
			t.Fatalf("GetRelativeRootPath returned error: %v", err)
		}

		if canonicalUtilityPath(t, root) != canonicalUtilityPath(t, rootDir) {
			t.Fatalf("expected root %q, got %q", rootDir, root)
		}
	})
}

func withWorkingDirectory(t *testing.T, dir string, fn func()) {
	t.Helper()

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatal(err)
		}
	}()

	fn()
}

func canonicalUtilityPath(t *testing.T, path string) string {
	t.Helper()

	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatalf("could not resolve path %q: %v", path, err)
	}

	return resolved
}
