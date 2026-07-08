/*
Package utilities
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package utilities

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

func FindToml(tomlName string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return FindTomlPath(cwd, tomlName)
}

func FindTomlPath(path, tomlName string) (string, error) {
	_, err := os.Stat(filepath.Join(path, tomlName))
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Attempting to find local file")
		}
		return "", err
	} else {
		return path, nil
	}
}

func FindFolderToml() (string, error) {
	return "", nil
}

func GetFolderToml(path string) bool {
	_, err := os.Stat(filepath.Join(path))
	return err == nil
}

func LoadChapterToml(path string) (*ChapterMetaData, error) {
	chapter := &ChapterMetaData{}
	file, err := os.ReadFile(filepath.Join(path, ".chapter.toml"))
	if err != nil {
		return nil, err
	}
	_, err = toml.Decode(string(file), chapter)
	if err != nil {
		return nil, err
	}
	return chapter, nil
}

func IsDraftcatProject() (bool, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return false, err
	}
	chapterPath := filepath.Join(cwd, ".chapter.toml")
	storyPath := filepath.Join(cwd, ".story.toml")
	if GetFolderToml(chapterPath) || GetFolderToml(storyPath) {
		return true, nil
	}
	return false, fmt.Errorf("not a draftcat project")
}
