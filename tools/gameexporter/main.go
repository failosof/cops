package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/failosof/cops/game"
	"github.com/failosof/cops/puzzle"
	"github.com/failosof/cops/util"
)

const MemoryLimit int64 = 10 * 1024 * 1024 * 1024

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

	games, err := LoadGamesIndex(gameIndexFile)
	if err != nil {
		log.Fatalf("Failed to load games index: %v", err)
	}
	defer games.Save(gameIndexFile + ".extended")

	ctx := context.Background()
	ExportGames(ctx, puzzles, games, gameIndexFile)
	slog.Info("Finished exporting games")
}

func LoadPuzzlesIndex(filename string) (*puzzle.Index, error) {
	puzzles, err := util.LoadBinary[puzzle.Index](filename)
	if err != nil {
		slog.Warn("failed to load puzzles index")
		return nil, err
	}
	slog.Info("loaded puzzles index", "from", filename, "size", puzzles.Size())
	return puzzles, nil
}

func LoadGamesIndex(filename string) (*game.Index, error) {
	games, err := util.LoadBinary[game.Index](filename)
	if err != nil {
		slog.Warn("failed to load games index")
		return nil, err
	}
	slog.Info("loaded games index", "from", filename, "size", games.Size())
	return games, nil
}

func ExportGames(ctx context.Context, puzzles *puzzle.Index, games *game.Index, file string) {
	total = float32(puzzles.Size())
	exported := games.Size()
	slog.Info("Starting games export", "count", int(total)-exported)

	toExport := make([]string, 0, game.MaxExportNumber)
	var failed int
	for i, puzzle := range puzzles.Collection {
		if game := games.Search(puzzle.GameID); game.Empty() {
			toExport = append(toExport, puzzle.GameID.String())
		}

		if len(toExport) == game.MaxExportNumber || i == puzzles.Size()-1 {
			exportedGame, err := game.Export(ctx, toExport)
			if err != nil {
				slog.Error("failed to export games", "count", len(toExport), "err", err)
				failed += len(toExport)
			}

			for _, exportedGame := range exportedGame {
				exportedGameID := game.IDFromString(exportedGame.GetTagPair("GameId").Value)
				games.InsertFromChess(exportedGameID, exportedGame)
			}

			if err := games.Save(file + ".extended"); err != nil {
				slog.Error("failed to save games index", "err", err)
			}

			exported += len(toExport)
			toExport = make([]string, 0, game.MaxExportNumber)
		}

		fmt.Printf("\rTotal: %f%%, Exported: %f%%, Failed: %f%%", percent(i), percent(exported), percent(failed))
	}
	
	fmt.Println()
}

var total float32

func percent(val int) float32 {
	return float32(val) / total * 100
}
