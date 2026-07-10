/*
Package databasehandler
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package databasehandler

import "database/sql"

func UpdateChapterPath(db *sql.DB, chapter *Chapter) error {
	_, err := db.Exec(`UPDATE chapters
		SET path = ?
		WHERE id = ?`, chapter.Path, chapter.ID)
	return err
}

func UpdateScenePath(db *sql.DB, scene *Scene) error {
	_, err := db.Exec(`UPDATE scenes
		SET path = ?
		WHERE id = ?`, scene.Path, scene.ID)
	return err
}
