/*
Package cmd
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package cmd

import (
	"bufio"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/moses-sf/DraftCat/utilities"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

//go:embed db/0001_init.sql
var schemaFS embed.FS

func CreateDirectory(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(path, 0o755)
			if err != nil {
				return err
			}
			return nil
		}
		return err
	} else {
		return fmt.Errorf("directory already exists ... cancelling")
	}
}

func CreateStoryDirectorys(name string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	path := filepath.Join(cwd, name)
	err = CreateDirectory(path)
	if err != nil {
		return "", err
	}
	return path, nil
}

func LoadAuthorData() (*utilities.AuthorConfig, error) {
	paths := ConfigPaths()
	ConfigDirCreation(paths.dirPath)

	authorConfig := &utilities.AuthorConfig{}
	authorConfig.UpdateFromOldConfig(paths.file)
	return authorConfig, nil
}

func UpdateAuthorData(authorConfig *utilities.AuthorConfig) (*utilities.AuthorConfig, error) {
	fmt.Println("Story Metadata")
	reader := bufio.NewReader(os.Stdin)
	authorConfig.Print()

	for {
		fmt.Printf("Do you want to edit these details for this story? (y/N): ")
		option, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		option = strings.ToLower(strings.TrimSpace(option))
		if option == "n" || option == "" {
			return authorConfig, nil
		} else if option == "y" {
			break
		} else {
			fmt.Printf("Invalid entry %s, please enter a valid value.\n", option)
		}
	}
	for {
		fmt.Printf("Author Name - %s (press enter to use existing value): ", authorConfig.Name)
		newValue, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		newValue = strings.TrimSpace(newValue)
		if newValue != "" {
			authorConfig.Name = newValue
		}

		fmt.Printf("Author Email - %s (press enter to use existing value): ", authorConfig.Email)
		newValue, err = reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		newValue = strings.TrimSpace(newValue)
		if newValue != "" {
			authorConfig.Email = newValue
		}

		fmt.Printf("Author Address - %s (press enter to use existing value): ", authorConfig.Address)
		newValue, err = reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		newValue = strings.TrimSpace(newValue)
		if newValue != "" {
			authorConfig.Address = newValue
		}

		fmt.Printf("Author Phone - %s (press enter to use existing value): ", authorConfig.Phone)
		newValue, err = reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		newValue = strings.TrimSpace(newValue)
		if newValue != "" {
			authorConfig.Phone = newValue
		}
		for {
			authorConfig.Print()
			fmt.Print("Are these values Correct? (y/n):")
			response, err := reader.ReadString('\n')
			if err != nil {
				return nil, err
			}
			response = strings.ToLower(strings.TrimSpace(response))
			if response == "y" {
				return authorConfig, nil
			} else if response != "n" {
				fmt.Println("Incorrect value entered")
			} else {
				break
			}
		}
	}
}

func CreateShortStory(name string, defaults bool) (string, error) {
	storyName := ""
	reader := bufio.NewReader(os.Stdin)
	if name == "" {

		fmt.Println("Set Short Story details")
		fmt.Print("Story Name: ")
		story, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		storyName = strings.TrimSpace(story)
		if storyName == "" {
			return "", fmt.Errorf("story name cannot be empty")
		}
	} else {
		storyName = name
	}
	path, err := CreateStoryDirectorys(storyName)
	if err != nil {
		return "", err
	}
	authorConfig, err := LoadAuthorData()
	if err != nil {
		return "", err
	}
	if !defaults {
		authorConfig, err = UpdateAuthorData(authorConfig)
		if err != nil {
			return "", nil
		}
	}
	storyTomlPath := filepath.Join(path, ".story.toml")
	storyConfig := &utilities.StoryConfig{
		Author: *authorConfig,
		MetaData: utilities.StoryMetaData{
			Name: storyName,
			Type: utilities.Short,
		},
	}
	storyConfig.WriteConfig(storyTomlPath)
	db, err := sql.Open("sqlite", filepath.Join(path, ".story.db"))
	if err != nil {
		return path, err
	}
	defer func(DB *sql.DB) {
		err = db.Close()
		if err != nil {
			log.Fatal("Error Closing DB")
		}
	}(db)
	schema, err := schemaFS.ReadFile("db/0001_init.sql")
	if err != nil {
		return "", nil
	}
	_, err = db.Exec(string(schema))
	if err != nil {
		return "", err
	}
	return path, nil
}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a new story in the current folder",
	Long: `Generate a new story on the basis of your information in the current folder
	Use the following template types: short, novella, novel`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		workType := args[0]
		switch workType {
		case "short":
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				fmt.Println("Error in retrieving name")
				return
			}
			defaults, err := cmd.Flags().GetBool("defaults")
			if err != nil {
				fmt.Println(err)
				return
			}
			path, err := CreateShortStory(name, defaults)
			if err != nil {
				fmt.Printf("Short Story creation failed: %s\n", err)
				return
			}
			vim, err := cmd.Flags().GetBool("vim")
			if err != nil {
				fmt.Println(err)
				return
			}
			if vim {
				ex := exec.Command("nvim", path)
				ex.Stdin = os.Stdin
				ex.Stdout = os.Stdout
				ex.Stderr = os.Stderr

				if err := ex.Run(); err != nil {
					log.Println("nvim exited with error:", err)
				}
			}
		default:
			log.Fatal("Incorrect argument, refer to help for valid types")
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	initCmd.Flags().StringP("name", "n", "", "Title of the Work")
	initCmd.Flags().BoolP("vim", "v", false, "Start neovim")
	initCmd.Flags().BoolP("defaults", "d", false, "Use global author default config")
}
