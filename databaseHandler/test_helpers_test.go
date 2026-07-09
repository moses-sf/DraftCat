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
