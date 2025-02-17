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
		return fmt.Errorf("failed to create app cache: %w", err)
	}

	dir = appDir
	return nil
}

func PathTo(name string) string {
	return filepath.Join(dir, name)
}

func Clear() error {
	if err := os.Remove(dir); err != nil {
		return fmt.Errorf("failed to invalidate cache: %w", err)
	}
	return nil
}
