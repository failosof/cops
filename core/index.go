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

type SearchType int8

const (
	MoveSequenceSearch SearchType = iota
	PositionSearch
)

func (t SearchType) String() string {
	switch t {
	case MoveSequenceSearch:
		return "By moves"
	case PositionSearch:
		return "By position"
	default:
		panic("unreachable")
	}
}

type Turn int8

const (
	EitherTurn Turn = iota
	WhiteTurn
	BlackTurn
)

func (t Turn) String() string {
	switch t {
	case WhiteTurn:
		return "White"
	case BlackTurn:
		return "Black"
	case EitherTurn:
		fallthrough
	default:
		return "Either"
	}
}

func (t Turn) ToChess() chess.Color {
	switch t {
	case WhiteTurn:
		return chess.White
	case BlackTurn:
		return chess.Black
	case EitherTurn:
		fallthrough
	default:
		return chess.NoColor
	}
}

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

type finding struct {
	puzzle PuzzleData
	game   Game
}

func (s *Index) SearchPuzzles(
	chessGame *chess.Game,
	strategy SearchType,
	turn chess.Color,
	maxMoves uint8,
) []PuzzleData {
	opening, moves := s.SearchOpening(chessGame)
	if opening.Empty() {
		return nil
	}

	// fast path
	if len(moves) == 0 {
		results := make([]PuzzleData, 0, 1000)
		for puzzle := range s.Puzzles.Filter(opening.Tag(), turn, maxMoves) {
			if _, ok := s.Games[puzzle.GameID]; ok {
				results = append(results, puzzle)
			}
		}
		return results
	}

	// slow path
	var wg sync.WaitGroup
	findingsCh := make(chan finding)
	puzzlesCh := make(chan PuzzleData)
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

	position := chessGame.Position()
	threads := runtime.NumCPU()
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for found := range findingsCh {
				var matches bool
				switch strategy {
				case MoveSequenceSearch:
					matches = found.game.ContainsMoves(moves)
				case PositionSearch:
					matches = found.game.ContainsPosition(position)
				}
				if matches {
					puzzlesCh <- found.puzzle
				}
			}
		}()
	}

	results := make([]PuzzleData, 0, 1000)
	for puzzle := range puzzlesCh {
		results = append(results, puzzle)
	}
	return results
}
