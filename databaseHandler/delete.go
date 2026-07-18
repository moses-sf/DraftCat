/*
Package databasehandler
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package databasehandler

import "database/sql"

func DeleteSceneUpdatePosition(db *sql.DB, id, chapterID, position int) error {
	_, err := db.Exec(`
		DELETE FROM scenes
		WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return UpdateDeletedScenePosition(db, chapterID, position)
}

func UpdateDeletedScenePosition(db *sql.DB, chapterID, position int) error {
	_, err := db.Exec(`
		UPDATE scenes
		SET position = position - 1
		WHERE position >= ? AND chapter_id = ?`, position, chapterID)
	return err
}

func DeleteScene(db *sql.DB, id int) error {
	_, err := db.Exec(`
		DELETE FROM scenes
		WHERE id = ?`, id)
	return err
}

func DeleteScenesFromChapter(db *sql.DB, chapterID int) error {
	_, err := db.Exec(`
		DELETE FROM scenes
		WHERE chapter_id = ?`, chapterID)
	return err
}

func DeleteChapterUpdatePosition(db *sql.DB, id, position int, parentID sql.NullInt64) error {
	_, err := db.Exec(`
		DELETE FROM chapters
		WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return UpdateDeletedChapterPosition(db, position, parentID)
}

func UpdateDeletedChapterPosition(db *sql.DB, position int, parentID sql.NullInt64) error {
	if parentID.Valid {
		_, err := db.Exec(`
		UPDATE chapters
		SET position = position - 1
		WHERE position >= ? AND parent_id = ?`, position, parentID)
		return err
	} else {
		_, err := db.Exec(`
		UPDATE chapters
		SET position = position - 1
		WHERE position >= ? AND parent_id IS NULL`, position)
		return err
	}
}
