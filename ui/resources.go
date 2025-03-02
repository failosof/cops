package ui

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/failosof/cops/resources"
	"github.com/failosof/cops/ui/board"
	"github.com/notnil/chess"
)

func LoadChessBoardConfig() (*chessboard.Config, error) {
	boardFile, err := resources.Assets.Open(filepath.Join("assets", "board", "brown.png"))
	if err != nil {
		return nil, fmt.Errorf("failed to load board asset: %w", err)
	}
	defer boardFile.Close()

	start := time.Now()
	boardTexture, err := chessboard.LoadTexture(boardFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load board texture: %w", err)
	}
	slog.Info("loaded board texture", "took", time.Since(start))

	start = time.Now()
	pieceTextures := make([]chessboard.Texture, 13)
	for piece := chess.WhiteKing; piece <= chess.BlackPawn; piece++ {
		fileName := filepath.Join("assets", "pieces", "aquarium", fmt.Sprintf("%s%s.png", piece.Color(), piece.Type()))
		pieceFile, err := resources.Assets.Open(fileName)
		if err != nil {
			return nil, fmt.Errorf("failed to load piece file %q: %w", fileName, err)
		}

		pieceTexture, err := chessboard.LoadTexture(pieceFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load piece texture %q: %w", fileName, err)
		}

		pieceTextures[piece] = pieceTexture
	}
	slog.Info("loaded piece textures", "took", time.Since(start))

	return chessboard.NewConfig(boardTexture, pieceTextures, chessboard.Colors{
		Hint:     Transparentize(GrayColor, 0.7),
		LastMove: Transparentize(YellowColor, 0.5),
		Primary:  Transparentize(GreenColor, 0.7),
		Info:     Transparentize(BlueColor, 0.7),
		Warning:  Transparentize(YellowColor, 0.7),
		Danger:   Transparentize(RedColor, 0.7),
	}), nil
}
