package util

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func Download(from, to string) error {
	resp, err := http.Get(from)
	if err != nil {
		return fmt.Errorf("failed to request url %q: %w", from, err)
	}
	defer resp.Body.Close()

	file, err := os.Create(to)
	if err != nil {
		return fmt.Errorf("failed to create %q file: %w", to, err)
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}
