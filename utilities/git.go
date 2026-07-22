/*
Package utilities
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package utilities

import (
	"fmt"
	"os/exec"
)

func CommitBackupSnapshot(command string) error {
	message := "draftcat|backup|" + command
	return CommitWithMessage(message)
}

func CommitWithMessage(message string) error {
	root, err := GetRelativeRootPath()
	if err != nil {
		return err
	}
	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git init: %w\n%s", err, output)
	}
	cmd = exec.Command("git", "commit", "--allow-empty", "-m", message)
	cmd.Dir = root
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git init: %w\n%s", err, output)
	}
	return nil
}

func CommitSnapshotReturnHash(root, message string) (string, error) {
	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git init: %w\n%s", err, output)
	}
	cmd = exec.Command("git", "commit", "--allow-empty", "-m", message)
	cmd.Dir = root
	output, err = cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git commit: %w\n%s", err, output)
	}
	hashOutput, err := RetrieveLastHash(root)
	if err != nil {
		return "", fmt.Errorf("git hash retrieval: %w\n%s", err, output)
	}
	return hashOutput, nil
}

func RetrieveLastHash(root string) (string, error) {
	cmd := exec.Command("git", "log", "-n", "1", "--format=%H")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git hash retrieval: %w\n%s", err, output)
	}
	return string(output), nil
}

func RestoreChanges() error {
	root, err := GetRelativeRootPath()
	if err != nil {
		return err
	}
	cmd := exec.Command("git", "restore", ".")
	cmd.Dir = root
	_, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git restore error: %w", err)
	}
	return nil
}
