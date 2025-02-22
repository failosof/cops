package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/failosof/cops/cache"
	"github.com/failosof/cops/core"
	"github.com/failosof/cops/ui"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := cache.EnsureExists(); err != nil {
		fmt.Printf("Failed to ensure core cache dir exists: %v\n", err)
		os.Exit(1)
	}

	logFile, err := os.Create(cache.PathTo("log.txt"))
	if err != nil {
		fmt.Printf("Failed to create log file: %v\n", err)
		os.Exit(1)
	}

	handler := slog.NewTextHandler(io.MultiWriter(os.Stdout, logFile), &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	slog.SetDefault(slog.New(handler))

	resources := core.Resources{
		Opening: core.OpeningResources{
			DatabaseDir: cache.PathTo("database"),
			IndexFile:   cache.PathTo("openings.index"),
		},
		Game: core.GameResources{
			IndexFile: cache.PathTo("games.index"),
		},
		Puzzle: core.PuzzleResources{
			DatabaseFile: cache.PathTo("puzzles.csv.zst"),
			IndexFile:    cache.PathTo("puzzles.index"),
		},
		Chess: core.ChessResources{
			BackgroundFile: cache.PathTo("assets", "board", "brown.png"),
			PiecesDir:      cache.PathTo("assets", "pieces", "aquarium"),
		},
	}

	state, err := core.LoadState(ctx, resources.Opening, resources.Game, resources.Puzzle)
	if err != nil {
		slog.Error("failed to load state", "err", err)
		os.Exit(1)
	}

	if err := core.LoadChessBoardResources(
		ctx,
		resources.Chess.BackgroundFile,
		resources.Chess.PiecesDir,
	); err != nil {
		slog.Error("failed to load chess board resources", "err", err)
		os.Exit(1)
	}

	window, err := ui.NewWindow(state, resources.Chess)
	if err != nil {
		slog.Error("failed to create main window", "err", err)
		os.Exit(1)
	}
	window.Show(ctx)
}
