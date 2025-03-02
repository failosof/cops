package util

import (
	"fmt"
	"os"
)

func RemoveFile(filename string) error {
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove file %q: %w", filename, err)
	}
	return nil
}
