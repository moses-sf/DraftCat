/*
Package databasehandler
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package databasehandler

import "database/sql"

func DeleteScene(db *sql.DB, id, chapterID, position int) error {
	_, err := db.Exec(`
		DELETE FROM scenes
		WHERE id = ? AND chapter_id = ?`, id, chapterID)
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
