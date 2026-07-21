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

type ItemKindOptions struct {
	ItemType ItemKind
	ID       int
}

func GenerateItemKindOptions(cmd *cobra.Command) (ItemKindOptions, error) {
	folder, err := cmd.Flags().GetBool("folder")
	if err != nil {
		return ItemKindOptions{}, err
	}
	scene, err := cmd.Flags().GetBool("scene")
	if err != nil {
		return ItemKindOptions{}, err
	}
	id, err := cmd.Flags().GetInt("id")
	if err != nil {
		return ItemKindOptions{}, err
	}
	if folder && scene {
		return ItemKindOptions{}, errors.New("cannot mark both folder and scene flags")
	}
	var itemKind ItemKind
	if folder {
		itemKind = Chapter
	}
	if scene {
		itemKind = Scene
	}
	return ItemKindOptions{ItemType: itemKind, ID: id}, nil
}

func DeleteChapter(db *sql.DB, id int) ([]string, error) {
	deletePaths := make([]string, 0)
	chapter, err := databasehandler.GetChapter(db, id)
	if err != nil {
		return nil, err
	}
	childChapters, err := databasehandler.GetChaptersWithParent(db, sql.NullInt64{Valid: true, Int64: int64(chapter.ID)})
	if err != nil {
		return nil, err
	}
	if len(childChapters) > 0 {
		for _, childChapter := range childChapters {
			paths, err := DeleteChapter(db, childChapter.ID)
			if err != nil {
				return nil, err
			}
			deletePaths = append(deletePaths, paths...)
		}
	}
	childScenes, err := databasehandler.GetScenesOfChapter(db, chapter.ID)
	if err != nil {
		return nil, err
	}
	for _, childScene := range childScenes {
		deletePaths = append(deletePaths, childScene.Path)
		err := os.Remove(childScene.Path)
		if err != nil {
			return nil, err
		}
	}
	err = databasehandler.DeleteScenesFromChapter(db, chapter.ID)
	if err != nil {
		return nil, err
	}
	err = os.RemoveAll(chapter.Path)
	if err != nil {
		return nil, err
	}
	err = databasehandler.DeleteChapterUpdatePosition(db, chapter.ID, chapter.Position, chapter.ParentID)
	if err != nil {
		return nil, err
	}
	deletePaths = append(deletePaths, chapter.Path)
	return deletePaths, nil
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

func DeleteItem(deleteOpts ItemKindOptions) ([]string, error) {
	db, err := utilities.OpenDB()
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
		_, err := utilities.IsDraftcatProject()
		if err != nil {
			log.Fatalf("%s", err)
		}
		j, err := cmd.Flags().GetBool("json")
		if err != nil {
			log.Fatalf(`{"status":false, "error":"%s"}`, err)
		}
		deleteOpts, err := GenerateItemKindOptions(cmd)
		if err != nil {
			if j {
				log.Fatalf(`{"status":false, "error":"%s"}`, err)
			} else {
				log.Fatal(err)
			}
			return
		}
		message := fmt.Sprintf("draftcat|backup|delete|%s|%d", deleteOpts.ItemType, deleteOpts.ID)
		err = utilities.CommitBackupSnapshot(message)
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
			errRestore := utilities.RestoreChanges()
			if errRestore != nil {
				err = fmt.Errorf("%w-%w", err, errRestore)
			}
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
