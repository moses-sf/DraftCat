/*
Package cmd
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package cmd

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	databasehandler "github.com/moses-sf/DraftCat/databaseHandler"
	"github.com/moses-sf/DraftCat/utilities"
	"github.com/spf13/cobra"
)

type ItemKind string

const (
	Chapter ItemKind = "folder"
	Scene   ItemKind = "scene"
)

type DeleteOptions struct {
	ItemType ItemKind
	ID       int
}

func GenerateDeleteOptions(cmd *cobra.Command) (DeleteOptions, error) {
	folder, err := cmd.Flags().GetBool("folder")
	if err != nil {
		return DeleteOptions{}, err
	}
	scene, err := cmd.Flags().GetBool("scene")
	if err != nil {
		return DeleteOptions{}, err
	}
	id, err := cmd.Flags().GetInt("id")
	if err != nil {
		return DeleteOptions{}, err
	}
	if folder && scene {
		return DeleteOptions{}, errors.New("cannot mark both folder and scene flags")
	}
	var itemKind ItemKind
	if folder {
		itemKind = Chapter
	}
	if scene {
		itemKind = Scene
	}
	return DeleteOptions{ItemType: itemKind, ID: id}, nil
}

func DeleteChapter(db *sql.DB, id int) ([]string, error) {
	return []string{""}, nil
}

func DeleteScene(db *sql.DB, id int) ([]string, error) {
	scene, err := databasehandler.GetScene(db, id)
	if err != nil {
		return nil, err
	}
	chapter, err := databasehandler.GetChapter(db, scene.ChapterID)
	if err != nil {
		return nil, err
	}
	chapterToml, err := utilities.LoadChapterToml(chapter.Path)
	if err != nil {
		return nil, err
	}
	err = databasehandler.DeleteSceneUpdatePosition(db, id, scene.ChapterID, scene.Position)
	if err != nil {
		return nil, err
	}
	err = os.Remove(scene.Path)
	if err != nil {
		return nil, err
	}
	err = chapterToml.RebuildSceneMetadata(db)
	if err != nil {
		return nil, err
	}
	chapterTomlPath := filepath.Join(chapter.Path, ".chapter.toml")
	err = utilities.EncodeToml(chapterTomlPath, chapterToml)
	if err != nil {
		return nil, err
	}
	return []string{scene.Path}, nil
}

func DeleteItem(deleteOpts DeleteOptions) ([]string, error) {
	dbPath, err := utilities.GetDBPath()
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		fmt.Println("Could not access story DB, please run drafcat story repair")
		return nil, err
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Println("Error closing DB:", closeErr)
		}
	}()
	switch deleteOpts.ItemType {
	case Chapter:
		return DeleteChapter(db, deleteOpts.ID)
	case Scene:
		return DeleteScene(db, deleteOpts.ID)
	}
	return nil, errors.New("did not execute")
}

type DeleteJSONResponse struct {
	Status       bool     `json:"status"`
	Error        string   `json:"error"`
	DeletedPaths []string `json:"deleted_paths"`
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a scene or chapter",
	Long:  "Delete a scene or chapter and all the chapters contents",
	Run: func(cmd *cobra.Command, args []string) {
		j, err := cmd.Flags().GetBool("json")
		if err != nil {
			log.Fatalf(`{"status":false, "error":"%s"}`, err)
		}
		deleteOpts, err := GenerateDeleteOptions(cmd)
		if err != nil {
			if j {
				log.Fatalf(`{"status":false, "error":"%s"}`, err)
			} else {
				log.Fatal(err)
			}
			return
		}
		output, err := DeleteItem(deleteOpts)
		if err != nil {
			if j {
				log.Fatalf(`{"status":false, "error":"%s"}`, err)
			} else {
				log.Fatal(err)
			}
			return
		}
		response := DeleteJSONResponse{
			Status:       true,
			Error:        "",
			DeletedPaths: output,
		}
		res, err := json.Marshal(response)
		if err != nil {
			if j {
				log.Fatalf(`{"status":false, "error":"%s"}`, err)
			} else {
				log.Fatal(err)
			}
			return
		}
		fmt.Printf("%s", string(res))
	},
}

func initDeleteCmd() {
	storyCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().BoolP("folder", "f", false, "Delete a folder")
	deleteCmd.Flags().BoolP("scene", "s", false, "Delete a scene")
	deleteCmd.Flags().BoolP("json", "j", false, "Output Json")
	deleteCmd.Flags().IntP("id", "i", 0, "Scene or Folder ID")
}
