/*
Package utilities
Copyright © 2026 Moses Sukumaran moses@solframe.in
*/
package utilities

import (
	"database/sql"
	"fmt"
	"path/filepath"
)

func GetDBPath() (string, error) {
	root, err := GetRelativeRootPath()
	if err != nil {
		return "", err
	}
	dbPath := filepath.Join(root, ".story.db")
	return dbPath, nil
}

func OpenDB() (*sql.DB, error) {
	dbPath, err := GetDBPath()
	if err != nil {
		return nil, fmt.Errorf("get database path: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	return db, nil
}
