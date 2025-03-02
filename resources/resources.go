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

func LoadIndex[I any](filename string) (i I, err error) {
	file, err := Indexes.Open(filename)
	if err != nil {
		err = fmt.Errorf("failed to open index file %s: %w", filename, err)
		return
	}
	defer file.Close()

	decoder, err := zstd.NewReader(file)
	if err != nil {
		err = fmt.Errorf("failed to create zst decoder for %s: %w", filename, err)
		return
	}
	defer decoder.Close()

	if err = gob.NewDecoder(decoder).Decode(&i); err != nil {
		err = fmt.Errorf("failed to read binary file %s: %w", err)
		return
	}

	return
}
