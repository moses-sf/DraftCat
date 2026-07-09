/*
Package cmd
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/moses-sf/DraftCat/utilities"
	"github.com/spf13/cobra"
)

func ConfigDirCreation(configDirPath string) error {
	// Create Config Directory
	exists, err := os.Stat(configDirPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(configDirPath, 0o755)
		}
		if err != nil {
			return err
		}
	} else if !exists.IsDir() {
		return fmt.Errorf("error something with this name not a directory exists: %s", configDirPath)
	}
	return nil
}

func NewConfigFromCommand(cmd *cobra.Command) (*utilities.AuthorConfig, error) {
	// Generate Config Toml
	authorConfig := &utilities.AuthorConfig{}
	if cmd.Flag("name").Changed {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return nil, err
		} else {
			authorConfig.Name = name
		}
	}

	if cmd.Flag("email").Changed {
		email, err := cmd.Flags().GetString("email")
		if err != nil {
			return nil, err
		} else {
			authorConfig.Email = email
		}
	}

	if cmd.Flag("address").Changed {
		address, err := cmd.Flags().GetString("address")
		if err != nil {
			return nil, err
		} else {
			authorConfig.Address = address
		}
	}

	if cmd.Flag("phone").Changed {
		phone, err := cmd.Flags().GetString("phone")
		if err != nil {
			return nil, err
		} else {
			authorConfig.Phone = phone
		}
	}
	return authorConfig, nil
}

type Paths struct {
	dirPath string
	file    string
}

func ConfigPaths() (*Paths, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	configDirPath := filepath.Join(configDir, "draftcat")
	configPath := filepath.Join(configDirPath, "draftcat-config.toml")
	path := &Paths{
		dirPath: configDirPath,
		file:    configPath,
	}
	return path, nil
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Set global config settings.",
	Long:  `Set global config settings.`,
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set Draftcat config variables",
	Long:  "Set the Draftcat config variables",
	Run: func(cmd *cobra.Command, args []string) {
		path, err := ConfigPaths()
		if err != nil {
			fmt.Println(err)
		}
		err = ConfigDirCreation(path.dirPath)
		if err != nil {
			fmt.Println(err)
			return
		}

		authorConfig, err := NewConfigFromCommand(cmd)
		if err != nil {
			fmt.Println("Error in generating config")
			return
		}
		authorConfig.UpdateFromOldConfig(path.file)

		config := utilities.Config{
			Author: *authorConfig,
		}

		config.WriteConfig(path.file)
	},
}

var configUnsetCmd = &cobra.Command{
	Use:   "unset",
	Short: "Unset config Variables",
	Long:  "Unset config variables",
	Run: func(cmd *cobra.Command, args []string) {
		path, err := ConfigPaths()
		if err != nil {
			fmt.Println(err)
			return
		}
		err = ConfigDirCreation(path.dirPath)
		if err != nil {
			fmt.Println(err)
			return
		}

		authorConfig := &utilities.AuthorConfig{}
		authorConfig.UpdateFromOldConfig(path.file)
		authorConfig.Unset(cmd, args)

		config := utilities.Config{
			Author: *authorConfig,
		}
		config.WriteConfig(path.file)
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show Config",
	Long:  "Show Config",
	Run: func(cmd *cobra.Command, args []string) {
		path, err := ConfigPaths()
		if err != nil {
			fmt.Println(err)
			return
		}
		err = ConfigDirCreation(path.dirPath)
		if err != nil {
			fmt.Println(err)
			return
		}
		authorConfig := &utilities.AuthorConfig{}
		authorConfig.UpdateFromOldConfig(path.file)
		fmt.Println(authorConfig)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configUnsetCmd)
	configCmd.AddCommand(configShowCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	configSetCmd.Flags().StringP("name", "n", "", "Set author name")
	configSetCmd.Flags().StringP("email", "e", "", "Set email")
	configSetCmd.Flags().StringP("address", "a", "", "Set address")
	configSetCmd.Flags().StringP("phone", "p", "", "Set Phone number")
}
