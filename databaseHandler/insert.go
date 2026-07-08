/*
Package databasehandler
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package databasehandler

import (
	"database/sql"
	"fmt"
)

type ChapterCreate struct {
	Name     string
	ParentID sql.NullInt64
	Path     string
	Position int
}

type SceneCreate struct {
	ChapterID int
	Name      string
	Path      string
	Position  int
}

func AddChapter(db *sql.DB, chapter ChapterCreate) (int, error) {
	res, err := db.Exec(`INSERT INTO chapters (name, path, position, parent_id) 
		VALUES (?, ?, ?, ?)`, chapter.Name, chapter.Path, chapter.Position, chapter.ParentID)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func InsertChapterAtPosition(db *sql.DB, chapter ChapterCreate) (int, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && rollbackErr != sql.ErrTxDone {
			fmt.Println("rollback error:", rollbackErr)
		}
	}()
	_, err = tx.Exec(`UPDATE chapters
    SET position = position + 1
    WHERE parent_id=?
    AND position >=?`, chapter.ParentID, chapter.Position)
	if err != nil {
		return 0, err
	}
	res, err := tx.Exec(`INSERT INTO chapters (name, path, position, parent_id)
    VALUES (?, ?, ?, ?)`, chapter.Name, chapter.Path, chapter.Position, chapter.ParentID)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func AddScene(db *sql.DB, scene SceneCreate) (int, error) {
	res, err := db.Exec(`INSERT INTO scenes (chapter_id, name, path, position)
		VALUES (?, ?, ?, ?)`, scene.ChapterID, scene.Name, scene.Path, scene.Position)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func InsertSceneAtPosition(db *sql.DB, scene SceneCreate) (int, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && rollbackErr != sql.ErrTxDone {
			fmt.Println("rollback error:", rollbackErr)
		}
	}()
	_, err = tx.Exec(`UPDATE scenes 
		SET position = position + 1
		WHERE chapter_id=?
		AND position>=?`, scene.ChapterID, scene.Position)
	if err != nil {
		return 0, err
	}
	res, err := tx.Exec(`INSERT INTO scenes (chapter_id, name, path, position)
		VALUES (?, ?, ?, ?)`, scene.ChapterID, scene.Name, scene.Path, scene.Position)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}
