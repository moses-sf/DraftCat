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
		err = db.QueryRow(`SELECT COALESCE(MAX(position), 0) FROM chapters WHERE parent_id=?`, parentID.Int64).Scan(&maxPosition)
	} else {
		err = db.QueryRow(`SELECT COALESCE(MAX(position), 0) FROM chapters WHERE parent_id IS NULL`).Scan(&maxPosition)
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

func GetChapter(db *sql.DB, id int) (*Chapter, error) {
	chapter := &Chapter{}
	res := db.QueryRow("SELECT id, parent_id, name, path, position, word_count FROM chapters WHERE id=?", id)
	err := res.Scan(
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
	return chapter, nil
}

func GetChaptersWithParent(db *sql.DB, parentID sql.NullInt64) ([]*Chapter, error) {
	chapters := make([]*Chapter, 0)
	var res *sql.Rows
	var err error
	if parentID.Valid {
		res, err = db.Query(`SELECT id, parent_id, name, path, position, word_count FROM chapters WHERE parent_id=?`, parentID.Int64)
	} else {
		res, err = db.Query(`SELECT id, parent_id, name, path, position, word_count FROM chapters WHERE parent_id IS NULL`)
	}
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := res.Close(); closeErr != nil {
			fmt.Println("error closing chapter rows:", closeErr)
		}
	}()
	for res.Next() {
		chapter := &Chapter{}
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

func GetChapterNodes(db *sql.DB) ([]Chapter, error) {
	chapters := make([]Chapter, 0)
	res, err := db.Query(`SELECT id, parent_id, name, path, position, word_count FROM chapters ORDER BY parent_id, position`)
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

func GetScene(db *sql.DB, id int) (*Scene, error) {
	scene := &Scene{}
	res := db.QueryRow("SELECT id, chapter_id, name, path, position, word_count FROM scenes WHERE id=?", id)
	err := res.Scan(
		&scene.ID,
		&scene.ChapterID,
		&scene.Name,
		&scene.Path,
		&scene.Position,
		&scene.WordCount,
	)
	if err != nil {
		return nil, err
	}
	return scene, nil
}

func GetScenesOfChapter(db *sql.DB, chapterID int) ([]Scene, error) {
	scenes := make([]Scene, 0)
	res, err := db.Query(`SELECT id, chapter_id, name, path, position, word_count FROM scenes WHERE chapter_id = ?`, chapterID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := res.Close(); closeErr != nil {
			fmt.Println("error closing chapter rows:", closeErr)
		}
	}()
	for res.Next() {
		scene := Scene{}
		err = res.Scan(
			&scene.ID,
			&scene.ChapterID,
			&scene.Name,
			&scene.Path,
			&scene.Position,
			&scene.WordCount,
		)
		if err != nil {
			return nil, err
		}
		scenes = append(scenes, scene)
	}
	if err := res.Err(); err != nil {
		return nil, err
	}
	return scenes, nil
}

func GetSceneNodes(db *sql.DB) ([]Scene, error) {
	scenes := make([]Scene, 0)
	res, err := db.Query(`SELECT id, chapter_id, name, path, position, word_count FROM scenes ORDER BY chapter_id, position`)
	if err != nil {
		return nil, err
	}

	defer func() {
		if closeErr := res.Close(); closeErr != nil {
			fmt.Println("error closing chapter rows:", closeErr)
		}
	}()

	for res.Next() {
		scene := Scene{}
		err = res.Scan(
			&scene.ID,
			&scene.ChapterID,
			&scene.Name,
			&scene.Path,
			&scene.Position,
			&scene.WordCount,
		)
		if err != nil {
			return nil, err
		}
		scenes = append(scenes, scene)
	}
	if err := res.Err(); err != nil {
		return nil, err
	}
	return scenes, nil
}
