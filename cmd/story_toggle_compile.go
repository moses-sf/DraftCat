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
	"path/filepath"

	databasehandler "github.com/moses-sf/DraftCat/databaseHandler"
	"github.com/moses-sf/DraftCat/utilities"
	"github.com/spf13/cobra"
)

func ToggleChapter(db *sql.DB, id int) error {
	return nil
}

func ToggleScene(db *sql.DB, id int) error {
	scene, err := databasehandler.GetScene(db, id)
	if err != nil {
		return err
	}
	chapter, err := databasehandler.GetChapter(db, scene.ChapterID)
	if err != nil {
		return err
	}
	chapterToml, err := utilities.LoadChapterToml(chapter.Path)
	if err != nil {
		return err
	}
	err = databasehandler.UpdateSceneCompile(db, id, !scene.Compile)
	if err != nil {
		return err
	}
	for _, childScene := range chapterToml.Scenes {
		if childScene.ID == id {
			childScene.Compile = !childScene.Compile
			break
		}
	}
	chapterTomlPath := filepath.Join(chapter.Path, ".chapter.toml")
	return utilities.EncodeToml(chapterTomlPath, chapterToml)
}

func ToggleCompile(toggleOptions ItemKindOptions) error {
	dbPath, err := utilities.GetDBPath()
	if err != nil {
		return err
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		fmt.Println("Could not access story DB, please run drafcat story repair")
		return err
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Println("Error closing DB:", closeErr)
		}
	}()

	switch toggleOptions.ItemType {
	case Chapter:

		return ToggleChapter(db, toggleOptions.ID)
	case Scene:
		return ToggleScene(db, toggleOptions.ID)
	}
	return errors.New("did not execute")
}

var toggleCompileCmd = &cobra.Command{
	Use:   "toggle-compile",
	Short: "Toggle Compile Setting",
	Long:  "Toggle Compile Setting for chapters and scenes",
	Run: func(cmd *cobra.Command, args []string) {
		j, err := cmd.Flags().GetBool("json")
		if err != nil {
			log.Fatalf(`{"status":false, "error":"%s"}`, err)
		}
		toggleOpts, err := GenerateItemKindOptions(cmd)
		if err != nil {
			if j {
				log.Fatalf(`{"status":false, "error":"%s"}`, err)
			} else {
				log.Fatal(err)
			}
			return
		}
		err = ToggleCompile(toggleOpts)
		if err != nil {
			if j {
				log.Fatalf(`{"status":false, "error":"%s"}`, err)
			} else {
				log.Fatal(err)
			}
			return
		}
		if err != nil {
			if j {
				log.Fatalf(`{"status":false, "error":"%s"}`, err)
			} else {
				log.Fatal(err)
			}
			return
		}
		fmt.Printf("%s", `{"status":true, "error":""}`)
	},
}

func initToggleCompileCmd() {
	storyCmd.AddCommand(toggleCompileCmd)

	toggleCompileCmd.Flags().BoolP("folder", "f", false, "Toggle a chapter compile")
	toggleCompileCmd.Flags().BoolP("scene", "s", false, "Toggle a scene compile")
	toggleCompileCmd.Flags().BoolP("json", "j", false, "Output Json")
	toggleCompileCmd.Flags().IntP("id", "i", 0, "Scene or Folder ID")
}
