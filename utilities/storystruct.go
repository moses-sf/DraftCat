/*
Package utilities
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package utilities

import (
	"encoding/json"
	"fmt"

	databasehandler "github.com/moses-sf/DraftCat/databaseHandler"
)

type StoryStructure struct {
	Name     string                       `json:"name"`
	Root     string                       `json:"root"`
	Type     string                       `json:"type"`
	RootNode *databasehandler.ChapterNode `json:"root_node"`
}

func (s *StoryStructure) JSONRender() {
	j, err := json.Marshal(s)
	if err != nil {
		fmt.Println(`{"error": "could not render struct"}`)
		return
	}
	fmt.Println(string(j))
}

func (s *StoryStructure) Render() {
	fmt.Println("Name: ", s.Name)
	fmt.Println("Type: ", s.Type)
	fmt.Println("Root: ", s.Root)
	s.RootNode.Render()
}
