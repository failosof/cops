package opening

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/failosof/cops/util"
	"github.com/notnil/chess"
)

const IDPoolSize = 1300 // 1.3k ids per file

type Index struct {
	Names       []Name
	Positions   []util.Position
	PositionIDs map[[16]byte]int
}

func (i *Index) Insert(name, moves string) error {
	pgn, err := chess.PGN(strings.NewReader(moves))
	if err != nil {
		return fmt.Errorf("failed to parse moves: %w", err)
	}

	game := chess.NewGame(pgn)
	if len(game.Moves()) == 0 {
		return fmt.Errorf("no moves parsed from %q", moves)
	}

	id := len(i.Names)
	position := util.PositionFromChess(game.Position())
	hash := position.Hash()

	i.Names = append(i.Names, ParseName(name))
	i.Positions = append(i.Positions, position)
	i.PositionIDs[hash] = id

	return nil
}

func (i *Index) Search(position util.Position) (n Name) {
	hash := position.Hash()
	id, ok := i.PositionIDs[hash]
	if ok {
		n = i.Names[id]
	}
	return
}

func (i *Index) Size() int {
	return len(i.Names)
}

func CreateIndex(dir string) (*Index, error) {
	size := len(filenames) * IDPoolSize
	index := Index{
		Names:       make([]Name, 0, size),
		Positions:   make([]util.Position, 0, size),
		PositionIDs: make(map[[16]byte]int, size),
	}

	var indexed int
	for _, filename := range filenames {
		filename = filepath.Join(dir, filename)
		n, err := parseFile(filename, &index)
		if err != nil {
			return nil, err
		}
		indexed += n
		slog.Debug("updated openings index", "from", filename, "indexed", n)
	}
	slog.Debug("created openings index", "from", dir, "indexed", indexed)

	return &index, nil
}

func parseFile(filename string, index *Index) (n int, err error) {
	file, err := os.Open(filename)
	if err != nil {
		err = fmt.Errorf("failed to open file %q: %w", filename, err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = '\t'
	reader.ReuseRecord = true

	var line []string
	for i := 0; ; i++ {
		lineNum := i + 1
		line, err = reader.Read()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				err = fmt.Errorf("failed to read file %q line %d: %w", filename, lineNum, err)
			} else {
				err = nil
			}
			return
		}

		if lineNum > 1 { // skip the header
			if len(line) != 3 {
				err = fmt.Errorf("file %q line %d: want 3 fields, have %d", filename, lineNum, len(line))
				return
			}
			if err = index.Insert(line[1], line[2]); err != nil {
				return
			}
			n++
		}
	}
}
