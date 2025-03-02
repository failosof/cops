package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/failosof/cops/core"
	"github.com/failosof/cops/tools/util"
)

const IDPoolSize = 1300 // 1.3k ids per file

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: openingindexer <openings-db-dir>")
	}

	dir := os.Args[1]
	ctx := context.Background()

	if !Cached(dir) {
		log.Printf("No openings database found in %q, downloading ...", dir)
		if _, err := DownloadDatabase(ctx, dir); err != nil {
			log.Fatalf("Failed to download openings database: %v", err)
		}
	}

	index, err := CreateOpeningsIndex(dir)
	if err != nil {
		log.Fatalf("Failed to create openings index: %v", err)
	}

	filename := "openings.index"
	if err := util.SaveBinary(filename, index); err != nil {
		log.Fatalf("Failed to save openings index: %v", err)
	}

	filename, _ = filepath.Abs(filename)
	log.Printf("Saved to %q", filename)
}

func CreateOpeningsIndex(dir string) (core.OpeningsIndex, error) {
	index := make(core.OpeningsIndex, len(filenames)*IDPoolSize)

	var indexed int
	for _, filename := range filenames {
		filename = filepath.Join(dir, filename)
		n, err := parseFile(filename, index)
		if err != nil {
			return nil, err
		}
		indexed += n
		log.Printf("Updated openings index from %q indexed %d", filename, n)
	}
	log.Printf("Created openings index from %q indexed %d", dir, indexed)

	return index, nil
}

func parseFile(filename string, index core.OpeningsIndex) (n int, err error) {
	file, err := os.Open(filename)
	if err != nil {
		err = fmt.Errorf("failed to open file %q: %w", filename, err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = '\t'
	reader.ReuseRecord = true

	var line []string
	for i := 0; ; i++ {
		lineNum := i + 1
		line, err = reader.Read()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				err = fmt.Errorf("failed to read file %q line %d: %w", filename, lineNum, err)
			} else {
				err = nil
			}
			return
		}

		if lineNum > 1 { // skip the header
			if len(line) != 3 {
				err = fmt.Errorf("file %q line %d: want 3 fields, have %d", filename, lineNum, len(line))
				return
			}
			if err = index.Insert(line[1], line[2]); err != nil {
				return
			}
			n++
		}
	}
}
