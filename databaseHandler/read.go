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
	var err error
	if parentID.Valid {
		err = db.QueryRow(`SELECT COALESCE(MAX(position), 0) FROM chapters WHERE parent_id=?`, parentID).Scan(&maxPosition)
	} else {
		err = db.QueryRow(`SELECT COALESCE(MAX(position), 0) FROM chapters WHERE parent_id IS NULL`, parentID).Scan(&maxPosition)
	}
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

func GetChapterNodes(db *sql.DB) ([]Chapter, error) {
	chapters := make([]Chapter, 0)
	res, err := db.Query(`SELECT id, parent_id, name, path, position, word_count FROM chapters ORDER BY parent_id`)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := res.Close(); closeErr != nil {
			fmt.Println("error closing chapter rows:", closeErr)
		}
	}()
	for res.Next() {
		chapter := Chapter{}
		err = res.Scan(
			&chapter.ID,
			&chapter.ParentID,
			&chapter.Name,
			&chapter.Path,
			&chapter.Position,
			&chapter.WordCount,
		)
		if err != nil {
			return nil, err
		}
		chapters = append(chapters, chapter)
	}

	if err := res.Err(); err != nil {
		return nil, err
	}
	return chapters, nil
}
