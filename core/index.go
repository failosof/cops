package core

import (
	"log/slog"
	"path/filepath"
	"runtime"
	"sync"
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
	for i := len(positions) - 1; i > 0; i-- {
		pos := PositionFromChess(positions[i]).Hash()
		if name, ok := s.Openings[pos]; ok {
			found = name
			leftover = game.Moves()[i:]
			return
		}
	}
	return
}

func (s *Index) SearchPuzzles(chessGame *chess.Game, turn chess.Color, maxMoves uint8) []PuzzleData {
	opening, moves := s.SearchOpening(chessGame)
	if opening.Empty() {
		return nil
	}

	type finding struct {
		puzzle PuzzleData
		game   Game
	}

	threads := runtime.NumCPU()
	position := chessGame.Position()
	findingsCh := make(chan finding)
	puzzlesCh := make(chan PuzzleData)
	var wg sync.WaitGroup
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for found := range findingsCh {
				if len(moves) == 0 || found.game.ContainsPosition(position) {
					puzzlesCh <- found.puzzle
				}
			}
		}()
	}

	go func() {
		halfMoveNum := len(chessGame.Moves())
		maxMoves += uint8(halfMoveNum / 2) // offset from search position

		for puzzle := range s.Puzzles.Filter(opening.Tag(), turn, maxMoves) {
			if game, ok := s.Games[puzzle.GameID]; ok {
				findingsCh <- finding{
					puzzle: puzzle,
					game:   game,
				}
			}
		}

		close(findingsCh)
		wg.Wait()
		close(puzzlesCh)
	}()

	results := make([]PuzzleData, 0, 1000)
	for puzzle := range puzzlesCh {
		results = append(results, puzzle)
	}

	return results
}
