/*
Package cmd
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package cmd

import (
	"fmt"
	"log"

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
			log.Fatal(err)
		}

		if !cmd.Flags().Changed("name") {
			log.Fatal("Name required")
		}
		j, err := cmd.Flags().GetBool("json")
		if err != nil {
			log.Fatalf(`{"status":false, "error":"%s"}`, err)
		}
		path, err := utilities.GetRelativeRootPath()
		if err != nil {
			if j {
				log.Fatalf(`{"status":false, "error":"%s"}`, err)
			} else {
				log.Fatal(err)
			}
			return
		}
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			if j {
				log.Fatalf(`{"status":false, "error":"%s"}`, err)
			} else {
				log.Fatal(err)
			}
			return
		}
		message := "draftcat|snapshot|" + name
		err = utilities.CommitWithMessage(path, message)
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

var lastSnapshotCmd = &cobra.Command{
	Use:   "last",
	Short: "Get last Snapshot",
	Long:  "Get's the last snapshot",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := utilities.IsDraftcatProject()
		if err != nil {
			log.Fatal(err)
		}
		path, err := utilities.GetRelativeRootPath()
		if err != nil {
			log.Fatal(err)
		}
		lashHash, err := utilities.RetrieveLastHash(path)
		if err != nil {
			log.Fatal(err)
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
