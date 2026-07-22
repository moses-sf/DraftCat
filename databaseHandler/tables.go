/*
Package databasehandler
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package databasehandler

import (
	"database/sql"
	"fmt"
)

type Chapter struct {
	ID            int
	ParentID      sql.NullInt64
	Name          string
	Depth         int
	Compile       bool
	Path          string
	Position      int
	WordCount     int
	AncestorIDs   []int
	DescendantIDs []int
}

type Scene struct {
	ID          int
	ChapterID   int
	Compile     bool
	Name        string
	Path        string
	Position    int
	WordCount   int
	AncestorIDs []int
}

type SceneNode struct {
	Scene Scene
	Depth int
}

func (s *SceneNode) Render() {
	indent := ""
	for range s.Depth {
		indent = fmt.Sprintf("%s  ", indent)
	}
	fmt.Println(indent, s.Scene)
}

type ChapterNode struct {
	Chapter  Chapter
	Depth    int
	Chapters []*ChapterNode
	Scenes   []*SceneNode
}

func (c *ChapterNode) Render() {
	indent := ""
	for range c.Depth {
		indent = fmt.Sprintf("%s  ", indent)
	}
	fmt.Println(indent, c.Chapter)
	for _, chapter := range c.Chapters {
		chapter.Render()
	}
	for _, scene := range c.Scenes {
		scene.Render()
	}
}

func MapNodes(chapters []Chapter, scenes []Scene) (*ChapterNode, error) {
	nodeMap := make(map[int]*ChapterNode, 0)
	nodeMap[0] = &ChapterNode{
		Depth:    0,
		Chapters: make([]*ChapterNode, 0),
		Scenes:   make([]*SceneNode, 0),
	}
	for _, c := range chapters {
		var parentID int
		if c.ParentID.Valid {
			parentID = int(c.ParentID.Int64)
		} else {
			parentID = 0
		}
		chapterNode := &ChapterNode{
			Chapter:  c,
			Chapters: make([]*ChapterNode, 0),
			Scenes:   make([]*SceneNode, 0),
		}
		_, ok := nodeMap[c.ID]
		if !ok {
			nodeMap[c.ID] = chapterNode
		}
		node, ok := nodeMap[parentID]
		if !ok {
			return nil, fmt.Errorf("missing id in parent map: %d", parentID)
		}
		node.Chapters = append(node.Chapters, chapterNode)
		chapterNode.Depth = node.Depth + 1
	}

	for _, s := range scenes {
		s.AncestorIDs = append(s.AncestorIDs, nodeMap[s.ChapterID].Chapter.AncestorIDs...)
		s.AncestorIDs = append(s.AncestorIDs, nodeMap[s.ChapterID].Chapter.ID)
		sceneNode := &SceneNode{
			Scene: s,
			Depth: 0,
		}
		node, ok := nodeMap[s.ChapterID]
		if !ok {
			return nil, fmt.Errorf("missing id in chapter map: %d", s.ChapterID)
		}
		sceneNode.Depth = node.Depth + 1
		node.Scenes = append(node.Scenes, sceneNode)
	}

	return nodeMap[0], nil
}
