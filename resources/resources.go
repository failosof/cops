package resources

import (
	"embed"
	"encoding/gob"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/failosof/giochess/board"
	"github.com/notnil/chess"
)

//go:embed indexes/*.index
var Indexes embed.FS

//go:embed assets
var Assets embed.FS

func LoadIndex[I any](filename string) (i I, err error) {
	file, err := Indexes.Open(filename)
	if err != nil {
		err = fmt.Errorf("failed to open index file %s: %w", filename, err)
		return
	}
	defer file.Close()

	if err = gob.NewDecoder(file).Decode(&i); err != nil {
		err = fmt.Errorf("failed to read binary file %s: %w", err)
		return
	}

	return
}

type ChessBoardTextures struct {
	Board  chessboard.Texture
	Pieces []chessboard.Texture
}

func LoadChessBoardTextures() (*ChessBoardTextures, error) {
	boardFile, err := Assets.Open(filepath.Join("assets", "board", "brown.png"))
	if err != nil {
		return nil, fmt.Errorf("failed to load board asset: %w", err)
	}
	defer boardFile.Close()

	var textures ChessBoardTextures

	start := time.Now()
	textures.Board, err = chessboard.LoadTexture(boardFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load board texture: %w", err)
	}
	slog.Info("loaded board texture", "took", time.Since(start))

	start = time.Now()
	textures.Pieces = make([]chessboard.Texture, 13)
	for piece := chess.WhiteKing; piece <= chess.BlackPawn; piece++ {
		fileName := filepath.Join("assets", "pieces", "aquarium", fmt.Sprintf("%s%s.png", piece.Color(), piece.Type()))
		pieceFile, err := Assets.Open(fileName)
		if err != nil {
			return nil, fmt.Errorf("failed to load piece file %q: %w", fileName, err)
		}

		pieceTexture, err := chessboard.LoadTexture(pieceFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load piece texture %q: %w", fileName, err)
		}

		textures.Pieces[piece] = pieceTexture
	}
	slog.Info("loaded piece textures", "took", time.Since(start))

	return &textures, nil
}
