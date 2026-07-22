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
	"os"
	"path/filepath"

	databasehandler "github.com/moses-sf/DraftCat/databaseHandler"
	"github.com/moses-sf/DraftCat/utilities"
	"github.com/spf13/cobra"
)

var renameCmd = &cobra.Command{
	Use:   "rename",
	Short: "Rename folders and files",
	Long:  "Rename folders and files",
}

type FolderRenameOptions struct {
	Name     string
	FolderID int
}

type RenameOutcome struct {
	Type    string `json:"type"`
	ID      int    `json:"id"`
	NewPath string `json:"new_path"`
}

func ChapterUpdateFromRoot(db *sql.DB, rootPath string, id sql.NullInt64) ([]*RenameOutcome, error) {
	outcomes := make([]*RenameOutcome, 0)
	chapters, err := databasehandler.GetChaptersWithParent(db, id)
	for _, chapter := range chapters {
		newFolderPath := filepath.Join(rootPath, chapter.Name)
		chapterToml, err := utilities.LoadChapterToml(newFolderPath)
		if err != nil {
			return nil, err
		}
		tomlPath := filepath.Join(newFolderPath, ".chapter.toml")
		chapter.Path = newFolderPath
		err = databasehandler.UpdateChapterPath(db, chapter)
		if err != nil {
			return nil, err
		}
		outcomes = append(outcomes, &RenameOutcome{
			Type:    "chapter",
			ID:      chapter.ID,
			NewPath: newFolderPath,
		})
		for _, scene := range chapterToml.Scenes {
			scene.Path = filepath.Join(newFolderPath, fmt.Sprintf("%s%s", scene.Name, ".md"))
			s := databasehandler.Scene{
				ID:   scene.ID,
				Path: scene.Path,
			}
			err = databasehandler.UpdateScenePath(db, s)
			if err != nil {
				return nil, err
			}
			outcomes = append(outcomes, &RenameOutcome{
				Type:    "scene",
				ID:      scene.ID,
				NewPath: scene.Path,
			})
		}
		err = utilities.EncodeToml(tomlPath, chapterToml)
		if err != nil {
			return nil, err
		}
		treeOutcomes, err := ChapterUpdateFromRoot(db, newFolderPath, sql.NullInt64{Valid: true, Int64: int64(chapter.ID)})
		if err != nil {
			return nil, err
		}
		outcomes = append(outcomes, treeOutcomes...)
	}
	if err != nil {
		return nil, err
	}
	return outcomes, nil
}

func RenameFolder(folderOpts FolderRenameOptions) ([]*RenameOutcome, error) {
	outcomes := make([]*RenameOutcome, 0)
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
	chapter, err := databasehandler.GetChapter(db, folderOpts.FolderID)
	if err != nil {
		return nil, err
	}
	if chapter.Name == folderOpts.Name {
		return nil, fmt.Errorf("same name, aborting")
	}
	chapterToml, err := utilities.LoadChapterToml(chapter.Path)
	if err != nil {
		return nil, err
	}
	upfolderPath := filepath.Join(chapter.Path, "..")
	newFolderPath := filepath.Join(upfolderPath, folderOpts.Name)
	newFolderExists := utilities.FolderExists(newFolderPath)
	if newFolderExists {
		return nil, fmt.Errorf("folder exists use another name %s", newFolderPath)
	}
	err = os.Rename(chapter.Path, newFolderPath)
	if err != nil {
		return nil, err
	}
	chapterToml.Name = folderOpts.Name
	chapter.Name = folderOpts.Name
	chapter.Path = newFolderPath
	err = databasehandler.UpdateChapterPathAndName(db, chapter)
	if err != nil {
		return nil, err
	}
	outcomes = append(outcomes, &RenameOutcome{
		Type:    "chapter",
		ID:      chapter.ID,
		NewPath: newFolderPath,
	})
	for _, scene := range chapterToml.Scenes {
		scene.Path = filepath.Join(newFolderPath, fmt.Sprintf("%s%s", scene.Name, ".md"))
		s := databasehandler.Scene{
			ID:   scene.ID,
			Path: scene.Path,
		}
		err = databasehandler.UpdateScenePath(db, s)
		if err != nil {
			return nil, err
		}
		outcomes = append(outcomes, &RenameOutcome{
			Type:    "scene",
			ID:      scene.ID,
			NewPath: scene.Path,
		})
	}
	tomlPath := filepath.Join(newFolderPath, ".chapter.toml")
	err = utilities.EncodeToml(tomlPath, chapterToml)
	if err != nil {
		return nil, err
	}
	treeOutcomes, err := ChapterUpdateFromRoot(db, newFolderPath, sql.NullInt64{Valid: true, Int64: int64(chapter.ID)})
	if err != nil {
		return nil, err
	}
	outcomes = append(outcomes, treeOutcomes...)
	return outcomes, nil
}

