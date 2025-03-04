package util

import (
	"encoding/gob"
	"fmt"
	"os"
)

func LoadBinary[T any](filename string) (t T, err error) {
	file, err := os.Open(filename)
	if err != nil {
		err = fmt.Errorf("failed to open binary file %q: %w", filename, err)
		return
	}
	defer file.Close()

	if err = gob.NewDecoder(file).Decode(&t); err != nil {
		err = fmt.Errorf("failed to read binary file: %w", err)
		return
	}

	return
}

func SaveBinary[T any](filename string, t T) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create binary file %q: %w", filename, err)
	}
	defer file.Close()

	if err = gob.NewEncoder(file).Encode(t); err != nil {
		return fmt.Errorf("failed to write binary file: %w", err)
	}

	if err = file.Sync(); err != nil {
		return fmt.Errorf("failed to sync binary file: %w", err)
	}

	return nil
}
