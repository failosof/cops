package core

import (
	"context"
	"iter"
	"log/slog"

	"github.com/failosof/cops/lichess"
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

func (s *State) SearchOpening(game *chess.Game) (name opening.Name, leftover []*chess.Move) {
	util.Assert(s.openings != nil, "openings must be loaded")

	positions := game.Positions()
	for i := 1; i < len(positions); i++ {
		pos := util.PositionFromChess(positions[i])
		candidate := s.openings.Search(pos)
		if candidate.Empty() {
			leftover = game.Moves()[i-1:]
			break
		}
		name = candidate
	}

	return
}

func (s *State) SearchPuzzles(
	ctx context.Context,
	game *chess.Game,
	turn chess.Color,
	minMoves, maxMoves uint8,
) iter.Seq[[]puzzle.Data] {
	util.Assert(s.puzzles != nil, "puzzles must be loaded")

	openingName, leftoverMoves := s.SearchOpening(game)
	if openingName.Empty() {
		return func(yield func([]puzzle.Data) bool) {
			return
		}
	}

	return func(yield func([]puzzle.Data) bool) {
		for puzzles, gameIDs := range s.puzzles.Search(openingName.Tag(), turn, minMoves, maxMoves) {
			games, err := lichess.ExportGames(ctx, gameIDs)
			if err != nil {
				slog.Error("failed to export games", "ids", gameIDs, "err", err)
				return
			}

			foundPuzzles := make([]puzzle.Data, 0, len(puzzles))
			for _, game := range games {
				if util.GameHasMoves(game, leftoverMoves) {
					foundGameID := game.GetTagPair("GameId").Value
					if len(foundGameID) > 0 {
						for i, gameID := range gameIDs {
							if gameID == foundGameID {
								foundPuzzles = append(foundPuzzles, puzzles[i])
								break
							}
						}
					}
				}
			}

			if len(foundPuzzles) > 0 {
				if !yield(foundPuzzles) {
					return
				}
			}
		}
	}
}
