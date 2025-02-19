package core

import (
	"context"
	"log/slog"

	"github.com/failosof/cops/opening"
	"github.com/failosof/cops/puzzle"
	"github.com/failosof/cops/util"
	"github.com/notnil/chess"
)

type State struct {
	openings *opening.Index
	puzzles  *puzzle.Index
}

func LoadState(ctx context.Context, openingRes OpeningResources, puzzleRes PuzzleResources) (*State, error) {
	openings, err := LoadOpeningsIndex(ctx, openingRes.DatabaseDir, openingRes.IndexFile)
	if err != nil {
		slog.Warn("failed to load opening resources")
		return nil, err
	}

	puzzles, err := LoadPuzzlesIndex(ctx, puzzleRes.DatabaseFile, puzzleRes.IndexFile)
	if err != nil {
		slog.Warn("failed to load puzzle resources")
		return nil, err
	}

	return &State{
		openings: openings,
		puzzles:  puzzles,
	}, nil
}

func (s *State) SearchOpening(game *chess.Game) (name opening.Name, leftover []chess.Move) {
	util.Assert(s.openings != nil, "openings must be loaded")

	positions := game.Positions()
	i := len(positions) - 1
	var candidate opening.Name
	var pos util.Position
	for ; i >= 0; i-- {
		pos = util.PositionFromChess(positions[i])
		candidate = s.openings.Search(pos)
		if !candidate.Empty() {
			break
		}
	}
	if i == 0 {
		slog.Warn("no opening found")
		return
	}

	slog.Info("found opening", "family", candidate.Family(), "variation", candidate.Variation())
	name = candidate

	return
}

func (s *State) SearchPuzzles(
	openingName opening.Name,
	turn chess.Color,
	minMoves, maxMoves uint8,
) (puzzles []puzzle.Data) {
	util.Assert(s.puzzles != nil, "puzzles must be loaded")

	puzzles = s.puzzles.Search(openingName.Tag(), turn, minMoves, maxMoves)
	if len(puzzles) == 0 {
		slog.Warn("no puzzles found")
		return nil
	}
	slog.Info("found puzzles", "count", len(puzzles))

	// todo: filter out puzzles by moves after the opening

	return
}
