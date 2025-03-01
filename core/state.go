package core

import (
	"context"
	"iter"
	"log/slog"

	"github.com/failosof/cops/game"
	"github.com/failosof/cops/opening"
	"github.com/failosof/cops/puzzle"
	"github.com/failosof/cops/util"
	"github.com/notnil/chess"
)

type State struct {
	openingRes OpeningResources
	gameRes    GameResources
	puzzleRes  PuzzleResources

	openings *opening.Index
	games    *game.Index
	puzzles  *puzzle.Index
}

func LoadState(
	ctx context.Context,
	openingRes OpeningResources,
	gameRes GameResources,
	puzzleRes PuzzleResources,
) (*State, error) {
	openings, err := LoadOpeningsIndex(ctx, openingRes.DatabaseDir, openingRes.IndexFile)
	if err != nil {
		slog.Warn("failed to load opening resources")
		return nil, err
	}

	games, err := LoadGamesIndex(ctx, gameRes.IndexFile)
	if err != nil {
		slog.Warn("failed to load games resources")
		return nil, err
	}

	for id, moves := range *games {
		if moves.Empty() {
			slog.Warn("apparently empty game indexed", "id", id)
		}
	}

	puzzles, err := LoadPuzzlesIndex(ctx, puzzleRes.DatabaseFile, puzzleRes.IndexFile)
	if err != nil {
		slog.Warn("failed to load puzzle resources")
		return nil, err
	}

	return &State{
		openingRes: openingRes,
		gameRes:    gameRes,
		puzzleRes:  puzzleRes,
		openings:   openings,
		games:      games,
		puzzles:    puzzles,
	}, nil
}

func (s *State) Save() {
	if err := s.games.Save(s.gameRes.IndexFile); err != nil {
		slog.Error("failed to save state", "err", err)
	}
	slog.Info("saved games index", "to", s.gameRes.IndexFile)
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

func (s *State) SearchPuzzles(chessGame *chess.Game, turn chess.Color, maxMoves uint8) iter.Seq[puzzle.Data] {
	util.Assert(s.games != nil, "games must be loaded")
	util.Assert(s.puzzles != nil, "puzzles must be loaded")

	openingName, leftoverMoves := s.SearchOpening(chessGame)
	if openingName.Empty() {
		return func(yield func(puzzle.Data) bool) {
			return
		}
	}

	slog.Info("searching puzzles", "opening", openingName.String(), "leftover", leftoverMoves)
	return func(yield func(puzzle.Data) bool) {
		var skipped int
		for foundPuzzle := range s.puzzles.Search(openingName.Tag(), turn, maxMoves) {
			moves := s.games.Search(foundPuzzle.GameID)
			if moves.Empty() {
				slog.Debug("puzzle game not indexed", "puzzle", foundPuzzle.ID, "game", foundPuzzle.GameID)
				skipped++
			} else if moves.Contain(leftoverMoves) {
				if !yield(foundPuzzle) {
					return
				}
			}
		}
		slog.Info("skipped puzzles", "opening", openingName, "moves", leftoverMoves, "count", skipped)
	}
}
