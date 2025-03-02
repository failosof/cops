package core

import (
	"crypto/md5"
	"fmt"

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
