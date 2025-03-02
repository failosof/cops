package main

import (
	"context"
	"fmt"
	"github.com/failosof/cops/core"
	"github.com/failosof/cops/tools/util"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime/debug"
	"slices"
	"time"
)

const MemoryLimit int64 = 10 * 1024 * 1024 * 1024 // 10 Gb

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("Usage: gameexporter <puzzles.index> <games.index>")
	}

	debug.SetGCPercent(-1)
	debug.SetMemoryLimit(MemoryLimit)

	puzzleIndexFile, gameIndexFile := os.Args[1], os.Args[2]

	puzzles, err := LoadPuzzlesIndex(puzzleIndexFile)
	if err != nil {
		log.Fatalf("Failed to load puzzles index: %v", err)
	}

	filename := "games.index.extended"
	games, err := LoadGamesIndex(gameIndexFile)
	if err != nil {
		log.Fatalf("Failed to load games index: %v", err)
	}
	defer util.SaveBinary(filename, games)

	log.Printf("Starting games export from %d", len(games))
	ExportGames(context.Background(), puzzles, games, filename)

	filename, _ = filepath.Abs(filename)
	log.Printf("Saved to %q", filename)
}

func LoadPuzzlesIndex(filename string) (core.PuzzlesIndex, error) {
	puzzles, err := util.LoadBinary[core.PuzzlesIndex](filename)
	if err != nil {
		log.Println("failed to load puzzles index")
		return nil, err
	}
	log.Printf("loaded %d puzzles from %q", len(puzzles), filename)
	return puzzles, nil
}

func LoadGamesIndex(filename string) (core.GamesIndex, error) {
	games, err := util.LoadBinary[core.GamesIndex](filename)
	if err != nil {
		log.Println("failed to load games index")
		return nil, err
	}
	log.Printf("loaded %d games from %q", len(games), filename)
	return games, nil
}

func ExportGames(ctx context.Context, puzzles core.PuzzlesIndex, gamesIndex core.GamesIndex, filename string) {
	log.Println("Collecting unexported games ...")
	toExport := make([]string, 0, len(gamesIndex))
	for _, puzzleCollection := range puzzles {
		for _, puzzle := range puzzleCollection {
			if game := gamesIndex[puzzle.GameID]; game.Empty() {
				toExport = append(toExport, puzzle.GameID.String())
			}
		}
	}

	total = float32(len(toExport))
	log.Printf("Collected %d unexported games", len(toExport))
	n := int(math.Ceil(float64(len(toExport)) / MaxExportNumber))
	nDur := time.Duration(n)
	log.Printf("%d export requests needed, min eta: %v, max eta: %v", n, nDur*limit, nDur*time.Minute)

	var exported, failed int
	for toExportChunk := range slices.Chunk(toExport, MaxExportNumber) {
		var fail bool

		games, err := ExportFromLichess(ctx, toExportChunk)
		if err != nil {
			fmt.Println()
			log.Printf("failed to export %d games: %v", len(toExportChunk), err)
			failed += len(games)
			fail = true
			incLimit()
		}

		for _, game := range games {
			exportedGameID := core.ParseGameID(game.GetTagPair("GameId").Value)
			gamesIndex.InsertFromChess(exportedGameID, game)
		}

		if err := util.SaveBinary(filename, gamesIndex); err != nil {
			fmt.Println()
			log.Printf("failed to save games: %v", err)
			fail = true
		}

		if !fail {
			exported += len(games)
		}

		fmt.Printf("\rExported: %10f%%, Failed: %10f%%", percent(exported), percent(failed))
	}

	fmt.Println()
}

var total float32

func percent(val int) float32 {
	return float32(val) / total * 100
}
