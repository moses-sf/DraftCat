/*
Package databasehandler
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package databasehandler

import "database/sql"

type ChapterCreate struct {
	Name string
	Path string
}

func AddChapter(db *sql.DB, chapter ChapterCreate) error {
	_, err := db.Exec(`INSERT INTO chapters (name, path) 
		VALUES (?, ?)`, chapter.Name, chapter.Path)
	return err
}
