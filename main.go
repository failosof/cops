package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/failosof/cops/app"
	"github.com/failosof/cops/cache"
)

func main() {
	if len(os.Args) < 5 {
		fmt.Printf("Usage: %s <pgn> <turn> <min> <max>\n", os.Args[0])
		os.Exit(1)
	}

	config, err := app.ParseConfig(os.Args)
	if err != nil {
		fmt.Printf("Failed to parse input: %v\n", err)
		os.Exit(1)
	}

	if err := cache.EnsureExists(); err != nil {
		fmt.Printf("Failed to ensure app cache dir exists: %v\n", err)
		os.Exit(1)
	}

	logFile, err := os.Create(cache.PathTo("log.txt"))
	if err != nil {
		fmt.Printf("Failed to create log file: %v\n", err)
		os.Exit(1)
	}

	log := slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	cops := app.State{
		Log:    log,
		Config: config,
		Files:  app.DefaultFiles(),
	}

	if err := cops.LoadOpenings(); err != nil {
		log.Error("failed to load openings", "err", err)
		os.Exit(1)
	}

	if err := cops.LoadPuzzles(); err != nil {
		log.Error("failed to load puzzles", "err", err)
		os.Exit(1)
	}

	openingName := cops.SearchOpening()
	if !openingName.Empty() {
		puzzles := cops.SearchPuzzles(openingName)
		for _, puzzle := range puzzles {
			fmt.Println(puzzle.URL())
		}
	}
}
