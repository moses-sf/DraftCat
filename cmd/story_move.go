/*
Package cmd
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package cmd

import (
	"database/sql"
	"fmt"

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

type MoveChapterOptions struct {
	ID       int
	ParentID sql.NullInt64
	Position int
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
