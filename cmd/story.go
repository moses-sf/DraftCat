/*
Package cmd
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// storyCmd represents the story command
var storyCmd = &cobra.Command{
	Use:   "story",
	Short: "Root command for story manipulation",
	Long:  `Story editing commands`,
}

func init() {
	rootCmd.AddCommand(storyCmd)
	initAddCmd()
	initMoveCmd()
	initShowCmd()
	initRenameCmd()
	initRepositionCmd()
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// storyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// storyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
