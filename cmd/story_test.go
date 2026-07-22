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

func TestCreateSceneCreatesEmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	scenePath := filepath.Join(tempDir, "scene.md")

	if err := CreateScene(scenePath); err != nil {
		t.Fatalf("CreateScene returned error: %v", err)
	}

	storyAssertFileExists(t, scenePath)

	content, err := os.ReadFile(scenePath)
	if err != nil {
		t.Fatalf("could not read scene file: %v", err)
	}

	if len(content) != 0 {
		t.Fatalf("expected empty scene file, got %d bytes", len(content))
	}
}

func TestCreateSceneFailsIfFileExists(t *testing.T) {
	tempDir := t.TempDir()
	scenePath := filepath.Join(tempDir, "scene.md")

	if err := os.WriteFile(scenePath, []byte("existing"), 0o644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	err := CreateScene(scenePath)
	if err == nil {
		t.Fatal("expected error when scene file already exists, got nil")
	}
}

func TestChapterExistsFindsRootChapter(t *testing.T) {
	db := setupStoryTestDB(t)

	_, err := db.Exec(`
		INSERT INTO chapters (name, path, position, parent_id)
		VALUES (?, ?, ?, NULL)
	`, "Chapter 1", "Chapter 1", 1)
	if err != nil {
		t.Fatalf("setup insert failed: %v", err)
	}

	exists, err := ChapterExists(db, "Chapter 1", sql.NullInt64{Valid: false})
	if err != nil {
		t.Fatalf("ChapterExists returned error: %v", err)
	}

	if !exists {
		t.Fatal("expected root chapter to exist")
	}
}

func TestChapterExistsFindsChildChapter(t *testing.T) {
	db := setupStoryTestDB(t)

	result, err := db.Exec(`
		INSERT INTO chapters (name, path, position, parent_id)
		VALUES (?, ?, ?, NULL)
	`, "Chapter 1", "Chapter 1", 1)
	if err != nil {
		t.Fatalf("setup parent insert failed: %v", err)
	}

	parentID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("could not get parent id: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO chapters (name, path, position, parent_id)
		VALUES (?, ?, ?, ?)
	`, "Notes", "Chapter 1/Notes", 1, parentID)
	if err != nil {
		t.Fatalf("setup child insert failed: %v", err)
	}

	exists, err := ChapterExists(db, "Notes", sql.NullInt64{
		Int64: parentID,
		Valid: true,
	})
	if err != nil {
		t.Fatalf("ChapterExists returned error: %v", err)
	}

	if !exists {
		t.Fatal("expected child chapter to exist")
	}
}

func TestChapterExistsReturnsFalseForMissingChapter(t *testing.T) {
	db := setupStoryTestDB(t)

	exists, err := ChapterExists(db, "Missing Chapter", sql.NullInt64{Valid: false})
	if err != nil {
		t.Fatalf("ChapterExists returned error: %v", err)
	}

	if exists {
		t.Fatal("expected missing chapter to return false")
	}
}

func TestCreateChapterCreatesFolderSceneTomlAndDatabaseRows(t *testing.T) {
	db := setupStoryTestDB(t)
	tempDir := t.TempDir()

	chapterPath := filepath.Join(tempDir, "Chapter 1")
	folderOpts := AddFolderOptions{
		Name:        "Chapter 1",
		Position:    0,
		ParentID:    sql.NullInt64{Valid: false},
		MaxPosition: 0,
	}
	err := CreateChapter(
		db,
		chapterPath,
		"..",
		folderOpts,
	)
	if err != nil {
		t.Fatalf("CreateChapter returned error: %v", err)
	}

	storyAssertDirExists(t, chapterPath)
	storyAssertFileExists(t, filepath.Join(chapterPath, "scene.md"))
	storyAssertFileExists(t, filepath.Join(chapterPath, ".chapter.toml"))

	var chapterCount int
	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM chapters
		WHERE name = ?
		  AND parent_id IS NULL
	`, "Chapter 1").Scan(&chapterCount)
	if err != nil {
		t.Fatalf("could not count chapters: %v", err)
	}

	if chapterCount != 1 {
		t.Fatalf("expected 1 chapter row, got %d", chapterCount)
	}

	var sceneCount int
	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM scenes
		WHERE name = ?
	`, "scene").Scan(&sceneCount)
	if err != nil {
		t.Fatalf("could not count scenes: %v", err)
	}

	if sceneCount != 1 {
		t.Fatalf("expected 1 scene row, got %d", sceneCount)
	}
}

func TestCreateChapterAtPositionShiftsExistingRootChapters(t *testing.T) {
	db := setupStoryTestDB(t)
	tempDir := t.TempDir()

	firstPath := filepath.Join(tempDir, "Chapter 1")
	secondPath := filepath.Join(tempDir, "Chapter 2")
	folderOptions := AddFolderOptions{
		Name:        "Chapter 1",
		Position:    0,
		ParentID:    sql.NullInt64{Valid: false},
		MaxPosition: 0,
	}
	if err := CreateChapter(db, firstPath, "..", folderOptions); err != nil {
		t.Fatalf("CreateChapter first returned error: %v", err)
	}
	folderOptions = AddFolderOptions{
		Name:        "Chapter 2",
		Position:    1,
		ParentID:    sql.NullInt64{Valid: false},
		MaxPosition: 1,
	}

	if err := CreateChapter(db, secondPath, "..", folderOptions); err != nil {
		t.Fatalf("CreateChapter second returned error: %v", err)
	}

	var chapterOnePosition int
	err := db.QueryRow(`
		SELECT position
		FROM chapters
		WHERE name = ?
	`, "Chapter 1").Scan(&chapterOnePosition)
	if err != nil {
		t.Fatalf("could not read Chapter 1 position: %v", err)
	}

	var chapterTwoPosition int
	err = db.QueryRow(`
		SELECT position
		FROM chapters
		WHERE name = ?
	`, "Chapter 2").Scan(&chapterTwoPosition)
	if err != nil {
		t.Fatalf("could not read Chapter 2 position: %v", err)
	}

	if chapterTwoPosition != 1 {
		t.Fatalf("expected Chapter 2 position 1, got %d", chapterTwoPosition)
	}

	if chapterOnePosition != 2 {
		t.Fatalf("expected Chapter 1 position 2 after shift, got %d", chapterOnePosition)
	}
}

func setupStoryTestDB(t *testing.T) *sql.DB {
	t.Helper()

	tempDir := t.TempDir()

	if err := InitialiseDB(tempDir); err != nil {
		t.Fatalf("InitialiseDB returned error: %v", err)
	}

	db, err := sql.Open("sqlite", filepath.Join(tempDir, ".story.db"))
	if err != nil {
		t.Fatalf("could not open test db: %v", err)
	}

	t.Cleanup(func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Fatal(closeErr)
		}
	})

	return db
}

func storyAssertDirExists(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected directory %q to exist: %v", path, err)
	}

	if !info.IsDir() {
		t.Fatalf("expected %q to be a directory", path)
	}
}

func storyAssertFileExists(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected file %q to exist: %v", path, err)
	}

	if info.IsDir() {
		t.Fatalf("expected %q to be a file, got directory", path)
	}
}
