/*
Package databasehandler
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package databasehandler

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	t.Cleanup(func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Fatalf("close test db: %v", closeErr)
		}
	})

	_, err = db.Exec(`
		CREATE TABLE chapters (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			parent_id INTEGER,
			name TEXT NOT NULL,
			path TEXT NOT NULL,
			position INTEGER NOT NULL,
			word_count INTEGER NOT NULL DEFAULT 0,
			FOREIGN KEY (parent_id) REFERENCES chapters(id)
		);

		CREATE TABLE scenes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			chapter_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			path TEXT NOT NULL,
			position INTEGER NOT NULL,
			word_count INTEGER NOT NULL DEFAULT 0,
			FOREIGN KEY (chapter_id) REFERENCES chapters(id)
		);
	`)
	if err != nil {
		t.Fatalf("create schema: %v", err)
	}

	return db
}

func assertChapterPosition(t *testing.T, db *sql.DB, name string, expectedPosition int) {
	t.Helper()

	var position int
	err := db.QueryRow(`
		SELECT position
		FROM chapters
		WHERE name = ?
	`, name).Scan(&position)
	if err != nil {
		t.Fatalf("read chapter %q position: %v", name, err)
	}

	if position != expectedPosition {
		t.Fatalf("expected chapter %q position %d, got %d", name, expectedPosition, position)
	}
}

func assertScenePosition(t *testing.T, db *sql.DB, name string, chapterID int, expectedPosition int) {
	t.Helper()

	var position int
	err := db.QueryRow(`
		SELECT position
		FROM scenes
		WHERE name = ?
		  AND chapter_id = ?
	`, name, chapterID).Scan(&position)
	if err != nil {
		t.Fatalf("read scene %q position: %v", name, err)
	}

	if position != expectedPosition {
		t.Fatalf("expected scene %q position %d, got %d", name, expectedPosition, position)
	}
}

func basicDBSetup(t *testing.T) *sql.DB {
	t.Helper()

	db := setupTestDB(t)

	parentOneID, err := AddChapter(db, ChapterCreate{
		Name:     "Chapter 1",
		Path:     "Chapter 1",
		Position: 1,
		ParentID: sql.NullInt64{Valid: false},
	})
	if err != nil {
		t.Fatal(err)
	}

	parentTwoID, err := AddChapter(db, ChapterCreate{
		Name:     "Chapter 2",
		Path:     "Chapter 2",
		Position: 2,
		ParentID: sql.NullInt64{Valid: false},
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = AddChapter(db, ChapterCreate{
		Name:     "Chapter 3",
		Path:     "Chapter 3",
		Position: 3,
		ParentID: sql.NullInt64{Valid: false},
	})
	if err != nil {
		t.Fatal(err)
	}
	parentOne := sql.NullInt64{Int64: int64(parentOneID), Valid: true}
	parentTwo := sql.NullInt64{Int64: int64(parentTwoID), Valid: true}

	_, err = AddChapter(db, ChapterCreate{
		Name:     "Notes",
		Path:     "Chapter 1/Notes",
		Position: 1,
		ParentID: parentOne,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = AddChapter(db, ChapterCreate{
		Name:     "Other",
		Path:     "Chapter 2/Other",
		Position: 1,
		ParentID: parentTwo,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = AddChapter(db, ChapterCreate{
		Name:     "Other2",
		Path:     "Chapter 2/Other2",
		Position: 2,
		ParentID: parentTwo,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = InsertChapterAtPosition(db, ChapterCreate{
		Name:     "Inserted Notes",
		Path:     "Chapter 1/Inserted Notes",
		Position: 1,
		ParentID: parentOne,
	})
	if err != nil {
		t.Fatalf("InsertChapterAtPosition returned error: %v", err)
	}
	_, err = AddChapter(db, ChapterCreate{
		Name:     "Final Notes",
		Path:     "Chapter 1/Final Notes",
		Position: 3,
		ParentID: parentOne,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = AddScene(db, SceneCreate{
		ChapterID: parentOneID,
		Name:      "scene",
		Path:      "Chapter 1/scene.md",
		Position:  1,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = InsertSceneAtPosition(db, SceneCreate{
		ChapterID: parentOneID,
		Name:      "Scene 2",
		Path:      "Chapter 1/Scene 2.md",
		Position:  2,
	})
	if err != nil {
		t.Fatalf("InsertSceneAtPosition returned error: %v", err)
	}
	_, err = InsertSceneAtPosition(db, SceneCreate{
		ChapterID: parentOneID,
		Name:      "Scene 3",
		Path:      "Chapter 1/Scene 3.md",
		Position:  3,
	})
	if err != nil {
		t.Fatalf("InsertSceneAtPosition returned error: %v", err)
	}
	return db
}