func renameFolderCommand(cmd *cobra.Command) (FolderRenameOptions, error) {
	_, err := utilities.IsDraftcatProject()
	if err != nil {
		return FolderRenameOptions{}, err
	}
	if !cmd.Flags().Changed("name") || !cmd.Flags().Changed("folderID") {
		return FolderRenameOptions{}, fmt.Errorf("both name and id required")
	}
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return FolderRenameOptions{}, fmt.Errorf("error retrieving name")
	}
	id, err := cmd.Flags().GetInt("folderID")
	if err != nil {
		return FolderRenameOptions{}, fmt.Errorf("error retrieving Folder ID")
	}
	return FolderRenameOptions{
		Name:     name,
		FolderID: id,
	}, nil
}

type FolderResponse struct {
	Status bool             `json:"status"`
	Error  error            `json:"error"`
	Path   []*RenameOutcome `json:"path"`
}

var renameFolderCmd = &cobra.Command{
	Use:   "folder",
	Short: "Rename Folder",
	Long:  "Rename Folder",
	Run: func(cmd *cobra.Command, args []string) {
		jsonValue, err := cmd.Flags().GetBool("json")
		if err != nil {
			fmt.Printf(`{"status":false, "error":"%s"}`, err)
		}
		folderOptions, err := renameFolderCommand(cmd)
		if err != nil {
			fmt.Printf(`{"status":false, "error":"%s"}`, err)
		}
		message := fmt.Sprintf("draftcat|backup|renameFolder|%d", folderOptions.FolderID)
		err = utilities.CommitBackupSnapshot(message)
		if err != nil {
			fmt.Printf(`{"status":false, "error":"%s"}`, err)
		}
		path, err := RenameFolder(folderOptions)
		response := FolderResponse{}
		if err != nil {
			errRestore := utilities.RestoreChanges()
			response.Status = false
			response.Error = fmt.Errorf("%w-%w", err, errRestore)
			if jsonValue {
				j, err := json.Marshal(response)
				if err != nil {
					log.Fatalf(`{"status":false, "error":"%s"}`, err)
				}
				log.Fatal(string(j))
				return
			} else {
				log.Fatal(err)
			}
			return
		}
		response.Status = true
		response.Path = path
		j, err := json.Marshal(response)
		if err != nil {
			log.Fatalf(`{"status":false, "error":"%s"}`, err)
		}
		fmt.Println(string(j))
	},
}

type SceneRenameOptions struct {
	Name    string
	SceneID int
}

