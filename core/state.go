package core

import (
	"context"
	"iter"
	"log/slog"
	"time"

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

func (s *State) SearchPuzzles(
	ctx context.Context,
	chessGame *chess.Game,
	turn chess.Color,
	maxMoves uint8,
) iter.Seq[puzzle.Data] {
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
		newGameIDs := make([]string, 0, game.MaxExportNumber)
		unsentPuzzles := make([]puzzle.Data, 0, game.MaxExportNumber)
		unusedLeftoverMovesSlice := make([][]*chess.Move, 0, game.MaxExportNumber)

		for foundPuzzle := range s.puzzles.Search(openingName.Tag(), turn, maxMoves) {
			moves := s.games.Search(foundPuzzle.GameID)
			if moves.Empty() {
				slog.Debug("puzzle game not indexed", "puzzle", foundPuzzle.ID, "game", foundPuzzle.GameID)

				newGameIDs = append(newGameIDs, foundPuzzle.GameID.String())
				unsentPuzzles = append(unsentPuzzles, foundPuzzle)
				unusedLeftoverMovesSlice = append(unusedLeftoverMovesSlice, leftoverMoves)

				if len(newGameIDs) == cap(newGameIDs) {
					if s.processNewGames(ctx, newGameIDs, unsentPuzzles, unusedLeftoverMovesSlice, yield) {
						return
					}

					newGameIDs = make([]string, 0, game.MaxExportNumber)
					unsentPuzzles = make([]puzzle.Data, 0, game.MaxExportNumber)
					unusedLeftoverMovesSlice = make([][]*chess.Move, 0, game.MaxExportNumber)
				}
			} else if moves.Contain(leftoverMoves) {
				if !yield(foundPuzzle) {
					return
				}
			}
		}

		if len(newGameIDs) > 0 {
			if s.processNewGames(ctx, newGameIDs, unsentPuzzles, unusedLeftoverMovesSlice, yield) {
				return
			}
		}
	}
}

func (s *State) processNewGames(
	ctx context.Context,
	gameIDs []string,
	puzzles []puzzle.Data,
	leftoverMovesSlice [][]*chess.Move,
	yield func(puzzle.Data) bool,
) bool {
	slog.Debug("exporting games from lichess", "count", len(gameIDs))
	start := time.Now()
	games, err := game.Export(ctx, gameIDs)
	if err != nil {
		slog.Error("failed to export games", "count", len(gameIDs), "err", err)
		return true
	}
	slog.Debug("exported", "took", time.Since(start))

	for _, chessGame := range games {
		exportedGameID := game.IDFromString(chessGame.GetTagPair("GameId").Value)
		s.games.InsertFromChess(exportedGameID, chessGame)

		for i, newGameID := range gameIDs {
			if newGameID == exportedGameID.String() {
				foundPuzzle := puzzles[i]
				leftoverMoves := leftoverMovesSlice[i]
				gameMoves := s.games.Search(exportedGameID)

				if !gameMoves.Empty() && gameMoves.Contain(leftoverMoves) {
					if !yield(foundPuzzle) {
						return true
					}
				}

				break
			}
		}
	}

	return false
}
