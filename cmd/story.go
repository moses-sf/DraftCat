/*
Package cmd
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package cmd

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	databasehandler "github.com/moses-sf/draftcat/databaseHandler"
	"github.com/spf13/cobra"
)

func FindToml(tomlName string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	_, err = os.Stat(filepath.Join(cwd, tomlName))
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Attempting to find local file")
		}
		return "", err
	} else {
		return cwd, nil
	}
}

type SceneMetaData struct {
	ID       int
	Name     string
	Path     string
	Position int
}

type ChapterMetaData struct {
	ID         int
	ParentID   sql.NullInt64
	PathToRoot string
	Position   int
	Scenes     []SceneMetaData
}

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

func CreateChapter(db *sql.DB, path, name, root string, parentID sql.NullInt64) error {
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
	newPosition, err := databasehandler.GetMaxChapterPosition(db, parentID)
	if err != nil {
		return err
	}
	chapID, err := databasehandler.AddChapter(db, databasehandler.ChapterCreate{
		Name:     name,
		ParentID: parentID,
		Path:     path,
		Position: newPosition + 1,
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
	scene := make([]SceneMetaData, 0)
	scene = append(scene, SceneMetaData{
		ID:       sceneID,
		Name:     "scene",
		Path:     path,
		Position: 1,
	})
	chapter := ChapterMetaData{
		ID: chapID,
		ParentID: sql.NullInt64{
			Valid: false,
		},
		PathToRoot: root,
		Position:   newPosition + 1,
		Scenes:     scene,
	}
	buf := new(bytes.Buffer)
	encoder := toml.NewEncoder(buf)
	err = encoder.Encode(chapter)
	if err != nil {
		return err
	}

	err = os.WriteFile(path+"/.chapter.toml", buf.Bytes(), 0o644)
	return err
}

func ChapterExists(db *sql.DB, chapterName string, parentID sql.NullInt64) (bool, error) {
	var chapters []string
	res, err := db.Query(`SELECT name FROM chapters WHERE name=? AND parent_id=?`, chapterName, parentID)
	defer func(r *sql.Rows) {
		err = r.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(res)
	if err != nil {
		fmt.Println("Could not run query")
		return false, err
	}
	for res.Next() {
		var currentString string
		err := res.Scan(&currentString)
		if err != nil {
			continue
		}
		chapters = append(chapters, currentString)
	}
	fmt.Println(chapters)
	if len(chapters) > 0 {
		return true, nil
	}
	return false, nil
}

// storyCmd represents the story command
var storyCmd = &cobra.Command{
	Use:   "story",
	Short: "Root command for story manipulation",
	Long:  `Story editing commands`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("story called")
	},
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add story element",
	Long:  "Add Story elements",
	Run: func(cmd *cobra.Command, args []string) {
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
		cwd, err := FindToml(".story.toml")
		if err != nil {
			cwd, err = FindToml(".chapter.toml")
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
		dbPath := filepath.Join(cwd, ".story.db")
		if parent {
			chapter, err := LoadChapterToml(cwd)
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
		defer func(d *sql.DB) {
			err = db.Close()
			if err != nil {
				fmt.Println(err)
			}
		}(db)
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
		err = CreateChapter(db, filepath.Join(cwd, chapterName), chapterName, root, parentID)
		if err != nil {
			fmt.Println("Chapter Creation Failed", err)
			return
		}
	},
}

func FindFolderToml() (string, error) {
	return "", nil
}

func GetFolderToml(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".chapter.toml"))
	return err == nil
}

func LoadChapterToml(path string) (*ChapterMetaData, error) {
	chapter := &ChapterMetaData{}
	file, err := os.ReadFile(filepath.Join(path, ".chapter.toml"))
	if err != nil {
		return nil, err
	}
	_, err = toml.Decode(string(file), chapter)
	if err != nil {
		return nil, err
	}
	return chapter, nil
}

var addSceneCmd = &cobra.Command{
	Use:   "scene",
	Short: "Add a scene.md",
	Long:  "Add a basic scene markdown file",
	Run: func(cmd *cobra.Command, args []string) {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Println("Error getting current directory")
			return
		}
		if !GetFolderToml(cwd) {
			fmt.Println("Incorrect folder, not created from Draftcat. Please make sure you're in a folder generated from DraftCat.")
			return
		}
		chapter, err := LoadChapterToml(cwd)
		if err != nil {
			fmt.Println("Error in loading metadata")
			return
		}
		db, err := sql.Open("sqlite", filepath.Join(cwd, chapter.PathToRoot, ".story.db"))
		if err != nil {
			fmt.Println("Could not access story DB, please run drafcat story repair")
			return
		}
		defer func(d *sql.DB) {
			err = db.Close()
			if err != nil {
				fmt.Println(err)
			}
		}(db)
		sceneName := "scene"
		position := 0
		if cmd.Flags().Changed("name") {
			sceneName, err = cmd.Flags().GetString("name")
			if err != nil {
				fmt.Println("Error in setting Scene name")
				return
			}
		}
		maxPosition := 0
		if cmd.Flags().Changed("position") {
			position, err = cmd.Flags().GetInt("position")
			if err != nil {
				fmt.Println("Error in getting position")
				return
			}
		} else {
			maxPosition, err = databasehandler.GetMaxScenePosition(db, chapter.ID)
			if err != nil {
				fmt.Println("Error in retrieving max position")
			}
			position = maxPosition + 1
		}

		scenePath := filepath.Join(cwd, sceneName+".md")
		err = CreateScene(scenePath)
		if err != nil {
			fmt.Println("Scene Creation Failed")
			return
		}
		scene := databasehandler.SceneCreate{
			ChapterID: chapter.ID,
			Name:      sceneName,
			Path:      scenePath,
			Position:  position,
		}

		sceneID := 0
		if maxPosition > position {
			sceneID, err = databasehandler.InsertSceneAtPosition(db, scene)
		} else {
			sceneID, err = databasehandler.AddScene(db, scene)
		}
		if sceneID == 0 {
			fmt.Println("Scene Id not updated")
			return
		}
		chapter.Scenes = append(chapter.Scenes, SceneMetaData{
			ID:       sceneID,
			Name:     sceneName,
			Path:     scenePath,
			Position: position,
		})

		buf := new(bytes.Buffer)
		encoder := toml.NewEncoder(buf)
		err = encoder.Encode(chapter)
		if err != nil {
			fmt.Println("Error encoding chapter data")
			return
		}

		err = os.WriteFile(cwd+"/.chapter.toml", buf.Bytes(), 0o644)
		if err != nil {
			fmt.Println("Error writing metadata")
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(storyCmd)
	storyCmd.AddCommand(addCmd)
	addCmd.AddCommand(addFolderCmd)
	addCmd.AddCommand(addSceneCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// storyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// storyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	addFolderCmd.Flags().StringP("name", "n", "", "Set the name of the folder")
	addFolderCmd.Flags().IntP("position", "p", 0, "Set the position of the folder in the project")
	addSceneCmd.Flags().StringP("name", "n", "", "Set the name of the scene")
	addSceneCmd.Flags().IntP("position", "p", 0, "Set the position of the scene in the chapter")
}
