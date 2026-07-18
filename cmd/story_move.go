/*
Package cmd
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package cmd

import (
	"database/sql"
	"fmt"
	"os"

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

type MoveChapterOptions struct {
	ID       int
	ParentID sql.NullInt64
	Position int
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
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			return
		}

		chapter, err := utilities.LoadChapterToml(cwd)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Moving Folder", chapter.ParentID)
		if !cmd.Flags().Changed("position") || !cmd.Flags().Changed("currentFolderID") {
			fmt.Println("Position or current Folder ID not defined")
			return
		}
	},
}

func initMoveCmd() {
	storyCmd.AddCommand(moveCmd)

	moveCmd.AddCommand(moveSceneCmd)
	moveCmd.AddCommand(moveFolderCmd)

	moveSceneCmd.Flags().IntP("folderID", "f", 0, "Set the folder to move the scene to, 0 refers to the root folder")
	moveSceneCmd.Flags().IntP("sceneID", "s", 0, "Scene ID to be moved")
	moveSceneCmd.Flags().IntP("position", "p", 0, "Set the position of the scene in the folder")
	moveFolderCmd.Flags().IntP("currentFolderID", "c", 0, "Current Folder ID to be moved")
	moveFolderCmd.Flags().IntP("newFolderID", "n", 0, "Set the folder to move the folder to, 0 refers to the root folder")
	moveFolderCmd.Flags().IntP("position", "p", 0, "Set the position of the folder in the folder")
}
