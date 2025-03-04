package resources

import (
	"embed"
	"encoding/gob"
	"fmt"
)

//go:embed indexes/*.index
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

	if err = gob.NewDecoder(file).Decode(&i); err != nil {
		err = fmt.Errorf("failed to read binary file %s: %w", err)
		return
	}

	return
}
