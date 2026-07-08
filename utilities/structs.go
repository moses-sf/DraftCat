/*
Package utilities
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package utilities

import "database/sql"

type SceneMetaData struct {
	ID       int
	Name     string
	Path     string
	Position int
}

type ChapterMetaData struct {
	ID         int
	ParentID   sql.NullInt64
	PathToRoot string
	Position   int
	Scenes     []SceneMetaData
}
