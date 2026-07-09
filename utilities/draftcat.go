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

func FolderTomlExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func LoadChapterToml(root string) (*ChapterMetaData, error) {
	chapter := &ChapterMetaData{}
	file, err := os.ReadFile(filepath.Join(root, ".chapter.toml"))
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
	if FolderTomlExists(chapterPath) || FolderTomlExists(storyPath) {
		return true, nil
	}
	return false, fmt.Errorf("not a draftcat project")
}

func GetRelativeRootPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	chapterPath := filepath.Join(cwd, ".chapter.toml")
	storyPath := filepath.Join(cwd, ".story.toml")
	if FolderTomlExists(storyPath) {
		return cwd, nil
	} else if FolderTomlExists(chapterPath) {
		chapter, err := LoadChapterToml(cwd)
		if err != nil {
			return "", fmt.Errorf("error loading chapter data %s", err)
		}
		return filepath.Join(cwd, chapter.PathToRoot), nil
	} else {
		return "", fmt.Errorf("could not get root")
	}
}
