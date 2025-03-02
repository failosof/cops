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

type OpeningsIndex map[[16]byte]OpeningName

func (i OpeningsIndex) Insert(name, moves string) error {
	pgn, err := chess.PGN(strings.NewReader(moves))
	if err != nil {
		return fmt.Errorf("failed to parse moves: %w", err)
	}

	game := chess.NewGame(pgn)
	if len(game.Moves()) == 0 {
		return fmt.Errorf("no moves parsed from %q", moves)
	}

	position := PositionFromChess(game.Position()).Hash()
	i[position] = ParseOpeningName(name)

	return nil
}
