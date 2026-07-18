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

func TestGetMaxChapterPositionForRootChapters(t *testing.T) {
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

	_, err = AddChapter(db, ChapterCreate{
		Name:     "Chapter 2",
		Path:     "Chapter 2",
		Position: 2,
		ParentID: sql.NullInt64{Valid: false},
	})
	if err != nil {
		t.Fatal(err)
	}

	maxPosition, err := GetMaxChapterPosition(db, sql.NullInt64{Valid: false})
	if err != nil {
		t.Fatalf("GetMaxChapterPosition returned error: %v", err)
	}

	if maxPosition != 2 {
		t.Fatalf("expected max root chapter position 2, got %d", maxPosition)
	}
}

func TestGetMaxChapterPositionForChildChapters(t *testing.T) {
	db := setupTestDB(t)

	parentID, err := AddChapter(db, ChapterCreate{
		Name:     "Chapter 1",
		Path:     "Chapter 1",
		Position: 1,
		ParentID: sql.NullInt64{Valid: false},
	})
	if err != nil {
		t.Fatal(err)
	}

	parent := sql.NullInt64{
		Int64: int64(parentID),
		Valid: true,
	}

	_, err = AddChapter(db, ChapterCreate{
		Name:     "Notes",
		Path:     "Chapter 1/Notes",
		Position: 1,
		ParentID: parent,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = AddChapter(db, ChapterCreate{
		Name:     "More Notes",
		Path:     "Chapter 1/More Notes",
		Position: 2,
		ParentID: parent,
	})
	if err != nil {
		t.Fatal(err)
	}

	maxPosition, err := GetMaxChapterPosition(db, parent)
	if err != nil {
		t.Fatalf("GetMaxChapterPosition returned error: %v", err)
	}

	if maxPosition != 2 {
		t.Fatalf("expected max child chapter position 2, got %d", maxPosition)
	}
}

func TestGetMaxScenePosition(t *testing.T) {
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

	_, err = AddScene(db, SceneCreate{
		ChapterID: chapterID,
		Name:      "Scene 2",
		Path:      "Chapter 1/Scene 2.md",
		Position:  2,
	})
	if err != nil {
		t.Fatal(err)
	}

	maxPosition, err := GetMaxScenePosition(db, chapterID)
	if err != nil {
		t.Fatalf("GetMaxScenePosition returned error: %v", err)
	}

	if maxPosition != 2 {
		t.Fatalf("expected max scene position 2, got %d", maxPosition)
	}
}

func TestGetChapterNodesReturnsInsertedChapters(t *testing.T) {
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

	chapters, err := GetChapterNodes(db)
	if err != nil {
		t.Fatalf("GetChapterNodes returned error: %v", err)
	}

	if len(chapters) != 1 {
		t.Fatalf("expected 1 chapter, got %d", len(chapters))
	}

	if chapters[0].Name != "Chapter 1" {
		t.Fatalf("expected chapter name %q, got %q", "Chapter 1", chapters[0].Name)
	}
}

func TestGetSceneNodesReturnsInsertedScenes(t *testing.T) {
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

	scenes, err := GetSceneNodes(db)
	if err != nil {
		t.Fatalf("GetSceneNodes returned error: %v", err)
	}

	if len(scenes) != 1 {
		t.Fatalf("expected 1 scene, got %d", len(scenes))
	}

	if scenes[0].Name != "scene" {
		t.Fatalf("expected scene name %q, got %q", "scene", scenes[0].Name)
	}
}

func TestMapNodesBuildsChapterSceneTree(t *testing.T) {
	rootChapter := Chapter{
		ID:       1,
		Name:     "Chapter 1",
		Path:     "Chapter 1",
		Position: 1,
		ParentID: sql.NullInt64{Valid: false},
	}

	childChapter := Chapter{
		ID:       2,
		Name:     "Notes",
		Path:     "Chapter 1/Notes",
		Position: 1,
		ParentID: sql.NullInt64{Int64: 1, Valid: true},
	}

	scene := Scene{
		ID:        1,
		ChapterID: 1,
		Name:      "scene",
		Path:      "Chapter 1/scene.md",
		Position:  1,
	}

	root, err := MapNodes([]Chapter{rootChapter, childChapter}, []Scene{scene})
	if err != nil {
		t.Fatalf("MapNodes returned error: %v", err)
	}

	if len(root.Chapters) != 1 {
		t.Fatalf("expected 1 root chapter, got %d", len(root.Chapters))
	}

	chapterNode := root.Chapters[0]
	if chapterNode.Chapter.Name != "Chapter 1" {
		t.Fatalf("expected root chapter %q, got %q", "Chapter 1", chapterNode.Chapter.Name)
	}

	if chapterNode.Depth != 1 {
		t.Fatalf("expected root chapter depth 1, got %d", chapterNode.Depth)
	}

	if len(chapterNode.Chapters) != 1 {
		t.Fatalf("expected 1 child chapter, got %d", len(chapterNode.Chapters))
	}

	if chapterNode.Chapters[0].Chapter.Name != "Notes" {
		t.Fatalf("expected child chapter %q, got %q", "Notes", chapterNode.Chapters[0].Chapter.Name)
	}

	if len(chapterNode.Scenes) != 1 {
		t.Fatalf("expected 1 scene under Chapter 1, got %d", len(chapterNode.Scenes))
	}

	if chapterNode.Scenes[0].Scene.Name != "scene" {
		t.Fatalf("expected scene %q, got %q", "scene", chapterNode.Scenes[0].Scene.Name)
	}

	if chapterNode.Scenes[0].Depth != 2 {
		t.Fatalf("expected scene depth 2, got %d", chapterNode.Scenes[0].Depth)
	}
}

func TestChapterSelectWithParent(t *testing.T) {
	db := basicDBSetup(t)
	chapters, err := GetChaptersWithParent(db, sql.NullInt64{Valid: false})
	if err != nil {
		t.Error(err)
	}
	if len(chapters) != 3 {
		t.Fatalf("Incorrect number of chapters root: %d", len(chapters))
	}
	chapters, err = GetChaptersWithParent(db, sql.NullInt64{Valid: true, Int64: 1})
	if err != nil {
		t.Error(err)
	}
	if len(chapters) != 3 {
		t.Fatalf("Incorrect number of chapters 1:  %d", len(chapters))
	}
	chapters, err = GetChaptersWithParent(db, sql.NullInt64{Valid: true, Int64: 2})
	if err != nil {
		t.Error(err)
	}
	if len(chapters) != 2 {
		t.Fatalf("Incorrect number of chapters 2: %d", len(chapters))
	}
}
