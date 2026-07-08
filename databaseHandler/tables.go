/*
Package databasehandler
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package databasehandler

import "database/sql"

type Chapter struct {
	ID        int
	ParentID  sql.NullInt64
	Name      string
	Path      string
	Position  int
	WordCount int
}

type Scene struct {
	ID        int
	ChapterID int
	Name      string
	Path      string
	Position  int
	WordCount int
}

type ChapterNode struct {
	Chapter  Chapter
	Chapters []ChapterNode
	Scenes   []Scene
}
