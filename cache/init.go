package cache

import (
	"fmt"
	"os"
	"path/filepath"
)

var dir string

func EnsureExists() error {
	userDir, err := os.UserCacheDir()
	if err != nil {
		return fmt.Errorf("no user cache dir: %w", err)
	}

	appDir := filepath.Join(userDir, "failosof", "cops")
	if err := os.MkdirAll(appDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create core cache: %w", err)
	}

	dir = appDir
	return nil
}

func PathTo(names ...string) string {
	parts := make([]string, 0, len(names)+1)
	parts = append(parts, dir)
	parts = append(parts, names...)
	return filepath.Join(parts...)
}

func Clear() error {
	if err := os.Remove(dir); err != nil {
		return fmt.Errorf("failed to invalidate cache: %w", err)
	}
	return nil
}
