/*
Package cmd
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	databasehandler "github.com/moses-sf/DraftCat/databaseHandler"
	"github.com/moses-sf/DraftCat/utilities"
	"github.com/spf13/cobra"
)

var moveCmd = &cobra.Command{
	Use:   "move",
	Short: "Move a folder or Scene",
	Long:  "Move a folder or scene",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := utilities.IsDraftcatProject()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Select a Subcommand")
	},
}

type MoveSceneOptions struct {
	SceneID      int
	OldChapterID int
	NewChapterID int
}

func GenerateMoveSceneOptions(db *sql.DB, cmd *cobra.Command, args []string) (MoveSceneOptions, error) {
	sceneID, err := cmd.Flags().GetInt("sceneID")
	if err != nil {
		return MoveSceneOptions{}, err
	}
	folderID, err := cmd.Flags().GetInt("folderID")
	if err != nil {
		return MoveSceneOptions{}, err
	}
	if sceneID <= 0 {
		return MoveSceneOptions{}, errors.New("scene ID cannot be less than 1")
	}
	if folderID <= 0 {
		return MoveSceneOptions{}, errors.New("target of reroot cannot be root or less than 0")
	}
	scene, err := databasehandler.GetScene(db, sceneID)
	if err != nil {
		return MoveSceneOptions{}, err
	}
	return MoveSceneOptions{
		SceneID:      sceneID,
		OldChapterID: scene.ChapterID,
		NewChapterID: folderID,
	}, nil
}

func MoveScene(db *sql.DB, moveSceneOptions MoveSceneOptions) error {
	scene, err := databasehandler.GetScene(db, moveSceneOptions.SceneID)
	if err != nil {
		return err
	}
	oldChapter, err := databasehandler.GetChapter(db, moveSceneOptions.OldChapterID)
	if err != nil {
		return err
	}
	oldChapterToml, err := utilities.LoadChapterToml(oldChapter.Path)
	if err != nil {
		return err
	}
	newChapter, err := databasehandler.GetChapter(db, moveSceneOptions.NewChapterID)
	if err != nil {
		return err
	}
	newChapterToml, err := utilities.LoadChapterToml(newChapter.Path)
	if err != nil {
		return err
	}
	maxPosition, err := databasehandler.GetMaxScenePosition(db, newChapter.ID)
	if err != nil {
		return err
	}
	newPath := filepath.Join(newChapter.Path, filepath.Base(scene.Path))
	if utilities.FileExists(newPath) {
		return errors.New("scene of same name exists at other chapter")
	}
	err = os.Rename(scene.Path, newPath)
	if err != nil {
		return err
	}
	scene.Path = newPath
	scene.ChapterID = newChapter.ID
	scene.Position = maxPosition + 1
	err = databasehandler.UpdateScenePathAndChapterAndPosition(db, scene)
	if err != nil {
		return err
	}
	err = oldChapterToml.RebuildSceneMetadata(db)
	if err != nil {
		return err
	}
	err = newChapterToml.RebuildSceneMetadata(db)
	if err != nil {
		return err
	}
	err = utilities.EncodeChapterToml(oldChapter.Path, *oldChapterToml)
	if err != nil {
		return err
	}
	return utilities.EncodeChapterToml(newChapter.Path, *newChapterToml)
}

var moveSceneCmd = &cobra.Command{
	Use:   "scene",
	Short: "move a scene's position or folder and position",
	Long:  "move a scene's position or folder and position",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := utilities.IsDraftcatProject()
		if err != nil {
			fmt.Println(err)
			return
		}
		j, err := cmd.Flags().GetBool("json")
		if err != nil {
			fmt.Printf(`{"status":false, "error":"%s"}`, err)
		}
		db, err := utilities.OpenDB()
		if err != nil {
			fmt.Printf(`{"status":false, "error":"%s"}`, err)
		}
		defer func() {
			if closeErr := db.Close(); closeErr != nil {
				fmt.Println("Error closing DB:", closeErr)
			}
		}()
		moveSceneOptions, err := GenerateMoveSceneOptions(db, cmd, args)
		if err != nil {
			fmt.Printf(`{"status":false, "error":"%s"}`, err)
		}
		message := fmt.Sprintf("draftcat|backup|moveScene|%d|%d|%d", moveSceneOptions.SceneID, moveSceneOptions.OldChapterID, moveSceneOptions.NewChapterID)
		err = utilities.CommitBackupSnapshot(message)
		if err != nil {
			fmt.Printf(`{"status":false, "error":"%s"}`, err)
		}
		err = MoveScene(db, moveSceneOptions)
		if err != nil {
			errRestore := utilities.RestoreChanges()
			if errRestore != nil {
				err = fmt.Errorf("%w-%w", err, errRestore)
			}
			if j {
				fmt.Printf(`{"status":false, "error":"%s"}`, err)
			} else {
				fmt.Printf("%s", err)
			}
			return
		}
		fmt.Printf(`{"status":true, "error":""}`)
	},
}

type MoveChapterOptions struct {
	ID          int
	ParentID    sql.NullInt64
	NewParentID sql.NullInt64
}

var moveFolderCmd = &cobra.Command{
	Use:   "folder",
	Short: "move a folder's position or to a different folder and position",
	Long:  "move a folder's position or to a different folder and position",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := utilities.IsDraftcatProject()
		if err != nil {
			fmt.Println(err)
			return
		}
		j, err := cmd.Flags().GetBool("json")
		if err != nil {
			fmt.Printf(`{"status":false, "error":"%s"}`, err)
		}
		if err != nil {
			if j {
				fmt.Printf(`{"status":false, "error":"%s"}`, err)
			} else {
				fmt.Printf("%s", err)
			}
			return
		}
	},
}

func initMoveCmd() {
	storyCmd.AddCommand(moveCmd)

	moveCmd.AddCommand(moveSceneCmd)
	moveCmd.AddCommand(moveFolderCmd)

	moveSceneCmd.Flags().IntP("folderID", "n", 0, "Set the folder to move the scene to, 0 refers to the root folder")
	moveSceneCmd.Flags().IntP("sceneID", "s", 0, "Scene ID to be moved")
	moveSceneCmd.Flags().BoolP("json", "j", false, "Output Json")
	moveFolderCmd.Flags().IntP("currentFolderID", "f", 0, "Current Folder ID to be moved")
	moveFolderCmd.Flags().IntP("folderID", "n", 0, "Set the folder to move the folder to, 0 refers to the root folder")
	moveFolderCmd.Flags().BoolP("json", "j", false, "Output Json")
}
