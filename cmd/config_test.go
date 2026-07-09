/*
Package cmd
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestConfigDirCreationCreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "draftcat")

	if err := ConfigDirCreation(configDir); err != nil {
		t.Fatalf("ConfigDirCreation returned error: %v", err)
	}

	assertDirExists(t, configDir)
}

func TestConfigDirCreationSucceedsIfDirectoryAlreadyExists(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "draftcat")

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	if err := ConfigDirCreation(configDir); err != nil {
		t.Fatalf("expected existing directory to be okay, got error: %v", err)
	}
}

func TestConfigDirCreationFailsIfPathIsFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "draftcat")

	if err := os.WriteFile(configPath, []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	err := ConfigDirCreation(configPath)
	if err == nil {
		t.Fatal("expected error when config path is a file, got nil")
	}
}

func TestNewConfigFromCommandWithNoFlagsChanged(t *testing.T) {
	cmd := newTestConfigSetCmd(t)

	authorConfig, err := NewConfigFromCommand(cmd)
	if err != nil {
		t.Fatalf("NewConfigFromCommand returned error: %v", err)
	}

	if authorConfig.Name != "" {
		t.Fatalf("expected empty name, got %q", authorConfig.Name)
	}

	if authorConfig.Email != "" {
		t.Fatalf("expected empty email, got %q", authorConfig.Email)
	}

	if authorConfig.Address != "" {
		t.Fatalf("expected empty address, got %q", authorConfig.Address)
	}

	if authorConfig.Phone != "" {
		t.Fatalf("expected empty phone, got %q", authorConfig.Phone)
	}
}

func TestNewConfigFromCommandReadsChangedFlags(t *testing.T) {
	cmd := newTestConfigSetCmd(t)

	if err := cmd.Flags().Set("name", "Moses Sukumaran"); err != nil {
		t.Fatal(err)
	}

	if err := cmd.Flags().Set("email", "moses@example.test"); err != nil {
		t.Fatal(err)
	}

	if err := cmd.Flags().Set("address", "New Delhi"); err != nil {
		t.Fatal(err)
	}

	if err := cmd.Flags().Set("phone", "+91 9999999999"); err != nil {
		t.Fatal(err)
	}

	authorConfig, err := NewConfigFromCommand(cmd)
	if err != nil {
		t.Fatalf("NewConfigFromCommand returned error: %v", err)
	}

	if authorConfig.Name != "Moses Sukumaran" {
		t.Fatalf("expected name %q, got %q", "Moses Sukumaran", authorConfig.Name)
	}

	if authorConfig.Email != "moses@example.test" {
		t.Fatalf("expected email %q, got %q", "moses@example.test", authorConfig.Email)
	}

	if authorConfig.Address != "New Delhi" {
		t.Fatalf("expected address %q, got %q", "New Delhi", authorConfig.Address)
	}

	if authorConfig.Phone != "+91 9999999999" {
		t.Fatalf("expected phone %q, got %q", "+91 9999999999", authorConfig.Phone)
	}
}

func TestConfigPathsReturnsDraftcatConfigPath(t *testing.T) {
	paths, err := ConfigPaths()
	if err != nil {
		t.Fatalf("ConfigPaths returned error: %v", err)
	}

	if filepath.Base(paths.dirPath) != "draftcat" {
		t.Fatalf("expected config dir base %q, got %q", "draftcat", filepath.Base(paths.dirPath))
	}

	if filepath.Base(paths.file) != "draftcat-config.toml" {
		t.Fatalf("expected config file base %q, got %q", "draftcat-config.toml", filepath.Base(paths.file))
	}

	if filepath.Dir(paths.file) != paths.dirPath {
		t.Fatalf("expected config file dir %q, got %q", paths.dirPath, filepath.Dir(paths.file))
	}
}

func newTestConfigSetCmd(t *testing.T) *cobra.Command {
	t.Helper()

	cmd := &cobra.Command{
		Use: "set",
	}

	cmd.Flags().StringP("name", "n", "", "Set author name")
	cmd.Flags().StringP("email", "e", "", "Set email")
	cmd.Flags().StringP("address", "a", "", "Set address")
	cmd.Flags().StringP("phone", "p", "", "Set Phone number")

	return cmd
}
