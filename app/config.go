package app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/failosof/cops/util"
	"github.com/notnil/chess"
)

type Config struct {
	Game     *chess.Game
	Position util.Position
	Turn     chess.Color
	MinMoves uint8
	MaxMoves uint8
}

func ParseConfig(args []string) (c Config, err error) {
	pgn, err := chess.PGN(strings.NewReader(args[1]))
	if err != nil {
		err = fmt.Errorf("bad pgn: %w", err)
		return
	}

	c.Game = chess.NewGame(pgn)
	c.Position = util.PositionFromChess(c.Game.Position())

	switch strings.ToLower(args[2]) {
	case "w", "white":
		c.Turn = chess.White
	case "b", "black":
		c.Turn = chess.Black
	case "n", "no":
		c.Turn = chess.NoColor
	default:
		err = fmt.Errorf("bad turn value: should be w(hite), b(lack), or n(o)")
		return
	}

	val, err := strconv.ParseUint(args[3], 10, 8)
	if err != nil {
		err = fmt.Errorf("bad min moves value: %w", err)
		return
	}
	c.MinMoves = uint8(val)

	val, err = strconv.ParseUint(args[4], 10, 8)
	if err != nil {
		err = fmt.Errorf("bad max moves value: %w", err)
		return
	}
	c.MaxMoves = uint8(val)

	return
}
