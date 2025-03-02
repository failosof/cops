package core

import (
	"encoding/binary"
	"fmt"
	"iter"
	"strconv"
	"strings"
	"unsafe"

	"github.com/notnil/chess"
)

type PuzzleID [5]uint8

func ParsePuzzleID(s string) (id PuzzleID) {
	copy(id[:], s)
	return
}

func (id PuzzleID) String() string {
	return string(id[:])
}

type PuzzleData struct {
	Move   uint8
	Turn   chess.Color
	ID     PuzzleID
	GameID GameID
}

func NewPuzzleData(id, gameURL, fen string) (d PuzzleData, err error) {
	fenParts := strings.Split(fen, " ")
	if len(fenParts) != 6 {
		err = fmt.Errorf("invalid fen format")
		return
	}

	d.Move, err = MoveNumber(fenParts)
	d.Turn, err = PlayingTurn(fenParts)

	// puzzle saved position is one ply behind
	// thus it has an inverse turn encoded
	d.Turn = d.Turn.Other()

	d.ID = ParsePuzzleID(id)
	d.GameID = ParseGameIDFromURL(gameURL)

	return
}

func (d PuzzleData) URL() (url string) {
	url = "https://lichess.org/training/" + d.ID.String()
	return
}

func (d PuzzleData) GobEncode() (out []byte, err error) {
	out = make([]byte, unsafe.Sizeof(d))
	_, err = binary.Encode(out, binary.LittleEndian, d)
	if err != nil {
		err = fmt.Errorf("puzzle encode: %w", err)
	}
	return
}

func (d *PuzzleData) GobDecode(data []byte) (err error) {
	_, err = binary.Decode(data, binary.LittleEndian, d)
	if err != nil {
		err = fmt.Errorf("puzzle decode: %w", err)
	}
	return
}

type PuzzlesIndex map[string][]PuzzleData

func (i PuzzlesIndex) Insert(puzzleID, fen, gameURL, openingTags string) error {
	puzzle, err := NewPuzzleData(puzzleID, gameURL, fen)
	if err != nil {
		return fmt.Errorf("failed to parse puzzle: %w", err)
	}

	tags := strings.Split(openingTags, " ")
	for _, tag := range tags {
		i[tag] = append(i[tag], puzzle)
	}

	return nil
}

func (i PuzzlesIndex) Search(openingTag string, side chess.Color, maxMoves uint8) iter.Seq[PuzzleData] {
	return func(yield func(PuzzleData) bool) {
		for _, puzzle := range i[openingTag] {
			if side == chess.NoColor || puzzle.Turn == side {
				if puzzle.Move <= maxMoves {
					if !yield(puzzle) {
						return
					}
				}
			}
		}
	}
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
