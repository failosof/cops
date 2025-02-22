package game

import (
	"regexp"
	"strings"

	"github.com/notnil/chess"
)

type ID [8]uint8

func (id ID) String() string {
	return string(id[:])
}

func IDFromString(str string) (id ID) {
	copy(id[:], str)
	return
}

var urlRe = regexp.MustCompile(`lichess\.org/([a-zA-Z0-9]+)`)

func IDFromURL(url string) (id ID) {
	matches := urlRe.FindStringSubmatch(url)
	if len(matches) > 1 {
		return IDFromString(matches[1])
	}
	return
}

type Move struct {
	From  chess.Square
	To    chess.Square
	Promo chess.PieceType
	Tags  chess.MoveTag
}

type Moves []Move

func (m Moves) Empty() bool {
	return len(m) == 0
}

var moveTags = [...]chess.MoveTag{
	chess.KingSideCastle,
	chess.QueenSideCastle,
	chess.Capture,
	chess.EnPassant,
	chess.Check,
}

func MovesFromChess(move *chess.Move) Move {
	var tags chess.MoveTag
	for _, tag := range moveTags {
		if move.HasTag(tag) {
			tags = tags | tag
		}
	}

	return Move{
		From:  move.S1(),
		To:    move.S2(),
		Promo: move.Promo(),
		Tags:  tags,
	}
}

func ParseMoves(moves string) (m Moves, err error) {
	pgn, err := chess.PGN(strings.NewReader(moves))
	if err != nil {
		return
	}

	game := chess.NewGame(pgn)
	m = make(Moves, len(game.Moves()))
	for i, move := range game.Moves() {
		m[i] = MovesFromChess(move)
	}

	return
}

func (m Moves) Contain(moves []*chess.Move) bool {
	if len(moves) == 0 {
		return true
	}

	var j int
	for i := range m {
		if m[i] == MovesFromChess(moves[j]) {
			j++
			if j == len(moves) {
				return true
			}
		} else if j > 0 {
			return false
		}
	}

	return false
}
