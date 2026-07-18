/*
Package databasehandler
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package databasehandler

import (
	"database/sql"
	"testing"
)

func TestChapterPathAndNameChange(t *testing.T) {
	db := basicDBSetup(t)
	err := UpdateChapterPathAndName(db, &Chapter{
		ID:   1,
		Name: "TestUpdate",
		Path: "TestUpdate/",
	})
	if err != nil {
		t.Fatal("Failed to update db")
	}

	chapter, err := GetChapter(db, 1)
	if err != nil {
		t.Fatal("Couldn't retrieve chapter")
	}

	if chapter.Name != "TestUpdate" {
		t.Fatal("Failed to update Name")
	}

	if chapter.Path != "TestUpdate/" {
		t.Fatal("Failed to update Path")
	}
}

func TestChapterPathChange(t *testing.T) {
	db := basicDBSetup(t)
	err := UpdateChapterPath(db, &Chapter{
		ID:   1,
		Name: "TestUpdate",
		Path: "TestUpdate/",
	})
	if err != nil {
		t.Fatal("Failed to update db")
	}

	chapter, err := GetChapter(db, 1)
	if err != nil {
		t.Fatal("Couldn't retrieve chapter")
	}

	if chapter.Path != "TestUpdate/" {
		t.Fatal("Failed to update Path")
	}
}

func TestScenePathChange(t *testing.T) {
	db := basicDBSetup(t)
	err := UpdateScenePath(db, &Scene{
		ID:   1,
		Path: "TestUpdate/scene.md",
	})
	if err != nil {
		t.Fatal("Failed to update db")
	}

	scene, err := GetScene(db, 1)
	if err != nil {
		t.Fatal("Couldn't retrieve scene")
	}

	if scene.Path != "TestUpdate/scene.md" {
		t.Fatal("Failed to update Path")
	}
}

func TestSceneRename(t *testing.T) {
	db := basicDBSetup(t)
	err := UpdateScenePathAndName(db, &Scene{
		ID:   1,
		Name: "TestUpdate",
		Path: "TestUpdate/TestUpdate.md",
	})
	if err != nil {
		t.Fatal("Failed to update db")
	}

	scene, err := GetScene(db, 1)
	if err != nil {
		t.Fatal("Couldn't retrieve chapter")
	}

	if scene.Name != "TestUpdate" {
		t.Fatal("Failed to update Name")
	}

	if scene.Path != "TestUpdate/TestUpdate.md" {
		t.Fatal("Failed to update Path")
	}
}

func TestUpdateScenePositionIncrease(t *testing.T) {
	db := basicDBSetup(t)
	err := UpdateScenePosition(db, 1, 1, 3, 1)
	if err != nil {
		t.Fatal("Failed to update DB")
	}
	sceneOne, err := GetScene(db, 1)
	if err != nil {
		t.Fatal("Couldn't retrieve chapter")
	}
	if sceneOne.Position != 3 {
		t.Fatalf("New position is %d", sceneOne.Position)
	}
	sceneTwo, err := GetScene(db, 2)
	if err != nil {
		t.Fatal("Couldn't retrieve chapter")
	}
	if sceneTwo.Position != 1 {
		t.Fatalf("New position is %d", sceneOne.Position)
	}
	sceneThree, err := GetScene(db, 3)
	if err != nil {
		t.Fatal("Couldn't retrieve chapter")
	}
	if sceneThree.Position != 2 {
		t.Fatalf("New position is %d", sceneOne.Position)
	}
}

func TestUpdateScenePositionDecrease(t *testing.T) {
	db := basicDBSetup(t)
	err := UpdateScenePosition(db, 3, 3, 1, 1)
	if err != nil {
		t.Fatal("Failed to update DB")
	}
	sceneOne, err := GetScene(db, 1)
	if err != nil {
		t.Fatal("Couldn't retrieve scene")
	}
	if sceneOne.Position != 2 {
		t.Fatalf("New position is %d", sceneOne.Position)
	}
	sceneTwo, err := GetScene(db, 2)
	if err != nil {
		t.Fatal("Couldn't retrieve scene")
	}
	if sceneTwo.Position != 3 {
		t.Fatalf("New position is %d", sceneTwo.Position)
	}
	sceneThree, err := GetScene(db, 3)
	if err != nil {
		t.Fatal("Couldn't retrieve scene")
	}
	if sceneThree.Position != 1 {
		t.Fatalf("New position is %d", sceneThree.Position)
	}
}

func TestUpdateChapterPositionIncreaseNull(t *testing.T) {
	db := basicDBSetup(t)
	err := UpdateChapterPosition(db, 1, 1, 3, sql.NullInt64{Valid: false})
	if err != nil {
		t.Fatal("Failed to update DB")
	}
	chapterOne, err := GetChapter(db, 1)
	if err != nil {
		t.Fatal("Couldn't retrieve chapter")
	}
	if chapterOne.Position != 3 {
		t.Fatalf("New position is %d", chapterOne.Position)
	}
	chapterTwo, err := GetChapter(db, 2)
	if err != nil {
		t.Fatal("Couldn't retrieve chapter")
	}
	if chapterTwo.Position != 1 {
		t.Fatalf("New position is %d", chapterTwo.Position)
	}
	chapterThree, err := GetChapter(db, 3)
	if err != nil {
		t.Fatal("Couldn't retrieve chapter")
	}
	if chapterThree.Position != 2 {
		t.Fatalf("New position is %d", chapterThree.Position)
	}
}

func TestUpdateChapterPositionDecreaseNull(t *testing.T) {
	db := basicDBSetup(t)
	err := UpdateChapterPosition(db, 3, 3, 1, sql.NullInt64{Valid: false})
	if err != nil {
		t.Fatal("Failed to update DB")
	}
	chapterOne, err := GetChapter(db, 1)
	if err != nil {
		t.Fatal("Couldn't retrieve chapter")
	}
	if chapterOne.Position != 2 {
		t.Fatalf("New position is %d", chapterOne.Position)
	}
	chapterTwo, err := GetChapter(db, 2)
	if err != nil {
		t.Fatal("Couldn't retrieve chapter")
	}
	if chapterTwo.Position != 3 {
		t.Fatalf("New position is %d", chapterTwo.Position)
	}
	chapterThree, err := GetChapter(db, 3)
	if err != nil {
		t.Fatal("Couldn't retrieve chapter")
	}
	if chapterThree.Position != 1 {
		t.Fatalf("New position is %d", chapterThree.Position)
	}
}

func TestUpdateChapterPositionIncrease(t *testing.T) {
	db := basicDBSetup(t)
	err := UpdateChapterPosition(db, 7, 1, 3, sql.NullInt64{Valid: true, Int64: 1})
	if err != nil {
		t.Fatal("Failed to update DB")
	}
	chapterOne, err := GetChapter(db, 7)
	if err != nil {
		t.Fatal("Couldn't retrieve chapter")
	}
	if chapterOne.Position != 3 {
		t.Fatalf("New position is %d", chapterOne.Position)
	}
	chapterTwo, err := GetChapter(db, 4)
	if err != nil {
		t.Fatal("Couldn't retrieve chapter")
	}
	if chapterTwo.Position != 1 {
		t.Fatalf("New position is %d", chapterTwo.Position)
	}
	chapterThree, err := GetChapter(db, 8)
	if err != nil {
		t.Fatal("Couldn't retrieve chapter")
	}
	if chapterThree.Position != 2 {
		t.Fatalf("New position 7 is %d", chapterThree.Position)
	}
}

func TestUpdateChapterPositionDecrease(t *testing.T) {
	db := basicDBSetup(t)
	err := UpdateChapterPosition(db, 7, 3, 1, sql.NullInt64{Valid: true, Int64: 1})
	if err != nil {
		t.Fatal("Failed to update DB")
	}
	chapterOne, err := GetChapter(db, 6)
	if err != nil {
		t.Fatal("Couldn't retrieve chapter")
	}
	if chapterOne.Position != 2 {
		t.Fatalf("New position is %d", chapterOne.Position)
	}
	chapterTwo, err := GetChapter(db, 4)
	if err != nil {
		t.Fatal("Couldn't retrieve chapter")
	}
	if chapterTwo.Position != 3 {
		t.Fatalf("New position is %d", chapterTwo.Position)
	}
	chapterThree, err := GetChapter(db, 7)
	if err != nil {
		t.Fatal("Couldn't retrieve chapter")
	}
	if chapterThree.Position != 1 {
		t.Fatalf("New position is %d", chapterThree.Position)
	}
}

func TestSceneToggleCompile(t *testing.T) {
	db := basicDBSetup(t)
	scene, err := GetScene(db, 1)
	if err != nil {
		t.Fatal(err)
	}
	if scene.Compile {
		t.Fatal("Incorrect startup")
	}
	err = UpdateSceneCompile(db, 1, true)
	if err != nil {
		t.Fatal(err)
	}
	scene, err = GetScene(db, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !scene.Compile {
		t.Fatal("Scene Didn't toggle")
	}
}

func TestSceneChapterWiseToggleCompile(t *testing.T) {
	db := basicDBSetup(t)
	err := UpdateSceneCompileChapter(db, 1, true)
	if err != nil {
		t.Fatal(err)
	}
	scenes, err := GetScenesOfChapter(db, 1)
	if err != nil {
		t.Fatal(err)
	}
	for _, scene := range scenes {
		if !scene.Compile {
			t.Fatalf("Scene didn't toggle ID: %d", scene.ID)
		}
	}
}
