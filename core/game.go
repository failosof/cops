package core

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/notnil/chess"
)

type GameID [8]uint8

func (id GameID) String() string {
	return string(id[:])
}

func ParseGameID(str string) (id GameID) {
	copy(id[:], str)
	return
}

var urlRe = regexp.MustCompile(`lichess\.org/([a-zA-Z0-9]+)`)

func ParseGameIDFromURL(url string) (id GameID) {
	matches := urlRe.FindStringSubmatch(url)
	if len(matches) > 1 {
		return ParseGameID(matches[1])
	}
	return
}

type Move struct {
	From  chess.Square
	To    chess.Square
	Promo chess.PieceType
	Tags  chess.MoveTag
}

func (m Move) String() string {
	var str strings.Builder
	str.WriteString(m.From.String())
	str.WriteString(m.To.String())
	str.WriteString(m.Promo.String())
	return str.String()
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

func (m Moves) ContainMoves(moves []*chess.Move) bool {
	if len(moves) == 0 {
		return true
	}

	// todo: check move number
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

func (m Moves) ContainPosition(pos *chess.Position) bool {
	if pos == nil {
		return true
	}

	game := chess.NewGame(chess.UseNotation(chess.UCINotation{}))
	for _, move := range m {
		if err := game.MoveStr(move.String()); err != nil {
			return false
		}
	}

	return game.Position().Hash() == pos.Hash()
}

type GamesIndex map[GameID]Moves

func (i GamesIndex) Insert(id, moves string) error {
	parsedID := ParseGameID(id)
	parsedMoves, err := ParseMoves(moves)
	if err != nil {
		return fmt.Errorf("failed to parse moves: %w", err)
	}
	i[parsedID] = parsedMoves
	return nil
}

func (i GamesIndex) InsertFromChess(id GameID, game *chess.Game) {
	moves := make(Moves, len(game.Moves()))
	for i, move := range game.Moves() {
		moves[i] = MovesFromChess(move)
	}
	i[id] = moves
}
