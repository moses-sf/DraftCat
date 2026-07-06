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

func GetMaxChapterPosition(db *sql.DB) (int, error) {
	var maxPosition int
	res, err := db.Query(`SELECT COALESCE(MAX(position), 0) FROM chapters`)
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
