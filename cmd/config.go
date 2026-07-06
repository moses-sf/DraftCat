/*
Package cmd
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package cmd

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

type AuthorConfig struct {
	Name    string
	Email   string
	Address string
	Phone   string
}

func (a *AuthorConfig) Print() {
	fmt.Println("Author Details -")
	fmt.Printf("Name: %s\n", a.Name)
	fmt.Printf("Email: %s\n", a.Email)
	fmt.Printf("Address: %s\n", a.Address)
	fmt.Printf("Phone Number: %s\n\n\n", a.Phone)
}

func (a *AuthorConfig) FillEmpty(oldconfig AuthorConfig) {
	if a.Name == "" {
		a.Name = oldconfig.Name
	}
	if a.Email == "" {
		a.Email = oldconfig.Email
	}
	if a.Address == "" {
		a.Address = oldconfig.Address
	}
	if a.Phone == "" {
		a.Phone = oldconfig.Phone
	}
}

func (a *AuthorConfig) UpdateFromOldConfig(configPath string) {
	_, err := os.Stat(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatalf("Error in file %s", err)
		}
	} else {
		oldConfigData, err := os.ReadFile(configPath)
		if err != nil {
			log.Printf("Error reading file %s", err)
		} else {
			oldConfig := &Config{}
			_, err = toml.Decode(string(oldConfigData), oldConfig)
			if err != nil {
				log.Printf("Error decoding config %s", err)
			} else {
				a.FillEmpty(oldConfig.Author)
			}
		}
	}
}

func (a *AuthorConfig) Unset(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		if slices.Contains(args, "name") {
			a.Name = ""
		}
		if slices.Contains(args, "address") {
			a.Address = ""
		}
		if slices.Contains(args, "email") {
			a.Email = ""
		}
		if slices.Contains(args, "phone") {
			a.Phone = ""
		}
	} else {
		fmt.Println("Select the variables to unset")
	}
}

type Config struct {
	Author AuthorConfig
}

func (c *Config) WriteConfig(path string) {
	buf := new(bytes.Buffer)
	encoder := toml.NewEncoder(buf)
	err := encoder.Encode(c)
	if err != nil {
		log.Fatalf("Error in creating config %s", err)
	}
	err = os.WriteFile(path, buf.Bytes(), 0o644)
	if err != nil {
		log.Fatalf("Error in creating config file %s", err)
	}
}

func ConfigDirCreation(configDirPath string) {
	// Create Config Directory
	exists, err := os.Stat(configDirPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(configDirPath, 0o755)
		}
		if err != nil {
			log.Fatalf("Error creating Directory %s", err)
		}
	} else if !exists.IsDir() {
		log.Fatalf("Error something with this name not a directory exists: %s", configDirPath)
	}
}

func NewConfigFromCommand(cmd *cobra.Command) AuthorConfig {
	// Generate Config Toml
	authorConfig := AuthorConfig{}
	if cmd.Flag("name").Changed {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			log.Println("error in setting name")
		} else {
			authorConfig.Name = name
		}
	}

	if cmd.Flag("email").Changed {
		email, err := cmd.Flags().GetString("email")
		if err != nil {
			log.Println("error in setting email")
		} else {
			authorConfig.Email = email
		}
	}

	if cmd.Flag("address").Changed {
		address, err := cmd.Flags().GetString("address")
		if err != nil {
			log.Println("error in setting email")
		} else {
			authorConfig.Address = address
		}
	}

	if cmd.Flag("phone").Changed {
		phone, err := cmd.Flags().GetString("phone")
		if err != nil {
			log.Println("error in setting email")
		} else {
			authorConfig.Phone = phone
		}
	}
	return authorConfig
}

type Paths struct {
	dirPath string
	file    string
}

func ConfigPaths() Paths {
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal("Error getting config directory")
	}
	configDirPath := filepath.Join(configDir, "draftcat")
	configPath := filepath.Join(configDirPath, "draftcat-config.toml")
	path := Paths{
		dirPath: configDirPath,
		file:    configPath,
	}
	return path
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Set global config settings.",
	Long:  `Set global config settings.`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set Draftcat config variables",
	Long:  "Set the Draftcat config variables",
	Run: func(cmd *cobra.Command, args []string) {
		path := ConfigPaths()
		ConfigDirCreation(path.dirPath)

		authorConfig := NewConfigFromCommand(cmd)
		authorConfig.UpdateFromOldConfig(path.file)

		config := Config{
			Author: authorConfig,
		}

		config.WriteConfig(path.file)
	},
}

var configUnsetCmd = &cobra.Command{
	Use:   "unset",
	Short: "Unset config Variables",
	Long:  "Unset config variables",
	Run: func(cmd *cobra.Command, args []string) {
		path := ConfigPaths()
		ConfigDirCreation(path.dirPath)

		authorConfig := &AuthorConfig{}
		authorConfig.UpdateFromOldConfig(path.file)
		authorConfig.Unset(cmd, args)

		config := Config{
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
		path := ConfigPaths()
		ConfigDirCreation(path.dirPath)

		authorConfig := &AuthorConfig{}
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
