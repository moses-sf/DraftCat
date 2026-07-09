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

func TestAddChapterPersistsRootChapter(t *testing.T) {
	db := setupTestDB(t)

	id, err := AddChapter(db, ChapterCreate{
		Name:     "Chapter 1",
		Path:     "Chapter 1",
		Position: 1,
		ParentID: sql.NullInt64{Valid: false},
	})
	if err != nil {
		t.Fatalf("AddChapter returned error: %v", err)
	}

	if id == 0 {
		t.Fatal("expected non-zero chapter id")
	}

	var count int
	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM chapters
		WHERE id = ?
		  AND name = ?
		  AND parent_id IS NULL
	`, id, "Chapter 1").Scan(&count)
	if err != nil {
		t.Fatalf("count chapter: %v", err)
	}

	if count != 1 {
		t.Fatalf("expected 1 persisted chapter, got %d", count)
	}
}

func TestInsertChapterAtPositionShiftsRootSiblings(t *testing.T) {
	db := setupTestDB(t)

	_, err := AddChapter(db, ChapterCreate{
		Name:     "Chapter 1",
		Path:     "Chapter 1",
		Position: 1,
		ParentID: sql.NullInt64{Valid: false},
	})
	if err != nil {
		t.Fatal(err)
	}

	insertedID, err := InsertChapterAtPosition(db, ChapterCreate{
		Name:     "Chapter 2",
		Path:     "Chapter 2",
		Position: 1,
		ParentID: sql.NullInt64{Valid: false},
	})
	if err != nil {
		t.Fatalf("InsertChapterAtPosition returned error: %v", err)
	}

	if insertedID == 0 {
		t.Fatal("expected non-zero inserted id")
	}

	assertChapterPosition(t, db, "Chapter 2", 1)
	assertChapterPosition(t, db, "Chapter 1", 2)
}

func TestInsertChapterAtPositionShiftsChildSiblingsOnly(t *testing.T) {
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

	_, err = InsertChapterAtPosition(db, ChapterCreate{
		Name:     "Inserted Notes",
		Path:     "Chapter 1/Inserted Notes",
		Position: 1,
		ParentID: parentOne,
	})
	if err != nil {
		t.Fatalf("InsertChapterAtPosition returned error: %v", err)
	}

	assertChapterPosition(t, db, "Inserted Notes", 1)
	assertChapterPosition(t, db, "Notes", 2)
	assertChapterPosition(t, db, "Other", 1)
}

func TestAddScenePersistsScene(t *testing.T) {
	db := setupTestDB(t)

	chapterID, err := AddChapter(db, ChapterCreate{
		Name:     "Chapter 1",
		Path:     "Chapter 1",
		Position: 1,
		ParentID: sql.NullInt64{Valid: false},
	})
	if err != nil {
		t.Fatal(err)
	}

	sceneID, err := AddScene(db, SceneCreate{
		ChapterID: chapterID,
		Name:      "scene",
		Path:      "Chapter 1/scene.md",
		Position:  1,
	})
	if err != nil {
		t.Fatalf("AddScene returned error: %v", err)
	}

	if sceneID == 0 {
		t.Fatal("expected non-zero scene id")
	}

	var count int
	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM scenes
		WHERE id = ?
		  AND chapter_id = ?
		  AND name = ?
	`, sceneID, chapterID, "scene").Scan(&count)
	if err != nil {
		t.Fatalf("count scene: %v", err)
	}

	if count != 1 {
		t.Fatalf("expected 1 persisted scene, got %d", count)
	}
}

func TestInsertSceneAtPositionShiftsScenesInSameChapter(t *testing.T) {
	db := setupTestDB(t)

	chapterID, err := AddChapter(db, ChapterCreate{
		Name:     "Chapter 1",
		Path:     "Chapter 1",
		Position: 1,
		ParentID: sql.NullInt64{Valid: false},
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = AddScene(db, SceneCreate{
		ChapterID: chapterID,
		Name:      "scene",
		Path:      "Chapter 1/scene.md",
		Position:  1,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = InsertSceneAtPosition(db, SceneCreate{
		ChapterID: chapterID,
		Name:      "Scene 2",
		Path:      "Chapter 1/Scene 2.md",
		Position:  1,
	})
	if err != nil {
		t.Fatalf("InsertSceneAtPosition returned error: %v", err)
	}

	assertScenePosition(t, db, "Scene 2", chapterID, 1)
	assertScenePosition(t, db, "scene", chapterID, 2)
}