func RenameScene(sceneOpts SceneRenameOptions) (string, error) {
	db, err := utilities.OpenDB()
	if err != nil {
		fmt.Println("Could not access story DB, please run drafcat story repair")
		return "", err
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Println("Error closing DB:", closeErr)
		}
	}()
	scene, err := databasehandler.GetScene(db, sceneOpts.SceneID)
	if err != nil {
		return "", err
	}
	if scene.Name == sceneOpts.Name {
		return "", fmt.Errorf("same name, aborting")
	}
	chapterPath := filepath.Dir(scene.Path)
	chapterTomlPath := filepath.Join(chapterPath, ".chapter.toml")
	chapterToml, err := utilities.LoadChapterToml(chapterPath)
	if err != nil {
		return "", err
	}
	newPath := filepath.Join(chapterPath, fmt.Sprintf("%s%s", sceneOpts.Name, ".md"))
	if utilities.FileExists(newPath) {
		return "", fmt.Errorf("file exists %s", newPath)
	}
	err = os.Rename(scene.Path, newPath)
	if err != nil {
		return "", err
	}
	for _, scene := range chapterToml.Scenes {
		if scene.ID == sceneOpts.SceneID {
			scene.Name = sceneOpts.Name
			scene.Path = newPath
			err = databasehandler.UpdateScenePathAndName(db, databasehandler.Scene{
				ID:   sceneOpts.SceneID,
				Name: scene.Name,
				Path: newPath,
			})
			if err != nil {
				return "", err
			}
			break
		}
	}

	err = utilities.EncodeToml(chapterTomlPath, chapterToml)
	if err != nil {
		return "", err
	}
	return newPath, nil
}

func GenerateSceneRenameOptions(cmd *cobra.Command) (SceneRenameOptions, error) {
	_, err := utilities.IsDraftcatProject()
	if err != nil {
		return SceneRenameOptions{}, err
	}
	if !cmd.Flags().Changed("name") || !cmd.Flags().Changed("sceneID") {
		return SceneRenameOptions{}, fmt.Errorf("name and sceneID required")
	}
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return SceneRenameOptions{}, err
	}
	id, err := cmd.Flags().GetInt("sceneID")
	if err != nil {
		return SceneRenameOptions{}, err
	}
	return SceneRenameOptions{
		Name:    name,
		SceneID: id,
	}, nil
}

type JSONStatus struct {
	Status bool   `json:"status"`
	Error  string `json:"error"`
}

var renameSceneCmd = &cobra.Command{
	Use:   "scene",
	Short: "Rename Scene",
	Long:  "Rename Scene",
	Run: func(cmd *cobra.Command, args []string) {
		j, err := cmd.Flags().GetBool("json")
		if err != nil {
			log.Fatalf(`{"status":false, "error":"%s"}`, err)
		}
		sceneOpts, err := GenerateSceneRenameOptions(cmd)
		if err != nil {
			log.Fatalf(`{"status":false, "error":"%s"}`, err)
		}
		message := fmt.Sprintf("draftcat|backup|renameScene|%d", sceneOpts.SceneID)
		err = utilities.CommitBackupSnapshot(message)
		if err != nil {
			log.Fatalf(`{"status":false, "error":"%s"}`, err)
		}
		path, err := RenameScene(sceneOpts)
		if err != nil {
			errRestore := utilities.RestoreChanges()
			if errRestore != nil {
				err = fmt.Errorf("%w-%w", err, errRestore)
			}
			if err != nil {
				log.Fatalf(`{"status":false, "error":"FATAL %s"}`, err)
			}
			if j {
				log.Fatalf(`{"status":false, "error":"%s"}`, err)
			} else {
				log.Fatal(err)
			}
			return
		}
		fmt.Printf(`{"status":true, "error":"", "path":"%s"}`, path)
	},
}

func initRenameCmd() {
	storyCmd.AddCommand(renameCmd)
	renameCmd.AddCommand(renameFolderCmd)
	renameCmd.AddCommand(renameSceneCmd)
	renameFolderCmd.Flags().IntP("folderID", "f", 0, "Folder ID to be changed")
	renameFolderCmd.Flags().StringP("name", "n", "", "New name of the folder")
	renameFolderCmd.Flags().BoolP("json", "j", false, "Json output")
	renameSceneCmd.Flags().IntP("sceneID", "s", 0, "Scene ID to be changed")
	renameSceneCmd.Flags().StringP("name", "n", "", "New name of the folder")
	renameSceneCmd.Flags().BoolP("json", "j", false, "Json output")
}
