/*
Package databasehandler
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package databasehandler

import "testing"

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
