package util

import (
	"fmt"
	"os"
)

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	return err == nil && info.Size() > 0
}

func RemoveFile(filename string) error {
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove file %q: %w", filename, err)
	}
	return nil
}
