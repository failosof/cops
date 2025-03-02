package core

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/notnil/chess"
	"golang.org/x/text/unicode/norm"
)

type OpeningName [2]string

func ParseOpeningName(s string) (n OpeningName) {
	parts := strings.Split(s, ":")
	n[0] = parts[0]
	if len(parts) > 1 {
		variation := strings.TrimSpace(parts[1])
		n[1] = strings.Split(variation, ",")[0]
	}
	return
}

func (n OpeningName) Empty() bool {
	return len(n[0]) == 0 && len(n[1]) == 0
}

func (n OpeningName) Family() string {
	return n[0]
}

func (n OpeningName) Variation() string {
	return n[1]
}

func (n OpeningName) String() string {
	var s strings.Builder
	s.WriteString(n[0])
	if len(n[1]) > 0 {
		s.WriteString(": ")
		s.WriteString(n[1])
	}
	return s.String()
}

func (n OpeningName) FamilyTag() string {
	return sanitizeOpeningName(n[0])
}

func (n OpeningName) VariationTag() string {
	return sanitizeOpeningName(n[1])
}

func (n OpeningName) Tag() string {
	var tag strings.Builder
	tag.WriteString(n.FamilyTag())
	if len(n[1]) > 0 {
		tag.WriteRune('_')
		tag.WriteString(n.VariationTag())
	}
	return tag.String()
}

var openingNameSanitizer = strings.NewReplacer("'", "", " ", "_")

func sanitizeOpeningName(s string) string {
	s = RemoveDiacritics(s)
	s = openingNameSanitizer.Replace(s)
	return s
}

func RemoveDiacritics(s string) string {
	t := norm.NFD.String(s)
	var sb strings.Builder
	for _, r := range t {
		if !unicode.IsMark(r) {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

type OpeningsIndex struct {
	Names       []OpeningName
	Positions   []Position
	PositionIDs map[[16]byte]int
}

func (i *OpeningsIndex) Insert(name, moves string) error {
	pgn, err := chess.PGN(strings.NewReader(moves))
	if err != nil {
		return fmt.Errorf("failed to parse moves: %w", err)
	}

	game := chess.NewGame(pgn)
	if len(game.Moves()) == 0 {
		return fmt.Errorf("no moves parsed from %q", moves)
	}

	id := len(i.Names)
	position := PositionFromChess(game.Position())
	hash := position.Hash()

	i.Names = append(i.Names, ParseOpeningName(name))
	i.Positions = append(i.Positions, position)
	i.PositionIDs[hash] = id

	return nil
}

func (i *OpeningsIndex) Search(position Position) (n OpeningName) {
	hash := position.Hash()
	id, ok := i.PositionIDs[hash]
	if ok {
		n = i.Names[id]
	}
	return
}

func (i *OpeningsIndex) Size() int {
	return len(i.Names)
}

//const IDPoolSize = 1300 // 1.3k ids per file

//func CreatePuzzleIndex(dir string) (*OpeningsIndex, error) {
//	size := len(filenames) * IDPoolSize
//	index := OpeningsIndex{
//		Names:       make([]Name, 0, size),
//		Positions:   make([]util.Position, 0, size),
//		PositionIDs: make(map[[16]byte]int, size),
//	}
//
//	var indexed int
//	for _, filename := range filenames {
//		filename = filepath.Join(dir, filename)
//		n, err := parseFile(filename, &index)
//		if err != nil {
//			return nil, err
//		}
//		indexed += n
//		slog.Debug("updated openings index", "from", filename, "indexed", n)
//	}
//	slog.Debug("created openings index", "from", dir, "indexed", indexed)
//
//	return &index, nil
//}
//
//func parseFile(filename string, index *OpeningsIndex) (n int, err error) {
//	file, err := os.Open(filename)
//	if err != nil {
//		err = fmt.Errorf("failed to open file %q: %w", filename, err)
//		return
//	}
//	defer file.Close()
//
//	reader := csv.NewReader(file)
//	reader.Comma = '\t'
//	reader.ReuseRecord = true
//
//	var line []string
//	for i := 0; ; i++ {
//		lineNum := i + 1
//		line, err = reader.Read()
//		if err != nil {
//			if !errors.Is(err, io.EOF) {
//				err = fmt.Errorf("failed to read file %q line %d: %w", filename, lineNum, err)
//			} else {
//				err = nil
//			}
//			return
//		}
//
//		if lineNum > 1 { // skip the header
//			if len(line) != 3 {
//				err = fmt.Errorf("file %q line %d: want 3 fields, have %d", filename, lineNum, len(line))
//				return
//			}
//			if err = index.Insert(line[1], line[2]); err != nil {
//				return
//			}
//			n++
//		}
//	}
//}
