package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/failosof/cops/cache"
	"github.com/failosof/cops/core"
	"github.com/failosof/cops/ui"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cancel()
	}()

	if err := cache.EnsureExists(); err != nil {
		fmt.Printf("Failed to ensure core cache dir exists: %v\n", err)
		os.Exit(1)
	}

	logFile, err := os.Create(cache.PathTo("log.txt"))
	if err != nil {
		fmt.Printf("Failed to create log file: %v\n", err)
		os.Exit(1)
	}

	handler := slog.NewTextHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	slog.SetDefault(slog.New(handler))

	resources := core.Resources{
		Opening: core.OpeningResources{
			DatabaseDir: cache.PathTo("database"),
			IndexFile:   cache.PathTo("openings.index"),
		},
		Puzzle: core.PuzzleResources{
			DatabaseFile: cache.PathTo("puzzles.csv.zst"),
			IndexFile:    cache.PathTo("puzzles.index"),
		},
		ChessBoard: core.ChessBoardResources{
			BackgroundFile: cache.PathTo("assets", "board", "brown.png"),
			PiecesDir:      cache.PathTo("assets", "pieces", "aquarium"),
		},
	}

	state, err := core.LoadState(ctx, resources.Opening, resources.Puzzle)
	if err != nil {
		slog.Error("failed to load state", "err", err)
		os.Exit(1)
	}

	if err := core.LoadChessBoardResources(
		ctx,
		resources.ChessBoard.BackgroundFile,
		resources.ChessBoard.PiecesDir,
	); err != nil {
		slog.Error("failed to load chess board resources", "err", err)
		os.Exit(1)
	}

	ui.DrawMainWindow(ctx, state, resources.ChessBoard)
}
