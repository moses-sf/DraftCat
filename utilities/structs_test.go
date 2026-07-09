/*
Package utilities
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package utilities

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestStoryConfigWriteAndLoadConfig(t *testing.T) {
	tempDir := t.TempDir()

	config := &StoryConfig{
		Author: AuthorConfig{
			Name:    "Moses",
			Email:   "moses@example.test",
			Address: "New Delhi",
			Phone:   "+91 9999999999",
		},
		MetaData: StoryMetaData{
			Name: "Monkey",
			Type: Short,
		},
	}

	path := filepath.Join(tempDir, ".story.toml")

	if err := config.WriteConfig(path); err != nil {
		t.Fatalf("WriteConfig returned error: %v", err)
	}

	loaded := &StoryConfig{}
	if err := loaded.LoadConfig(tempDir); err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if loaded.MetaData.Name != "Monkey" {
		t.Fatalf("expected story name Monkey, got %q", loaded.MetaData.Name)
	}

	if loaded.MetaData.Type != Short {
		t.Fatalf("expected story type %q, got %q", Short, loaded.MetaData.Type)
	}

	if loaded.Author.Name != "Moses" {
		t.Fatalf("expected author Moses, got %q", loaded.Author.Name)
	}
}

func TestAuthorConfigFillEmpty(t *testing.T) {
	current := AuthorConfig{
		Name: "Moses",
	}

	old := AuthorConfig{
		Name:    "Old Name",
		Email:   "old@example.test",
		Address: "Old Address",
		Phone:   "12345",
	}

	current.FillEmpty(old)

	if current.Name != "Moses" {
		t.Fatalf("expected existing name to remain Moses, got %q", current.Name)
	}

	if current.Email != "old@example.test" {
		t.Fatalf("expected email to be filled, got %q", current.Email)
	}

	if current.Address != "Old Address" {
		t.Fatalf("expected address to be filled, got %q", current.Address)
	}

	if current.Phone != "12345" {
		t.Fatalf("expected phone to be filled, got %q", current.Phone)
	}
}

func TestAuthorConfigUpdateFromOldConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "draftcat-config.toml")

	config := Config{
		Author: AuthorConfig{
			Name:    "Moses",
			Email:   "moses@example.test",
			Address: "New Delhi",
			Phone:   "12345",
		},
	}

	if err := config.WriteConfig(configPath); err != nil {
		t.Fatalf("WriteConfig returned error: %v", err)
	}

	author := &AuthorConfig{
		Name: "Override Name",
	}

	if err := author.UpdateFromOldConfig(configPath); err != nil {
		t.Fatalf("UpdateFromOldConfig returned error: %v", err)
	}

	if author.Name != "Override Name" {
		t.Fatalf("expected existing name to remain, got %q", author.Name)
	}

	if author.Email != "moses@example.test" {
		t.Fatalf("expected email from old config, got %q", author.Email)
	}

	if author.Address != "New Delhi" {
		t.Fatalf("expected address from old config, got %q", author.Address)
	}

	if author.Phone != "12345" {
		t.Fatalf("expected phone from old config, got %q", author.Phone)
	}
}

func TestConfigWriteConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "draftcat-config.toml")

	config := &Config{
		Author: AuthorConfig{
			Name:  "Moses",
			Email: "moses@example.test",
		},
	}

	if err := config.WriteConfig(configPath); err != nil {
		t.Fatalf("WriteConfig returned error: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("could not read config: %v", err)
	}

	var loaded Config
	if _, err := toml.Decode(string(data), &loaded); err != nil {
		t.Fatalf("could not decode config: %v", err)
	}

	if loaded.Author.Name != "Moses" {
		t.Fatalf("expected author Moses, got %q", loaded.Author.Name)
	}
}

func TestStoryStructureJSONRenderOutputsJSON(t *testing.T) {
	story := &StoryStructure{
		Name: "Monkey",
		Root: "/tmp/Monkey",
		Type: string(Short),
	}

	output := captureStdout(t, func() {
		story.JSONRender()
	})

	output = strings.TrimSpace(output)

	var decoded StoryStructure
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("JSONRender output was not valid JSON: %v\noutput: %s", err, output)
	}

	if decoded.Name != "Monkey" {
		t.Fatalf("expected name Monkey, got %q", decoded.Name)
	}

	if decoded.Root != "/tmp/Monkey" {
		t.Fatalf("expected root /tmp/Monkey, got %q", decoded.Root)
	}
}

func writeChapterTomlForTest(path string, chapter ChapterMetaData) error {
	buf := new(bytes.Buffer)

	if err := toml.NewEncoder(buf).Encode(chapter); err != nil {
		return err
	}

	return os.WriteFile(path, buf.Bytes(), 0o644)
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stdout = writer

	fn()

	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	os.Stdout = oldStdout

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}

	return string(output)
}

// Compile-time guard so sql stays imported for ChapterMetaData patterns in this package's tests.
var _ = sql.NullInt64{}
