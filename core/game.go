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

type Game []Move

var moveTags = [...]chess.MoveTag{
	chess.KingSideCastle,
	chess.QueenSideCastle,
	chess.Capture,
	chess.EnPassant,
	chess.Check,
}

func GameFromChess(move *chess.Move) Move {
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

func ParseGame(moves string) (g Game, err error) {
	pgn, err := chess.PGN(strings.NewReader(moves))
	if err != nil {
		return
	}

	game := chess.NewGame(pgn)
	g = make(Game, len(game.Moves()))
	for i, move := range game.Moves() {
		g[i] = GameFromChess(move)
	}

	return
}

func (g Game) Empty() bool {
	return len(g) == 0
}

func (g Game) ContainsMoves(moves []*chess.Move) bool {
	if len(moves) == 0 {
		return true
	}

	// todo: check move number
	var j int
	for i := range g {
		if g[i] == GameFromChess(moves[j]) {
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

func (g Game) ContainsPosition(pos *chess.Position) bool {
	if pos == nil {
		return true
	}

	game := chess.NewGame(chess.UseNotation(chess.UCINotation{}))
	for _, move := range g {
		if err := game.MoveStr(move.String()); err != nil {
			return false
		}
		if game.Position().Hash() == pos.Hash() {
			return true
		}
	}

	return false
}

type GamesIndex map[GameID]Game

func (i GamesIndex) Insert(id, moves string) error {
	parsedID := ParseGameID(id)
	parsedMoves, err := ParseGame(moves)
	if err != nil {
		return fmt.Errorf("failed to parse moves: %w", err)
	}
	i[parsedID] = parsedMoves
	return nil
}

func (i GamesIndex) InsertFromChess(id GameID, chessGame *chess.Game) {
	game := make(Game, len(chessGame.Moves()))
	for i, move := range chessGame.Moves() {
		game[i] = GameFromChess(move)
	}
	i[id] = game
}
