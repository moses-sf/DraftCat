/*
Package databasehandler
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package databasehandler

import "testing"

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
