/*
Package cmd
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	databasehandler "github.com/moses-sf/draftcat/databaseHandler"
	"github.com/spf13/cobra"
)

func FindStoryToml() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	_, err = os.Stat(filepath.Join(cwd, ".story.toml"))
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Attempting to find local file")
		}
		return "", err
	} else {
		return cwd, nil
	}
}

func CreateChapter(db *sql.DB, path string, name string) error {
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
	err = databasehandler.AddChapter(db, databasehandler.ChapterCreate{
		Name: name,
		Path: path,
	})
	return err
}

func ChapterExists(db *sql.DB, chapterName string) (bool, error) {
	var chapters []string
	res, err := db.Query(`SELECT name FROM chapters WHERE name=?`, chapterName)
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

var addFolderCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a folder",
	Long: `Add a folder. 
	-b for initialising without basic.md file`,
	Run: func(cmd *cobra.Command, args []string) {
		var chapterName string
		cwd, err := FindStoryToml()
		if err != nil {
			fmt.Println("Error retrieving root directory. Does .story.toml exist in root?")
			return
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
		db, err := sql.Open("sqlite", filepath.Join(cwd, ".story.db"))
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
		nameExists, err := ChapterExists(db, chapterName)
		if err != nil {
			fmt.Println("error retrieving data")
			return
		}
		for nameExists {
			chapterName = fmt.Sprintf("%s copy", chapterName)
			nameExists, err = ChapterExists(db, chapterName)
			if err != nil {
				fmt.Println("error retrieving data")
				return
			}
		}
		err = CreateChapter(db, filepath.Join(cwd, chapterName), chapterName)
		if err != nil {
			fmt.Println("Chapter Creation Failed", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(storyCmd)
	storyCmd.AddCommand(addFolderCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// storyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// storyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	addFolderCmd.Flags().String("name", "", "Set the name of the folder")
}
