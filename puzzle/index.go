package puzzle

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/failosof/cops/lichess"
	"github.com/klauspost/compress/zstd"
	"github.com/notnil/chess"
)

const AssumedPuzzleCount = 1_500_000 // puzzle db is ~4.5m records, only ~1m are from openings

type Index struct {
	Collection   []Data
	ByOpeningTag map[string][]int
}

func (i *Index) Insert(puzzleID, fen, gameURL, openingTags string) error {
	puzzle, err := NewData(puzzleID, gameURL, fen)
	if err != nil {
		return fmt.Errorf("failed to parse puzzle: %w", err)
	}

	id := len(i.Collection)
	i.Collection = append(i.Collection, puzzle)
	tags := strings.Split(openingTags, " ")
	for _, tag := range tags {
		i.ByOpeningTag[tag] = append(i.ByOpeningTag[tag], id)
	}

	return nil
}

func (i *Index) Search(openingTag string, side chess.Color, minMoves, maxMoves uint8) iter.Seq2[[]Data, []string] {
	return func(yield func([]Data, []string) bool) {
		for ids := range slices.Chunk(i.ByOpeningTag[openingTag], lichess.MaxExportGames) {
			puzzles := make([]Data, 0, len(ids))
			gameIDs := make([]string, 0, len(ids))
			for _, id := range ids {
				puzzle := i.Collection[id]
				if side == chess.NoColor || puzzle.Turn == side {
					if minMoves >= puzzle.Move && puzzle.Move <= maxMoves {
						puzzles = append(puzzles, puzzle)
						gameIDs = append(gameIDs, puzzle.GameID.String())
					}
				}
			}
			if len(puzzles) > 0 {
				if !yield(puzzles, gameIDs) {
					return
				}
			}
		}
	}
}

func CreateIndex(from string) (*Index, error) {
	index := Index{
		Collection:   make([]Data, 0, AssumedPuzzleCount),
		ByOpeningTag: make(map[string][]int, AssumedPuzzleCount),
	}

	filename := from
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %w", filename, err)
	}
	defer file.Close()

	decoder, err := zstd.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create zst decoder for %q: %w", filename, err)
	}
	defer decoder.Close()

	reader := csv.NewReader(decoder)
	reader.ReuseRecord = true

	var indexed, processed int
	for i := 0; ; i++ {
		lineNum := i + 1
		line, err := reader.Read()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return nil, fmt.Errorf("failed to read file %q line %d: %w", filename, lineNum, err)
			}
			break
		}

		if lineNum > 1 { // skip the header
			if len(line) != 10 {
				return nil, fmt.Errorf("file %q line %d: want 10 fields, have %d", filename, lineNum, len(line))
			}

			if len(line[9]) > 0 {
				if err := index.Insert(line[0], line[1], line[8], line[9]); err != nil {
					return nil, fmt.Errorf("file %q line %d: %w", filename, lineNum, err)
				}
				indexed++
			}
			processed++
		}
	}
	slog.Debug("created puzzles index", "from", from, "processed", processed, "indexed", indexed)

	return &index, nil
}
