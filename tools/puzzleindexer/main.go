package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/failosof/cops/core"
	"github.com/failosof/cops/tools/util"
	"github.com/klauspost/compress/zstd"
)

const DatabaseURL = "https://database.lichess.org/lichess_db_puzzle.csv.zst"
const AssumedPuzzleCount = 1_050_000 // puzzle db is ~4.5m records, only ~1m are from openings

func main() {
	var filename string
	if len(os.Args) < 2 {
		log.Println("Hint: you may specify filename as first parameter to omit downloading")
	} else {
		filename = os.Args[1]
	}

	ctx := context.Background()
	if len(filename) == 0 {
		log.Println("Downloading lichess puzzle database ...")
		filename = "lichess_db_puzzle.csv.zst"
		if err := util.Download(ctx, DatabaseURL, filename); err != nil {
			log.Fatalf("Failed to download lichess puzzle database: %v", err)
		}
	}

	log.Println("Indexing opening puzzles ...")
	index, err := CreatePuzzlesIndex(filename)
	if err != nil {
		log.Fatalf("Failed to create puzzles index: %v", err)
	}

	log.Println("Saving index ...")
	filename = "puzzles.index"
	if err := util.SaveBinary(filename, index); err != nil {
		log.Fatalf("Failed to save puzzles index: %v", err)
	}

	filename, _ = filepath.Abs(filename)
	log.Printf("Saved to %q", filename)
}

func CreatePuzzlesIndex(from string) (core.PuzzlesIndex, error) {
	index := make(core.PuzzlesIndex, AssumedPuzzleCount)

	filename := from
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %w", filename, err)
	}
	defer file.Close()

	decoder, err := zstd.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create zst decoder for %q: %w", filename, err)
	}
	defer decoder.Close()

	reader := csv.NewReader(decoder)
	reader.ReuseRecord = true

	var indexed, processed int
	for i := 0; ; i++ {
		lineNum := i + 1
		line, err := reader.Read()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return nil, fmt.Errorf("failed to read file %q line %d: %w", filename, lineNum, err)
			}
			break
		}

		if lineNum > 1 { // skip the header
			if len(line) != 10 {
				return nil, fmt.Errorf("file %q line %d: want 10 fields, have %d", filename, lineNum, len(line))
			}

			if len(line[9]) > 0 {
				if err := index.Insert(line[0], line[1], line[8], line[9]); err != nil {
					return nil, fmt.Errorf("file %q line %d: %w", filename, lineNum, err)
				}
				indexed++
			}
			processed++
		}

		fmt.Printf("\rProcessed: %d, Indexed: %d (~%.2f%%)", processed, indexed, percent(indexed, AssumedPuzzleCount))
	}

	fmt.Println()
	slog.Debug("created puzzles index", "from", from, "processed", processed, "indexed", indexed)

	return index, nil
}

func percent(num, of int) float32 {
	return float32(num) / float32(of) * 100
}
