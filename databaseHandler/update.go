/*
Package databasehandler
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package databasehandler

import (
	"database/sql"
	"errors"
)

func UpdateChapterPathAndName(db *sql.DB, chapter Chapter) error {
	_, err := db.Exec(`UPDATE chapters
		SET path = ?,
		name = ?
		WHERE id = ?`, chapter.Path, chapter.Name, chapter.ID)
	return err
}

func UpdateChapterPath(db *sql.DB, chapter Chapter) error {
	_, err := db.Exec(`UPDATE chapters
		SET path = ?
		WHERE id =?`, chapter.Path, chapter.ID)
	return err
}

func UpdateChapterPathAndDepth(db *sql.DB, chapter Chapter) error {
	_, err := db.Exec(`UPDATE chapters
		SET path = ?,
				depth = ?
		WHERE id =?`, chapter.Path, chapter.Depth, chapter.ID)
	return err
}

func UpdateChapterPosition(db *sql.DB, chapterID, oldPosition, newPosition int, rootID sql.NullInt64) error {
	if oldPosition == newPosition {
		return errors.New("old and new position cannot be equal")
	}
	if oldPosition < newPosition {
		return UpdateChapterPositionIncrease(db, chapterID, oldPosition, newPosition, rootID)
	} else {
		return UpdateChapterPositionDecrease(db, chapterID, oldPosition, newPosition, rootID)
	}
}

func UpdateChapterPositionIncrease(db *sql.DB, chapterID, oldPosition, newPosition int, rootID sql.NullInt64) error {
	if rootID.Valid {
		_, err := db.Exec(`UPDATE chapters
		SET position = CASE
			WHEN id = ? THEN ?
			WHEN position > ? AND position <= ? THEN position - 1
			ELSE position
		END
		WHERE parent_id = ? AND (id = ? OR (position > ? AND position <= ?))`, chapterID, newPosition, oldPosition, newPosition, rootID.Int64, chapterID, oldPosition, newPosition)
		return err
	} else {
		_, err := db.Exec(`UPDATE chapters
		SET position = CASE
			WHEN id = ? THEN ?
			WHEN position > ? AND position <= ? THEN position - 1
			ELSE position
		END
		WHERE parent_id IS NULL AND (id = ? OR (position > ? AND position <= ?))`, chapterID, newPosition, oldPosition, newPosition, chapterID, oldPosition, newPosition)
		return err
	}
}

func UpdateChapterParentPathAppendPosition(db *sql.DB, chapter Chapter) error {
	if chapter.ParentID.Valid {
		_, err := db.Exec(`
		UPDATE chapters
			SET parent_id = ?,
			 		path = ?,
					position = ?,
					depth = ?
		WHERE id = ?`, chapter.ParentID.Int64, chapter.Path, chapter.Position, chapter.Depth, chapter.ID)
		if err != nil {
			return err
		}
	} else {
		_, err := db.Exec(`
		UPDATE chapters
			SET parent_id = NULL,
			 		path = ?,
					position = ?,
					depth = ?
		WHERE id = ?`, chapter.Path, chapter.Position, chapter.Depth, chapter.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateChapterPositionDecrease(db *sql.DB, chapterID, oldPosition, newPosition int, rootID sql.NullInt64) error {
	if rootID.Valid {
		_, err := db.Exec(`UPDATE chapters
		SET position = CASE
			WHEN id = ? THEN ?
			WHEN position < ? AND position >= ? THEN position + 1
			ELSE position
		END
		WHERE parent_id = ? AND (id = ? OR (position < ? AND position >= ?))`, chapterID, newPosition, oldPosition, newPosition, rootID.Int64, chapterID, oldPosition, newPosition)
		return err
	} else {
		_, err := db.Exec(`UPDATE chapters
		SET position = CASE
			WHEN id = ? THEN ?
			WHEN position < ? AND position >= ? THEN position + 1
			ELSE position
		END
		WHERE parent_id IS NULL AND (id = ? OR (position < ? AND position >= ?))`, chapterID, newPosition, oldPosition, newPosition, chapterID, oldPosition, newPosition)
		return err
	}
}

func UpdateScenePath(db *sql.DB, scene Scene) error {
	_, err := db.Exec(`UPDATE scenes
		SET path = ?
		WHERE id = ?`, scene.Path, scene.ID)
	return err
}

func UpdateScenePathAndChapter(db *sql.DB, scene Scene) error {
	_, err := db.Exec(`UPDATE scenes
		SET path = ?,
		    chapter_id = ?
		WHERE id = ?`, scene.Path, scene.ChapterID, scene.ID)
	return err
}

func UpdateScenePathChapterAppendPosition(db *sql.DB, scene Scene) error {
	_, err := db.Exec(`UPDATE scenes
		SET path = ?,
		    chapter_id = ?,
				position = ?
		WHERE id = ?`, scene.Path, scene.ChapterID, scene.Position, scene.ID)
	return err
}

func UpdateScenePathAndName(db *sql.DB, scene Scene) error {
	_, err := db.Exec(`UPDATE scenes
		SET path = ?,
		name = ?
		WHERE id = ?`, scene.Path, scene.Name, scene.ID)
	return err
}

func UpdateScenePosition(db *sql.DB, sceneID, oldPosition, newPosition, chapterID int) error {
	if newPosition > oldPosition {
		_, err := db.Exec(`UPDATE scenes
		SET position = CASE 
			WHEN id = ? THEN ?
			WHEN position > ? AND position <= ? THEN position - 1
			ELSE position
		END
		WHERE chapter_id = ? AND (id = ? OR (position > ? AND position <= ?))`, sceneID, newPosition, oldPosition, newPosition, chapterID, sceneID, oldPosition, newPosition)
		return err
	} else {
		_, err := db.Exec(`UPDATE scenes
		SET position = CASE 
			WHEN id = ? THEN ?
			WHEN position < ? AND position >= ? THEN position + 1
			ELSE position
		END
		WHERE chapter_id =? AND (id = ? OR (position < ? AND position >= ?))`, sceneID, newPosition, oldPosition, newPosition, chapterID, sceneID, oldPosition, newPosition)
		return err
	}
}

func UpdateSceneCompile(db *sql.DB, sceneID int, compile bool) error {
	_, err := db.Exec(`UPDATE scenes
		SET compile = ?
		WHERE id = ?`, compile, sceneID)
	return err
}

func UpdateSceneCompileChapter(db *sql.DB, chapterID int, compile bool) error {
	_, err := db.Exec(`UPDATE scenes
		SET compile = ?
		WHERE chapter_id = ?`, compile, chapterID)
	return err
}

func UpdateChapterCompile(db *sql.DB, chapterID int, compile bool) error {
	_, err := db.Exec(`
		UPDATE chapters
		SET compile = ?
		WHERE id = ?`, compile, chapterID)
	return err
}
