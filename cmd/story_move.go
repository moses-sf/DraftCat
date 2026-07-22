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
	"slices"

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
	if scene.ChapterID == folderID {
		return MoveSceneOptions{}, errors.New("cannot reroot on the same chapter")
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
	err = databasehandler.UpdateScenePathChapterAppendPosition(db, scene)
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

type MoveFolderOptions struct {
	ID          int
	ParentID    sql.NullInt64
	NewParentID sql.NullInt64
}

func GenerateMoveFolderOptions(db *sql.DB, cmd *cobra.Command, args []string) (MoveFolderOptions, error) {
	id, err := cmd.Flags().GetInt("currentFolderID")
	if err != nil {
		return MoveFolderOptions{}, err
	}
	newFolderID, err := cmd.Flags().GetInt("folderID")
	if err != nil {
		return MoveFolderOptions{}, err
	}
	chapter, err := databasehandler.GetChapter(db, id)
	if err != nil {
		return MoveFolderOptions{}, err
	}
	if id == newFolderID {
		return MoveFolderOptions{}, errors.New("cannot reroot a folder on itself")
	}
	oldParentID := 0
	if chapter.ParentID.Valid {
		oldParentID = int(chapter.ParentID.Int64)
	}
	if oldParentID == newFolderID {
		return MoveFolderOptions{}, errors.New("cannot reroot a folder to the same parent")
	}
	if slices.Contains(chapter.DescendantIDs, newFolderID) {
		return MoveFolderOptions{}, errors.New("cannot reroot a folder to its descendants")
	}
	newParentID := sql.NullInt64{
		Valid: false,
	}
	if newFolderID != 0 {
		newParentID.Valid = true
		newParentID.Int64 = int64(newFolderID)
	}
	return MoveFolderOptions{
		ID:          id,
		ParentID:    chapter.ParentID,
		NewParentID: newParentID,
	}, nil
}

func UpdateChildChapterMove(db *sql.DB, chapter databasehandler.Chapter, root bool) error {
	var newPathToRoot string
	chapters, err := databasehandler.GetChaptersWithParent(db, sql.NullInt64{Valid: true, Int64: int64(chapter.ID)})
	if err != nil {
		return err
	}
	if chapter.ParentID.Valid {
		parentToml, err := utilities.LoadChapterToml(filepath.Join(chapter.Path, ".."))
		if err != nil {
			return fmt.Errorf("%w : %+v", err, chapter)
		}
		newPathToRoot = filepath.Join(parentToml.PathToRoot, "..")
	} else {
		newPathToRoot = ".."
	}
	chapterToml, err := utilities.LoadChapterToml(chapter.Path)
	if err != nil {
		return err
	}
	chapterToml.ParentID = chapter.ParentID
	chapterToml.PathToRoot = newPathToRoot
	chapterToml.Depth = chapter.Depth
	chapterToml.Position = chapter.Position
	err = chapterToml.RebuildScenePaths(db, chapter.Path)
	if err != nil {
		return err
	}
	err = utilities.EncodeChapterToml(chapter.Path, *chapterToml)
	if err != nil {
		return err
	}
	for _, childChapter := range chapters {
		childChapter.Path = filepath.Join(chapter.Path, childChapter.Name)
		childChapter.Depth = chapter.Depth + 1
		err = databasehandler.UpdateChapterPathAndDepth(db, childChapter)
		if err != nil {
			return err
		}
		err = UpdateChildChapterMove(db, childChapter, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func MoveFolder(db *sql.DB, moveFolderOptions MoveFolderOptions) error {
	var newParent databasehandler.Chapter
	var newPath string
	var depth int
	chapter, err := databasehandler.GetChapter(db, moveFolderOptions.ID)
	if err != nil {
		return err
	}
	chapterToml, err := utilities.LoadChapterToml(chapter.Path)
	if err != nil {
		return err
	}
	maxPosition, err := databasehandler.GetMaxChapterPosition(db, moveFolderOptions.NewParentID)
	if err != nil {
		return err
	}
	if moveFolderOptions.NewParentID.Valid {
		newParent, err = databasehandler.GetChapter(db, int(moveFolderOptions.NewParentID.Int64))
		if err != nil {
			return err
		}
		newPath = filepath.Join(newParent.Path, chapter.Name)
		if utilities.FolderExists(newPath) {
			return errors.New("folder with the same name exists in destination, rename or delete the target folder")
		}
		depth = newParent.Depth + 1

	} else {
		newPath = filepath.Join(chapter.Path, chapterToml.PathToRoot, chapter.Name)
		if utilities.FolderExists(newPath) {
			return errors.New("folder with the same name exists in root, rename or delete the target folder")
		}
		depth = 1
	}
	err = os.Rename(chapter.Path, newPath)
	if err != nil {
		return err
	}
	chapter.Position = maxPosition + 1
	chapter.Path = newPath
	chapter.Depth = depth
	chapter.ParentID = moveFolderOptions.NewParentID
	err = databasehandler.UpdateChapterParentPathAppendPosition(db, chapter)
	if err != nil {
		return err
	}
	err = UpdateChildChapterMove(db, chapter, true)
	if err != nil {
		return fmt.Errorf("child update error - %w", err)
	}
	return nil
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
		db, err := utilities.OpenDB()
		if err != nil {
			fmt.Printf(`{"status":false, "error":"%s"}`, err)
		}
		moveFolderOptions, err := GenerateMoveFolderOptions(db, cmd, args)
		if err != nil {
			fmt.Printf(`{"status":false, "error":"%s"}`, err)
		}
		message := fmt.Sprintf("draftcat|backup|moveFolder|%d|%d|%d", moveFolderOptions.ID, moveFolderOptions.ParentID.Int64, moveFolderOptions.NewParentID.Int64)
		err = utilities.CommitBackupSnapshot(message)
		if err != nil {
			fmt.Printf(`{"status":false, "error":"%s"}`, err)
		}
		err = MoveFolder(db, moveFolderOptions)
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
