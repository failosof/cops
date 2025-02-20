package core

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/failosof/cops/lichess"
	"github.com/failosof/cops/opening"
	"github.com/failosof/cops/puzzle"
	"github.com/failosof/cops/util"
	"github.com/notnil/chess"
)

type Resources struct {
	Opening OpeningResources
	Puzzle  PuzzleResources
	Chess   ChessResources
}

type OpeningResources struct {
	DatabaseDir string
	IndexFile   string
}

type PuzzleResources struct {
	DatabaseFile string
	IndexFile    string
}

type ChessResources struct {
	BackgroundFile string
	PiecesDir      string
}

func LoadOpeningsIndex(ctx context.Context, dbDir, indexFile string) (*opening.Index, error) {
	if err := os.MkdirAll(dbDir, os.ModePerm); err != nil {
		slog.Warn("failed to create openings database dbDir", "in", dbDir)
		return nil, err
	}

	if util.FileExists(indexFile) {
		slog.Info("loading openings index", "from", indexFile)
		openings, err := util.LoadBinary[opening.Index](indexFile)
		if err != nil {
			slog.Warn("failed to load openings index")
			return nil, err
		}
		return openings, nil
	}

	if !opening.Cached(dbDir) {
		slog.Info("downloading openings database", "from", opening.DatabaseURL, "to", dbDir)
		if _, err := opening.DownloadDatabase(ctx, dbDir); err != nil {
			slog.Warn("failed to download openings database")
			return nil, err
		}

		slog.Info("removing old openings index", "from", indexFile)
		if err := util.RemoveFile(indexFile); err != nil {
			slog.Warn("failed to remove old openings index")
			return nil, err
		}
	}

	slog.Info("creating openings index", "from", dbDir)
	openings, err := opening.CreateIndex(dbDir)
	if err != nil {
		slog.Warn("failed to create openings index")
		return nil, err
	}

	slog.Info("saving openings index", "to", indexFile)
	if err := util.SaveBinary(indexFile, openings); err != nil {
		slog.Warn("failed to save openings index")
		return nil, err
	}

	return openings, nil
}

func LoadPuzzlesIndex(ctx context.Context, dbFile, indexFile string) (*puzzle.Index, error) {
	if util.FileExists(indexFile) {
		slog.Info("loading puzzles index", "from", indexFile)
		puzzles, err := util.LoadBinary[puzzle.Index](indexFile)
		if err != nil {
			slog.Warn("failed to load puzzles index")
			return nil, err
		}
		return puzzles, nil
	}

	if !util.FileExists(dbFile) {
		slog.Info("downloading puzzles database", "from", lichess.PuzzleDatabaseURL, "to", dbFile)
		if err := util.Download(ctx, lichess.PuzzleDatabaseURL, dbFile); err != nil {
			slog.Warn("failed to download puzzles database")
			return nil, fmt.Errorf("failed to download: %w", err)
		}

		slog.Info("removing old puzzles index", "from", indexFile)
		if err := util.RemoveFile(indexFile); err != nil {
			slog.Warn("failed to remove old puzzles index")
			return nil, err
		}
	}

	slog.Info("creating puzzles index", "from", dbFile)
	puzzles, err := puzzle.CreateIndex(dbFile)
	if err != nil {
		slog.Warn("failed to create puzzles index")
		return nil, err
	}

	slog.Info("saving puzzles index", "to", indexFile)
	if err := util.SaveBinary(indexFile, puzzles); err != nil {
		slog.Warn("failed to save puzzles index", "err", err)
		return nil, err
	}

	return puzzles, nil
}

const (
	ChessBoardBackgroundURL = "https://raw.githubusercontent.com/failosof/cops/refs/heads/main/assets/board/brown.png"
	ChessPieceThemeURL      = "https://raw.githubusercontent.com/failosof/cops/refs/heads/main/assets/pieces/aquarium/"
)

func LoadChessBoardResources(ctx context.Context, backgroundImage, pieceThemeDir string) error {
	backgroundDir := filepath.Dir(backgroundImage)
	if err := os.MkdirAll(backgroundDir, os.ModePerm); err != nil {
		slog.Warn("failed to create chess background image dir", "in", backgroundDir)
		return err
	}

	if err := os.MkdirAll(pieceThemeDir, os.ModePerm); err != nil {
		slog.Warn("failed to create chess pieces theme dir", "in", pieceThemeDir)
		return err
	}

	if !util.FileExists(backgroundImage) {
		if err := util.Download(ctx, ChessBoardBackgroundURL, backgroundImage); err != nil {
			slog.Warn("failed to download chess board background image", "from", ChessBoardBackgroundURL)
			return err
		}
	}

	for _, color := range "wb" {
		slog.Info("loading chess pieces", "color", string(color))
		for _, pt := range chess.PieceTypes() {
			filename := fmt.Sprintf("%c%s.png", color, pt.String())
			fullPath := filepath.Join(pieceThemeDir, filename)
			if !util.FileExists(fullPath) {
				url := ChessPieceThemeURL + filename
				if err := util.Download(ctx, url, fullPath); err != nil {
					slog.Warn("failed to download chess piece file", "from", url)
					return err
				}
				slog.Debug("downloaded chess piece", "color", string(color), "type", pt.String())
			}
		}
	}

	return nil
}
