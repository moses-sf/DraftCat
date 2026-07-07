/*
Package databasehandler
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package databasehandler

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func GetMaxChapterPosition(db *sql.DB, parentID sql.NullInt64) (int, error) {
	var maxPosition int
	res, err := db.Query(`SELECT COALESCE(MAX(position), 0) FROM chapters WHERE parent_id=?`, parentID)
	if err != nil {
		return 0, err
	}
	defer func(r *sql.Rows) {
		err := r.Close()
		if err != nil {
			fmt.Println("error closing rows")
		}
	}(res)
	res.Next()
	err = res.Scan(&maxPosition)
	if err != nil {
		return 0, err
	}
	return maxPosition, nil
}

func GetMaxScenePosition(db *sql.DB, chapter int) (int, error) {
	var maxPosition int
	res := db.QueryRow(`SELECT COALESCE(MAX(position), 0) FROM scenes WHERE chapter_id=?`, chapter)
	err := res.Scan(&maxPosition)
	if err != nil {
		return 0, err
	}
	return maxPosition, nil
}
