/*
Package cmd
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package cmd

import (
	"fmt"

	"github.com/moses-sf/DraftCat/utilities"
	"github.com/spf13/cobra"
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Take a snapshot of the project",
	Long:  "Take a snapshot of the project",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := utilities.IsDraftcatProject()
		if err != nil {
			fmt.Println(err)
		}

		if !cmd.Flags().Changed("name") {
			fmt.Println("Name required")
		}
		j, err := cmd.Flags().GetBool("json")
		if err != nil {
			fmt.Printf(`{"status":false, "error":"%s"}`, err)
		}
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			if j {
				fmt.Printf(`{"status":false, "error":"%s"}`, err)
			} else {
				fmt.Println(err)
			}
			return
		}
		message := "draftcat|snapshot|" + name
		err = utilities.CommitWithMessage(message)
		if err != nil {
			if j {
				fmt.Printf(`{"status":false, "error":"%s"}`, err)
			} else {
				fmt.Println(err)
			}
			return
		}
		fmt.Printf("%s", `{"status":true, "error":""}`)
	},
}

var lastSnapshotCmd = &cobra.Command{
	Use:   "last",
	Short: "Get last Snapshot",
	Long:  "Get's the last snapshot",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := utilities.IsDraftcatProject()
		if err != nil {
			fmt.Println(err)
		}
		path, err := utilities.GetRelativeRootPath()
		if err != nil {
			fmt.Println(err)
		}
		lashHash, err := utilities.RetrieveLastHash(path)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(lashHash)
	},
}

func init() {
	rootCmd.AddCommand(snapshotCmd)
	snapshotCmd.AddCommand(lastSnapshotCmd)

	snapshotCmd.Flags().StringP("name", "n", "", "Name of the snapshot")
	snapshotCmd.Flags().BoolP("json", "j", false, "return json output")
}
