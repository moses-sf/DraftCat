/*
Package cmd
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package cmd

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
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
		ID: chapID,
		ParentID: sql.NullInt64{
			Valid: false,
		},
		PathToRoot: root,
		Position:   newPosition,
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

// storyCmd represents the story command
var storyCmd = &cobra.Command{
	Use:   "story",
	Short: "Root command for story manipulation",
	Long:  `Story editing commands`,
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
			fmt.Println("Chapter Creation Failed", err)
			return
		}
	},
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
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Println("Error getting current directory")
			return
		}
		if !utilities.FolderTomlExists(filepath.Join(cwd, ".chapter.toml")) {
			fmt.Println("Incorrect folder, not created from Draftcat. Please make sure you're in a folder generated from DraftCat.")
			return
		}
		chapter, err := utilities.LoadChapterToml(cwd)
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
		maxPosition, err := databasehandler.GetMaxScenePosition(db, chapter.ID)
		if err != nil {
			fmt.Println("Error in retrieving max position")
		}
		if cmd.Flags().Changed("position") {
			position, err = cmd.Flags().GetInt("position")
			if err != nil {
				fmt.Println("Error in getting position")
				return
			}
		} else {
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
		if maxPosition >= position {
			sceneID, err = databasehandler.InsertSceneAtPosition(db, scene)
		} else {
			sceneID, err = databasehandler.AddScene(db, scene)
		}
		if sceneID == 0 {
			fmt.Println("Scene Id not updated")
			return
		}
		chapter.Scenes = append(chapter.Scenes, utilities.SceneMetaData{
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
		fmt.Println("Moving Scene")
		if !cmd.Flags().Changed("position") || !cmd.Flags().Changed("sceneID") {
			fmt.Println("Position or current Scene ID not defined")
			return
		}
	},
}

var moveFolderCmd = &cobra.Command{
	Use:   "scene",
	Short: "move a folder's position or to a different folder and position",
	Long:  "move a folder's position or to a different folder and position",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := utilities.IsDraftcatProject()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Moving Folder")
		if !cmd.Flags().Changed("position") || !cmd.Flags().Changed("currentFolderID") {
			fmt.Println("Position or current Folder ID not defined")
			return
		}
	},
}

func BuildStoryProject() (*utilities.StoryStructure, error) {
	_, err := utilities.IsDraftcatProject()
	if err != nil {
		return nil, err
	}
	root, err := utilities.GetRelativeRootPath()
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(root, ".story.db")
	db, err := sql.Open("sqlite", dbPath)
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

func init() {
	rootCmd.AddCommand(storyCmd)
	storyCmd.AddCommand(addCmd)
	storyCmd.AddCommand(moveCmd)
	storyCmd.AddCommand(showCmd)
	addCmd.AddCommand(addFolderCmd)
	addCmd.AddCommand(addSceneCmd)
	moveCmd.AddCommand(moveSceneCmd)
	moveCmd.AddCommand(moveFolderCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// storyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// storyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	showCmd.Flags().BoolP("json", "j", false, "Output json to stdout")
	addFolderCmd.Flags().StringP("name", "n", "", "Set the name of the folder")
	addFolderCmd.Flags().IntP("position", "p", 0, "Set the position of the folder in the project")
	addSceneCmd.Flags().StringP("name", "n", "", "Set the name of the scene")
	addSceneCmd.Flags().IntP("position", "p", 0, "Set the position of the scene in the chapter")
	moveSceneCmd.Flags().IntP("folderID", "f", 0, "Set the folder to move the scene to, 0 refers to the root folder")
	moveSceneCmd.Flags().IntP("sceneID", "s", 0, "Scene ID to be moved")
	moveSceneCmd.Flags().IntP("position", "p", 0, "Set the position of the scene in the folder")
	moveFolderCmd.Flags().IntP("currentFolderID", "c", 0, "Current Folder ID to be moved")
	moveFolderCmd.Flags().IntP("newFolderID", "n", 0, "Set the folder to move the folder to, 0 refers to the root folder")
	moveFolderCmd.Flags().IntP("position", "p", 0, "Set the position of the folder in the folder")
}
