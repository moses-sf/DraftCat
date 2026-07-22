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

func GetChapter(db *sql.DB, id int) (Chapter, error) {
	chapter := Chapter{}
	res := db.QueryRow("SELECT id, parent_id,depth, name, compile, path, position, word_count FROM chapters WHERE id=?", id)
	err := res.Scan(
		&chapter.ID,
		&chapter.ParentID,
		&chapter.Depth,
		&chapter.Name,
		&chapter.Compile,
		&chapter.Path,
		&chapter.Position,
		&chapter.WordCount,
	)
	if err != nil {
		return Chapter{}, err
	}
	return EnrichChapterWithAncestorsAndDescendants(db, chapter)
}

func EnrichChapterWithAncestorsAndDescendants(db *sql.DB, chapter Chapter) (Chapter, error) {
	ancestorIDs := make([]int, 0)
	descendantIDs := make([]int, 0)
	res, err := db.Query(`
		WITH RECURSIVE ancestors AS (
			SELECT parent.id,
						 parent.parent_id,
						 1 AS depth
		  FROM chapters AS current
			JOIN chapters AS parent
				ON current.parent_id = parent.id
			WHERE current.id = ?

			UNION ALL

			SELECT chapters.id,
						 chapters.parent_id,
						 ancestors.depth + 1
			FROM chapters
			JOIN ancestors
				ON chapters.id = ancestors.parent_id
		)
		SELECT id
		FROM ancestors
		ORDER BY id DESC`, chapter.ID)
	if err != nil {
		return Chapter{}, err
	}
	defer func() {
		if closeErr := res.Close(); closeErr != nil {
			fmt.Println("error closing chapter rows:", closeErr)
		}
	}()
	for res.Next() {
		var ancestor int
		err = res.Scan(
			&ancestor,
		)
		if err != nil {
			return Chapter{}, err
		}
		ancestorIDs = append(ancestorIDs, ancestor)
	}
	if err := res.Err(); err != nil {
		return Chapter{}, err
	}
	res, err = db.Query(`
		WITH RECURSIVE descendants AS (
			SELECT child.id,
						 child.parent_id,
						 1 AS depth
		  FROM chapters AS current
			JOIN chapters AS child
				ON current.id = child.parent_id
			WHERE current.id = ?

			UNION ALL

			SELECT chapters.id,
						 chapters.parent_id,
						 descendants.depth + 1
			FROM chapters
			JOIN descendants
				ON chapters.parent_id = descendants.id
		)
		SELECT id
		FROM descendants
		ORDER BY id DESC`, chapter.ID)
	if err != nil {
		return Chapter{}, err
	}
	defer func() {
		if closeErr := res.Close(); closeErr != nil {
			fmt.Println("error closing chapter rows:", closeErr)
		}
	}()
	for res.Next() {
		var descendant int
		err = res.Scan(
			&descendant,
		)
		if err != nil {
			return Chapter{}, err
		}
		descendantIDs = append(descendantIDs, descendant)
	}
	if err := res.Err(); err != nil {
		return Chapter{}, err
	}
	chapter.AncestorIDs = ancestorIDs
	chapter.DescendantIDs = descendantIDs
	return chapter, nil
}

func GetChaptersWithParent(db *sql.DB, parentID sql.NullInt64) ([]Chapter, error) {
	chapters := make([]Chapter, 0)
	var res *sql.Rows
	var err error
	if parentID.Valid {
		res, err = db.Query(`SELECT id, parent_id, name, compile, path, position, word_count FROM chapters WHERE parent_id=?`, parentID.Int64)
	} else {
		res, err = db.Query(`SELECT id, parent_id, name, compile, path, position, word_count FROM chapters WHERE parent_id IS NULL`)
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
		chapter := Chapter{}
		err = res.Scan(
			&chapter.ID,
			&chapter.ParentID,
			&chapter.Name,
			&chapter.Compile,
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

type TreeList struct {
	Ancestors   []int
	Descendants []int
}

func GetChapterNodes(db *sql.DB) ([]Chapter, error) {
	chapters := make([]Chapter, 0)
	res, err := db.Query(`SELECT id, parent_id, name, compile, path, position, word_count FROM chapters ORDER BY depth, parent_id, position`)
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
			&chapter.Compile,
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
	chapterMap := make(map[int]*TreeList, 0)
	for _, chapter := range chapters {
		chapterMap[chapter.ID] = &TreeList{
			Ancestors:   make([]int, 0),
			Descendants: make([]int, 0),
		}
	}
	for _, chapter := range chapters {
		if chapter.ParentID.Valid {
			chapterMap[chapter.ID].Ancestors = append(chapterMap[chapter.ID].Ancestors, int(chapter.ParentID.Int64))
			chapterMap[chapter.ID].Ancestors = append(
				chapterMap[chapter.ID].Ancestors,
				chapterMap[int(chapter.ParentID.Int64)].Ancestors...,
			)
			for _, ancestorID := range chapterMap[chapter.ID].Ancestors {
				chapterMap[ancestorID].Descendants = append(
					chapterMap[ancestorID].Descendants,
					chapter.ID,
				)
			}
		}
	}
	for index := range chapters {
		chapters[index].AncestorIDs = append(
			chapters[index].AncestorIDs,
			chapterMap[chapters[index].ID].Ancestors...,
		)

		chapters[index].DescendantIDs = append(
			chapters[index].DescendantIDs,
			chapterMap[chapters[index].ID].Descendants...,
		)
	}

	return chapters, nil
}

func GetScene(db *sql.DB, id int) (Scene, error) {
	scene := Scene{}
	res := db.QueryRow("SELECT id, chapter_id, name, compile, path, position, word_count FROM scenes WHERE id=?", id)
	err := res.Scan(
		&scene.ID,
		&scene.ChapterID,
		&scene.Name,
		&scene.Compile,
		&scene.Path,
		&scene.Position,
		&scene.WordCount,
	)
	if err != nil {
		return Scene{}, err
	}
	return scene, nil
}

func GetScenesOfChapter(db *sql.DB, chapterID int) ([]Scene, error) {
	scenes := make([]Scene, 0)
	res, err := db.Query(`SELECT id, chapter_id, name, compile, path, position, word_count FROM scenes WHERE chapter_id = ?`, chapterID)
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
			&scene.Compile,
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
	res, err := db.Query(`SELECT id, chapter_id, name, compile, path, position, word_count FROM scenes ORDER BY chapter_id, position`)
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
			&scene.Compile,
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

func GetChapterDescendants(db *sql.DB, chapterID int) ([]int, error) {
	descendants := make([]int, 0)
	res, err := db.Query(`
		WITH RECURSIVE descendants AS (
			SELECT
		    child.id,
		    child.parent_id,
		    1 as depth
		  FROM chapters AS current
		  JOIN chapters AS child
		    ON current.id = child.parent_id
		  WHERE current.id = ?

		  UNION ALL

		  SELECT 
				child.id,
				child.parent_id,
		    descendants.depth + 1
		  FROM chapters AS child
		  JOIN descendants
		    ON child.parent_id == descendants.id
		)
		SELECT id
		FROM descendants
		ORDER BY depth DESC`, chapterID)
	if err != nil {
		return nil, err
	}

	defer func() {
		if closeErr := res.Close(); closeErr != nil {
			fmt.Println("error closing chapter rows:", closeErr)
		}
	}()

	for res.Next() {
		id := 0
		err = res.Scan(&id)
		if err != nil {
			return nil, err
		}
		descendants = append(descendants, id)
	}
	if err := res.Err(); err != nil {
		return nil, err
	}
	return descendants, nil
}
