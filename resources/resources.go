package resources

import (
	"embed"
	"encoding/gob"
	"fmt"

	"github.com/klauspost/compress/zstd"
)

//go:embed indexes
var Indexes embed.FS

//go:embed assets
var Assets embed.FS

func LoadIndex[T any](filename string) (*T, error) {
	file, err := Indexes.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open index file %s: %w", filename, err)
	}
	defer file.Close()

	decoder, err := zstd.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create zst decoder for %s: %w", filename, err)
	}
	defer decoder.Close()

	var t T
	if err := gob.NewDecoder(decoder).Decode(&t); err != nil {
		return nil, fmt.Errorf("failed to read binary file %s: %w", err)
	}

	return &t, nil
}
