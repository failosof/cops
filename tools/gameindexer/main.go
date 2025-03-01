package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"maps"
	"math"
	"os"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"

	"github.com/failosof/cops/game"
	"github.com/failosof/cops/util"
	"github.com/goccy/go-json"
	"golang.org/x/exp/mmap"
)

var (
	MemoryLimit      int64   = 10 * 1024 * 1024 * 1024
	TotalRecords     float64 = 2_969_948
	FileRecords              = int(math.Ceil(TotalRecords / 4)) // need to split db to 4 files
	AssumedGameCount         = 607_870                          // combined db has only ~608k opening puzzles
)

var processed, indexed atomic.Int64

type Record struct {
	Puzzle struct {
		OpeningFamily string `json:"OpeningFamily"`
	} `json:"puzzle"`
	Game struct {
		ID    string `json:"id"`
		Moves string `json:"moves"`
	} `json:"game"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: gameindexer <combined_puzzle_db.ndjson.part_n> ...")
	}

	debug.SetGCPercent(-1)
	debug.SetMemoryLimit(MemoryLimit)

	filenames := os.Args[1:]

	log.Printf("Indexing games from %s ...", strings.Join(filenames, ", "))
	index, err := createIndex(filenames)
	if err != nil {
		log.Fatalf("Failed to create games index: %v", err)
	}

	log.Println("Saving file ...")
	file := "games.index"
	if err := util.SaveBinary(file, &index); err != nil {
		log.Fatalf("Failed to save games index: %v", err)
	}

	log.Printf("Index of %d games created in %q\n", len(index), file)
}

type fileResult struct {
	index game.Index
	err   error
}

func createIndex(filenames []string) (game.Index, error) {
	resChan := make(chan fileResult, len(filenames))
	defer close(resChan)

	for _, filename := range filenames {
		go func(filename string) {
			idx, err := processFile(filename)
			resChan <- fileResult{index: idx, err: err}
		}(filename)
	}

	index := make(game.Index, AssumedGameCount)
	workers := len(filenames)
loop:
	for {
		select {
		case res := <-resChan:
			if res.err != nil {
				return nil, res.err
			}
			maps.Copy(index, res.index)
			workers--
			if workers == 0 {
				break loop
			}
		case <-time.After(500 * time.Millisecond):
			fmt.Printf("\rProgress: %.2f%%", float64(processed.Load())/TotalRecords*100)
		}
	}

	fmt.Printf("\rIndexed %d games out of %d records\n", indexed.Load(), processed.Load())

	return index, nil
}

func processFile(filename string) (game.Index, error) {
	file, err := mmap.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %w", filename, err)
	}
	defer file.Close()

	index := make(game.Index, FileRecords)
	
	sr := io.NewSectionReader(file, 0, int64(file.Len()))
	decoder := json.NewDecoder(sr)
	for {
		var record Record
		if err := decoder.Decode(&record); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("failed to decode record: %w", err)
		}

		if len(record.Puzzle.OpeningFamily) > 0 {
			if err := index.Insert(record.Game.ID, record.Game.Moves); err != nil {
				return nil, fmt.Errorf("failed to index game %s: %w", record.Game.ID, err)
			}
			indexed.Add(1)
		}

		processed.Add(1)
	}

	return index, nil
}
