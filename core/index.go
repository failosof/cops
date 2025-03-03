package core

import (
	"iter"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/failosof/cops/resources"
	"github.com/notnil/chess"
)

type Index struct {
	Openings OpeningsIndex
	Games    GamesIndex
	Puzzles  PuzzlesIndex
}

func LoadIndex() (*Index, error) {
	start := time.Now()
	openings, err := resources.LoadIndex[OpeningsIndex](filepath.Join("indexes", "openings.index"))
	if err != nil {
		slog.Warn("failed to load openings index")
		return nil, err
	}
	slog.Info("loaded openings index", "size", len(openings), "took", time.Since(start))

	start = time.Now()
	games, err := resources.LoadIndex[GamesIndex](filepath.Join("indexes", "games.index"))
	if err != nil {
		slog.Warn("failed to load games index")
		return nil, err
	}
	slog.Info("loaded games index", "size", len(games), "took", time.Since(start))

	start = time.Now()
	puzzles, err := resources.LoadIndex[PuzzlesIndex](filepath.Join("indexes", "puzzles.index"))
	if err != nil {
		slog.Warn("failed to load puzzles index")
		return nil, err
	}
	slog.Info("loaded puzzles index", "size", len(puzzles), "took", time.Since(start))

	return &Index{
		Openings: openings,
		Games:    games,
		Puzzles:  puzzles,
	}, nil
}

func (s *Index) SearchOpening(game *chess.Game) (found OpeningName, leftover []*chess.Move) {
	positions := game.Positions()
	for i := 1; i < len(positions); i++ {
		pos := PositionFromChess(positions[i]).Hash()
		if name, ok := s.Openings[pos]; ok {
			found = name
		} else {
			leftover = game.Moves()[i-1:]
			break
		}
	}
	return
}

func (s *Index) SearchPuzzles(chessGame *chess.Game, turn chess.Color, maxMoves uint8) iter.Seq[PuzzleData] {
	opening, moves := s.SearchOpening(chessGame)
	if opening.Empty() {
		return func(yield func(PuzzleData) bool) {
			return
		}
	}

	// calculate offset from search position
	maxMoves += uint8(len(chessGame.Moves()) / 2)
	position := chessGame.Position()

	return func(yield func(PuzzleData) bool) {
		start := time.Now()

		var found, skip int
		for foundPuzzle := range s.Puzzles.Search(opening.Tag(), turn, maxMoves) {
			if game, ok := s.Games[foundPuzzle.GameID]; ok {
				if game.ContainPosition(position) {
					found++
					if !yield(foundPuzzle) {
						break
					}
				}
			} else {
				skip++
				slog.Warn("puzzle game not indexed", "puzzle", foundPuzzle.ID, "game", foundPuzzle.GameID)
			}
		}

		took := time.Since(start)
		slog.Info("puzzle search", "opening", opening, "moves", moves, "found", found, "skip", skip, "took", took)
	}
}
