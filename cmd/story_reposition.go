/*
Package cmd
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"

	databasehandler "github.com/moses-sf/DraftCat/databaseHandler"
	"github.com/moses-sf/DraftCat/utilities"
	"github.com/spf13/cobra"
)

var repositionCmd = &cobra.Command{
	Use:   "reposition",
	Short: "reposition folders and scenes",
	Long:  "reposition folders and scenes",
}

type FolderRepositionOptions struct {
	ChapterID   int
	NewPosition int
}

func GetFolderRepositionOptions(cmd *cobra.Command) (FolderRepositionOptions, error) {
	if !cmd.Flags().Changed("folderID") || !cmd.Flags().Changed("position") {
		return FolderRepositionOptions{}, fmt.Errorf("folderid and new position not provided")
	}
	folderID, err := cmd.Flags().GetInt("folderID")
	if err != nil {
		return FolderRepositionOptions{}, err
	}
	position, err := cmd.Flags().GetInt("position")
	if err != nil {
		return FolderRepositionOptions{}, err
	}
	return FolderRepositionOptions{ChapterID: folderID, NewPosition: position}, nil
}

func UpdateFolderPosition(db *sql.DB, folderOpts FolderRepositionOptions) error {
	chapter, err := databasehandler.GetChapter(db, folderOpts.ChapterID)
	if err != nil {
		return err
	}
	if chapter.Position == folderOpts.NewPosition {
		return fmt.Errorf("old and new position can't be the same")
	}
	maxFolderPosition, err := databasehandler.GetMaxChapterPosition(db, chapter.ParentID)
	if err != nil {
		return err
	}
	if folderOpts.NewPosition < 1 {
		return fmt.Errorf("new position cannot be lower than 1")
	}
	if maxFolderPosition < folderOpts.NewPosition {
		folderOpts.NewPosition = maxFolderPosition
	}
	err = databasehandler.UpdateChapterPosition(db, chapter.ID, chapter.Position, folderOpts.NewPosition, chapter.ParentID)
	if err != nil {
		return err
	}
	chapterToml, err := utilities.LoadChapterToml(chapter.Path)
	if err != nil {
		return err
	}
	return chapterToml.RebuildParentChapterPositionMetadata(db)
}

func RepositionFolder(folderOpts FolderRepositionOptions) error {
	db, err := utilities.OpenDB()
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Println("Error closing DB:", closeErr)
		}
	}()

	return UpdateFolderPosition(db, folderOpts)
}

var repositionFolderCmd = &cobra.Command{
	Use:   "folder",
	Short: "reposition folders",
	Long:  "reposition folders",
	Run: func(cmd *cobra.Command, args []string) {
		j, err := cmd.Flags().GetBool("json")
		if err != nil {
			state, err := json.Marshal(JSONStatus{
				Status: false,
				Error:  fmt.Sprintf("%s", err),
			})
			if err != nil {
				log.Fatalf(`{"status":false, "error":"%s"}`, err)
				return
			}
			log.Fatal(string(state))
			return
		}
		folderOpts, err := GetFolderRepositionOptions(cmd)
		if err != nil {
			log.Fatalf(`{"status":false, "error":"%s"}`, err)
			return
		}
		err = RepositionFolder(folderOpts)
		if err != nil {
			log.Fatalf(`{"status":false, "error":"%s"}`, err)
			return
		}
		if j {
			state, err := json.Marshal(JSONStatus{
				Status: true,
			})
			if err != nil {
				log.Fatalf(`{"status":false, "error":"%s"}`, err)
				return
			}
			fmt.Println(string(state))
		} else {
			fmt.Println("Other")
		}
	},
}

type RepositionSceneOptions struct {
	SceneID     int
	NewPosition int
}

func GetRepositionSceneOptions(cmd *cobra.Command) (RepositionSceneOptions, error) {
	if !cmd.Flags().Changed("sceneID") || !cmd.Flags().Changed("position") {
		return RepositionSceneOptions{}, fmt.Errorf("sceneid and new position not provided")
	}
	sceneID, err := cmd.Flags().GetInt("sceneID")
	if err != nil {
		return RepositionSceneOptions{}, err
	}
	position, err := cmd.Flags().GetInt("position")
	if err != nil {
		return RepositionSceneOptions{}, err
	}
	return RepositionSceneOptions{SceneID: sceneID, NewPosition: position}, nil
}

func RepositionScene(sceneOpts RepositionSceneOptions) error {
	db, err := utilities.OpenDB()
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Println("Error closing DB:", closeErr)
		}
	}()
	scene, err := databasehandler.GetScene(db, sceneOpts.SceneID)
	if err != nil {
		return err
	}
	if scene.Position == sceneOpts.NewPosition {
		return fmt.Errorf("old and new position can't be the same")
	}
	chapter, err := databasehandler.GetChapter(db, scene.ChapterID)
	if err != nil {
		return err
	}
	chapterTomlPath := filepath.Join(chapter.Path, ".chapter.toml")
	chapterToml, err := utilities.LoadChapterToml(chapter.Path)
	if err != nil {
		return err
	}
	maxScenePosition, err := databasehandler.GetMaxScenePosition(db, chapter.ID)
	if err != nil {
		return err
	}
	if sceneOpts.NewPosition < 1 {
		return fmt.Errorf("new position cannot be lower than 1")
	}
	if maxScenePosition < sceneOpts.NewPosition {
		sceneOpts.NewPosition = maxScenePosition
	}
	err = databasehandler.UpdateScenePosition(db, sceneOpts.SceneID, scene.Position, sceneOpts.NewPosition, chapter.ID)
	if err != nil {
		return err
	}
	err = chapterToml.RebuildSceneMetadata(db)
	if err != nil {
		return err
	}
	return utilities.EncodeToml(chapterTomlPath, chapterToml)
}

var repositionSceneCmd = &cobra.Command{
	Use:   "scene",
	Short: "reposition scene",
	Long:  "reposition scenes",
	Run: func(cmd *cobra.Command, args []string) {
		j, err := cmd.Flags().GetBool("json")
		if err != nil {
			state, err := json.Marshal(JSONStatus{
				Status: false,
				Error:  fmt.Sprintf("%s", err),
			})
			if err != nil {
				log.Fatalf(`{"status":false, "error":"%s"}`, err)
				return
			}
			log.Fatal(string(state))
			return
		}

		sceneOpts, err := GetRepositionSceneOptions(cmd)
		if err != nil {
			log.Fatalf(`{"status":false, "error":"%s"}`, err)
			return
		}

		err = RepositionScene(sceneOpts)
		if err != nil {
			log.Fatalf(`{"status":false, "error":"%s"}`, err)
			return
		}

		if j {
			state, err := json.Marshal(JSONStatus{
				Status: true,
			})
			if err != nil {
				log.Fatalf(`{"status":false, "error":"%s"}`, err)
				return
			}
			fmt.Println(string(state))
		} else {
			fmt.Println("Other")
		}
	},
}

func initRepositionCmd() {
	storyCmd.AddCommand(repositionCmd)

	repositionCmd.AddCommand(repositionFolderCmd)
	repositionCmd.AddCommand(repositionSceneCmd)

	repositionSceneCmd.Flags().IntP("sceneID", "s", 0, "Scene ID to be changed")
	repositionSceneCmd.Flags().IntP("position", "p", 0, "New Position of the scene to be moved")
	repositionSceneCmd.Flags().BoolP("json", "j", false, "Json output")
	repositionFolderCmd.Flags().IntP("folderID", "f", 0, "Folder ID to be changed")
	repositionFolderCmd.Flags().IntP("position", "p", 0, "New Position of the folder to be moved")
	repositionFolderCmd.Flags().BoolP("json", "j", false, "Json output")
}
