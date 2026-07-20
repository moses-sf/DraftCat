/*
Package cmd
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	databasehandler "github.com/moses-sf/DraftCat/databaseHandler"
	"github.com/moses-sf/DraftCat/utilities"
	"github.com/spf13/cobra"
)

func CreateScene(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return fmt.Errorf("file exists")
	}
	b := new([]byte)
	err = os.WriteFile(path, *b, 0o644)
	if err != nil {
		return err
	}
	return nil
}

func CreateChapter(db *sql.DB, path, name, root string, parentID sql.NullInt64, position int) error {
	info, err := os.Stat(path)

	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("path already exists and is not a folder: %s", path)
		}
		return fmt.Errorf("folder already exists: %s", path)
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("check folder %s: %w", path, err)
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("create folder %s: %w", path, err)
	}
	newPosition := position
	maxPosition, err := databasehandler.GetMaxChapterPosition(db, parentID)
	if err != nil {
		return err
	}
	if newPosition == 0 || newPosition > maxPosition {
		newPosition = maxPosition + 1
	}
	chapID, err := databasehandler.InsertChapterAtPosition(db, databasehandler.ChapterCreate{
		Name:     name,
		ParentID: parentID,
		Path:     path,
		Position: newPosition,
	})
	if err != nil {
		return err
	}
	scenePath := filepath.Join(path, "scene.md")
	err = CreateScene(scenePath)
	if err != nil {
		return err
	}
	sceneID, err := databasehandler.AddScene(db, databasehandler.SceneCreate{
		ChapterID: chapID,
		Name:      "scene",
		Path:      scenePath,
		Position:  1,
	})
	if err != nil {
		return err
	}
	scene := make([]utilities.SceneMetaData, 0)
	scene = append(scene, utilities.SceneMetaData{
		ID:       sceneID,
		Name:     "scene",
		Path:     scenePath,
		Position: 1,
	})
	chapter := utilities.ChapterMetaData{
		ID:         chapID,
		Name:       name,
		ParentID:   parentID,
		PathToRoot: root,
		Position:   newPosition,
		Scenes:     scene,
	}
	tomlPath := filepath.Join(path, ".chapter.toml")
	return utilities.EncodeToml(tomlPath, chapter)
}

func ChapterExists(db *sql.DB, chapterName string, parentID sql.NullInt64) (bool, error) {
	var count int
	var err error

	if parentID.Valid {
		err = db.QueryRow(`
			SELECT COUNT(*)
			FROM chapters
			WHERE name = ?
			  AND parent_id = ?
		`, chapterName, parentID.Int64).Scan(&count)
	} else {
		err = db.QueryRow(`
			SELECT COUNT(*)
			FROM chapters
			WHERE name = ?
			  AND parent_id IS NULL
		`, chapterName).Scan(&count)
	}

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add story element",
	Long:  "Add Story elements",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := utilities.IsDraftcatProject()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("use add [element] to add an element in your current folder")
	},
}

var addFolderCmd = &cobra.Command{
	Use:   "folder",
	Short: "Add a folder",
	Long: `Add a folder. 
	-b for initialising without basic.md file`,
	Run: func(cmd *cobra.Command, args []string) {
		var chapterName string
		parent := false
		root := ".."
		position := 0
		_, err := utilities.IsDraftcatProject()
		if err != nil {
			fmt.Println(err)
			return
		}
		cwd, err := utilities.FindToml(".story.toml")
		if err != nil {
			cwd, err = utilities.FindToml(".chapter.toml")
			if err != nil {
				fmt.Println("Error retrieving root directory. Does .story.toml or .chapter.toml exist in this folder?")
			}
			parent = true
		}

		if cmd.Flags().Changed("name") {
			chapterName, err = cmd.Flags().GetString("name")
			if err != nil {
				fmt.Println("Couldn't register name")
				return
			}
		} else {
			fmt.Print("No name provided, creating folder called Chapter 1\n")
			chapterName = "Chapter 1"
		}
		parentID := sql.NullInt64{
			Valid: false,
		}

		if cmd.Flags().Changed("position") {
			pos, err := cmd.Flags().GetInt("position")
			if err != nil {
				fmt.Println("Can't register position")
				return
			}
			position = pos
		}
		j, err := cmd.Flags().GetBool("json")
		if err != nil {
			fmt.Println("Failed to load json")
			return
		}
		dbPath := filepath.Join(cwd, ".story.db")
		if parent {
			chapter, err := utilities.LoadChapterToml(cwd)
			if err != nil {
				fmt.Println("Couldn't load folder data")
				return
			}
			root = "../" + chapter.PathToRoot
			parentID = sql.NullInt64{
				Valid: true,
				Int64: int64(chapter.ID),
			}
			dbPath = filepath.Join(cwd, chapter.PathToRoot, ".story.db")
		}
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			fmt.Println("Could not access story DB, please run drafcat story repair")
			return
		}
		defer func() {
			if closeErr := db.Close(); closeErr != nil {
				log.Println("Error closing DB:", closeErr)
			}
		}()
		nameExists, err := ChapterExists(db, chapterName, parentID)
		if err != nil {
			fmt.Println("error retrieving data")
			return
		}
		for nameExists {
			chapterName = fmt.Sprintf("%s copy", chapterName)
			nameExists, err = ChapterExists(db, chapterName, parentID)
			if err != nil {
				fmt.Println("error retrieving data")
				return
			}
		}
		err = CreateChapter(db, filepath.Join(cwd, chapterName), chapterName, root, parentID, position)
		if err != nil {
			if j {
				fmt.Printf(`{"status":false, "error":"%s"}`, err)
				return
			}
			fmt.Println("Chapter Creation Failed", err)
			return
		}
		if j {
			fmt.Printf(`{"status":true, "error":""}`)
		}
	},
}

type AddSceneOptions struct {
	Name        string
	Position    int
	ChapterID   sql.NullInt64
	MaxPosition int
}

func GenerateAddSceneOptions(db *sql.DB, cmd *cobra.Command, args []string) (AddSceneOptions, error) {
	sceneName := "scene"
	position := 0
	folderID, err := cmd.Flags().GetInt("folderID")
	if err != nil {
		return AddSceneOptions{}, err
	}
	if cmd.Flags().Changed("name") {
		sceneName, err = cmd.Flags().GetString("name")
		if err != nil {
			return AddSceneOptions{}, err
		}
	}
	maxPosition, err := databasehandler.GetMaxScenePosition(db, folderID)
	if err != nil {
		fmt.Println("Error in retrieving max position")
	}
	if cmd.Flags().Changed("position") {
		position, err = cmd.Flags().GetInt("position")
		if err != nil {
			return AddSceneOptions{}, err
		}
	} else {
		position = maxPosition + 1
	}

	var chapterID sql.NullInt64
	if folderID == 0 {
		chapterID = sql.NullInt64{Valid: false}
	} else {
		chapterID = sql.NullInt64{Valid: true, Int64: int64(folderID)}
	}

	return AddSceneOptions{
		Name:        sceneName,
		Position:    position,
		ChapterID:   chapterID,
		MaxPosition: maxPosition,
	}, nil
}

func AddScene(db *sql.DB, sceneOpts AddSceneOptions) error {
	if !sceneOpts.ChapterID.Valid {
		return errors.New("cannot insert scene at root")
	}

	chapter, err := databasehandler.GetChapter(db, int(sceneOpts.ChapterID.Int64))
	if err != nil {
		return err
	}
	chapterToml, err := utilities.LoadChapterToml(chapter.Path)
	if err != nil {
		return err
	}

	scenePath := filepath.Join(chapter.Path, sceneOpts.Name+".md")
	err = CreateScene(scenePath)
	if err != nil {
		return err
	}
	scene := databasehandler.SceneCreate{
		ChapterID: chapterToml.ID,
		Name:      sceneOpts.Name,
		Path:      scenePath,
		Position:  sceneOpts.Position,
	}
	sceneID := 0
	if sceneOpts.MaxPosition >= sceneOpts.Position {
		sceneID, err = databasehandler.InsertSceneAtPosition(db, scene)
	} else {
		sceneID, err = databasehandler.AddScene(db, scene)
	}
	if err != nil {
		return err
	}
	chapterToml.Scenes = append(chapterToml.Scenes, utilities.SceneMetaData{
		ID:       sceneID,
		Name:     sceneOpts.Name,
		Path:     scenePath,
		Position: sceneOpts.Position,
	})

	return utilities.EncodeToml(filepath.Join(chapter.Path, ".chapter.toml"), chapterToml)
}

var addSceneCmd = &cobra.Command{
	Use:   "scene",
	Short: "Add a scene.md",
	Long:  "Add a basic scene markdown file",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := utilities.IsDraftcatProject()
		if err != nil {
			fmt.Println(err)
			return
		}
		j, err := cmd.Flags().GetBool("json")
		if err != nil {
			fmt.Println("Failed to load json")
			return
		}
		db, err := utilities.OpenDB()
		if err != nil {
			fmt.Println(err)
			return
		}
		defer func() {
			if closeErr := db.Close(); closeErr != nil {
				fmt.Println("Error closing DB:", closeErr)
			}
		}()
		sceneOpts, err := GenerateAddSceneOptions(db, cmd, args)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = AddScene(db, sceneOpts)
		if err != nil {
			if j {
				fmt.Printf(`{"state":false, "error":"%s"}`, err)
				return
			}
			fmt.Println("Chapter Creation Failed", err)
			return
		}
		if j {
			fmt.Printf(`{"status":true, "error":""}`)
		}
	},
}

func initAddCmd() {
	storyCmd.AddCommand(addCmd)

	addCmd.AddCommand(addFolderCmd)
	addCmd.AddCommand(addSceneCmd)

	addFolderCmd.Flags().StringP("name", "n", "", "Set the name of the folder")
	addFolderCmd.Flags().IntP("position", "p", 0, "Set the position of the folder in the project")
	addFolderCmd.Flags().IntP("folderID", "f", 0, "Chapter ID for scene to be inserted in")
	addFolderCmd.Flags().BoolP("json", "j", false, "Json output")
	addSceneCmd.Flags().StringP("name", "n", "", "Set the name of the scene")
	addSceneCmd.Flags().IntP("position", "p", 0, "Set the position of the scene in the chapter")
	addSceneCmd.Flags().IntP("folderID", "f", 0, "Chapter ID for scene to be inserted in")
	addSceneCmd.Flags().BoolP("json", "j", false, "Json output")
}
