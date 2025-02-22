package util

import (
	"crypto/md5"
	"fmt"
	"strconv"

	"github.com/notnil/chess"
)

type Position struct{ chess.Position }

func PositionFromChess(p *chess.Position) Position {
	return Position{*p}
}

func (p Position) GobEncode() (out []byte, err error) {
	if p.Board() == nil {
		return nil, nil
	}
	out, err = p.MarshalBinary()
	if err != nil {
		err = fmt.Errorf("position encode: %w", err)
	}
	return
}

func (p *Position) GobDecode(data []byte) (err error) {
	if len(data) == 0 {
		return nil
	}
	err = p.UnmarshalBinary(data)
	if err != nil {
		err = fmt.Errorf("position decode: %w", err)
	}
	return
}

func (p Position) Hash() [16]byte {
	// (p)oop moment
	// underlying package detects en-passant
	// square incorrectly, thus requires to
	// implement this bug prone hash function
	return md5.Sum([]byte(p.Board().String()))
}

func MoveNumber(fen []string) (n uint8, err error) {
	v, err := strconv.ParseUint(fen[5], 10, 8)
	if err != nil {
		return 0, fmt.Errorf("invalid move number: %v", err)
	}
	n = uint8(v)

	return
}

func PlayingTurn(fen []string) (c chess.Color, err error) {
	switch fen[1] {
	case "w":
		c = chess.White
	case "b":
		c = chess.Black
	default:
		err = fmt.Errorf("invalid playing side: %s", fen[1])
	}
	return
}

func GameHasMoves(game *chess.Game, moves []*chess.Move) bool {
	var i, j int
	gameMoves := game.Moves()
	for i < len(gameMoves) && j < len(moves) {
		sameMove := *gameMoves[i] == *moves[j]
		if j > 0 && !sameMove {
			return false
		}
		if sameMove {
			j++
		}
		i++
	}

	return j != 0
}
