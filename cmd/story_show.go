/*
Package cmd
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package cmd

import (
	"fmt"
	"log"

	databasehandler "github.com/moses-sf/DraftCat/databaseHandler"
	"github.com/moses-sf/DraftCat/utilities"
	"github.com/spf13/cobra"
)

func BuildStoryProject() (*utilities.StoryStructure, error) {
	_, err := utilities.IsDraftcatProject()
	if err != nil {
		return nil, err
	}
	root, err := utilities.GetRelativeRootPath()
	if err != nil {
		return nil, err
	}

	db, err := utilities.OpenDB()
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Println("Error closing DB:", closeErr)
		}
	}()
	chapters, err := databasehandler.GetChapterNodes(db)
	if err != nil {
		return nil, err
	}

	scenes, err := databasehandler.GetSceneNodes(db)
	if err != nil {
		return nil, err
	}
	rootNode, err := databasehandler.MapNodes(chapters, scenes)
	if err != nil {
		return nil, err
	}
	storyConfig := &utilities.StoryConfig{}
	err = storyConfig.LoadConfig(root)
	if err != nil {
		return nil, err
	}
	return &utilities.StoryStructure{
		Name:     storyConfig.MetaData.Name,
		Root:     root,
		Type:     string(storyConfig.MetaData.Type),
		RootNode: rootNode,
	}, nil
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "display the story structure",
	Long:  "display the story structure, use j for json output",
	Run: func(cmd *cobra.Command, args []string) {
		json, err := cmd.Flags().GetBool("json")
		if err != nil {
			fmt.Println(err)
			return
		}
		story, err := BuildStoryProject()
		if err != nil {
			fmt.Println(err)
			return
		}
		if json {
			story.JSONRender()
		} else {
			fmt.Println("Displaying story structure")
			story.Render()
		}
	},
}

func initShowCmd() {
	storyCmd.AddCommand(showCmd)

	showCmd.Flags().BoolP("json", "j", false, "Output json to stdout")
}
