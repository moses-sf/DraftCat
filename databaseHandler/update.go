/*
Package databasehandler
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package databasehandler

import "database/sql"

func UpdateChapterPathAndName(db *sql.DB, chapter *Chapter) error {
	_, err := db.Exec(`UPDATE chapters
		SET path = ?,
		name = ?
		WHERE id = ?`, chapter.Path, chapter.Name, chapter.ID)
	return err
}

func UpdateScenePath(db *sql.DB, scene *Scene) error {
	_, err := db.Exec(`UPDATE scenes
		SET path = ?
		WHERE id = ?`, scene.Path, scene.ID)
	return err
}

func UpdateScenePathAndName(db *sql.DB, scene *Scene) error {
	_, err := db.Exec(`UPDATE scenes
		SET path = ?,
		name = ?
		WHERE id = ?`, scene.Path, scene.Name, scene.ID)
	return err
}
