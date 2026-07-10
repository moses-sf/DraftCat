/*
Package utilities
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package utilities

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

type SceneMetaData struct {
	ID       int
	Name     string
	Path     string
	Position int
}

type ChapterMetaData struct {
	ID         int
	Name       string
	ParentID   sql.NullInt64
	PathToRoot string
	Position   int
	Scenes     []*SceneMetaData
}
type StoryType string

const (
	Short   StoryType = "Short Story"
	Novella StoryType = "Novella"
	Novel   StoryType = "Novel"
)

type StoryMetaData struct {
	Name string
	Type StoryType
}

type StoryConfig struct {
	Author   AuthorConfig
	MetaData StoryMetaData
}

func (s *StoryConfig) WriteConfig(path string) error {
	buf := new(bytes.Buffer)
	encoder := toml.NewEncoder(buf)
	err := encoder.Encode(s)
	if err != nil {
		return err
	}
	err = os.WriteFile(path, buf.Bytes(), 0o644)
	if err != nil {
		return err
	}
	return nil
}

func (s *StoryConfig) LoadConfig(root string) error {
	file, err := os.ReadFile(filepath.Join(root, ".story.toml"))
	if err != nil {
		return err
	}
	_, err = toml.Decode(string(file), s)
	if err != nil {
		return err
	}
	return nil
}

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

func (a *AuthorConfig) UpdateFromOldConfig(configPath string) error {
	_, err := os.Stat(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
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
	return nil
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

func (c *Config) WriteConfig(path string) error {
	buf := new(bytes.Buffer)
	encoder := toml.NewEncoder(buf)
	err := encoder.Encode(c)
	if err != nil {
		return err
	}
	err = os.WriteFile(path, buf.Bytes(), 0o644)
	if err != nil {
		return err
	}
	return nil
}
