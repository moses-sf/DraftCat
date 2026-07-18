/*
Package databasehandler
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package databasehandler

import (
	"database/sql"
	"testing"
)

func TestDeleteSceneInitial(t *testing.T) {
	db := basicDBSetup(t)
	scene, err := GetScene(db, 1)
	if err != nil {
		t.Fatal(err)
	}
	err = DeleteSceneUpdatePosition(db, 1, scene.ChapterID, scene.Position)
	if err != nil {
		t.Fatal(err)
	}
	sceneTwo, err := GetScene(db, 2)
	if err != nil {
		t.Fatal(err)
	}
	if sceneTwo.Position != 1 {
		t.Fatalf("incorrect position %d", sceneTwo.Position)
	}
	sceneThree, err := GetScene(db, 3)
	if err != nil {
		t.Fatal(err)
	}
	if sceneThree.Position != 2 {
		t.Fatalf("incorrect position %d", sceneThree.Position)
	}
}

func TestDeleteSceneMiddle(t *testing.T) {
	db := basicDBSetup(t)
	scene, err := GetScene(db, 2)
	if err != nil {
		t.Fatal(err)
	}
	err = DeleteSceneUpdatePosition(db, 2, scene.ChapterID, scene.Position)
	if err != nil {
		t.Fatal(err)
	}
	sceneOne, err := GetScene(db, 1)
	if err != nil {
		t.Fatal(err)
	}
	if sceneOne.Position != 1 {
		t.Fatalf("incorrect position %d", sceneOne.Position)
	}
	sceneThree, err := GetScene(db, 3)
	if err != nil {
		t.Fatal(err)
	}
	if sceneThree.Position != 2 {
		t.Fatalf("incorrect position %d", sceneThree.Position)
	}
}

func TestDeleteLeafChapter(t *testing.T) {
	db := basicDBSetup(t)
	chapter, err := GetChapter(db, 5)
	if err != nil {
		t.Fatal(err)
	}
	chapterTwo, err := GetChapter(db, 6)
	if err != nil {
		t.Fatal(err)
	}
	if chapterTwo.Position != 2 {
		t.Fatalf("incorrect position 6: %d", chapterTwo.Position)
	}
	err = DeleteChapterUpdatePosition(db, chapter.ID, chapter.Position, chapter.ParentID)
	if err != nil {
		t.Fatal(err)
	}
	chapterTwo, err = GetChapter(db, 6)
	if err != nil {
		t.Fatal(err)
	}
	if chapterTwo.Position != 1 {
		t.Fatalf("incorrect position 6: %d", chapterTwo.Position)
	}
}

func TestDeleteAllScenesOfChapter(t *testing.T) {
	db := setupTestDB(t)
	err := DeleteScenesFromChapter(db, 1)
	if err != nil {
		t.Fatal(err)
	}
	_, err = GetScene(db, 1)
	if err != sql.ErrNoRows {
		t.Fatal(err)
	}
}
