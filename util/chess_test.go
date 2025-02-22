package util_test

import (
	"strings"
	"testing"

	"github.com/failosof/cops/util"
	"github.com/notnil/chess"
)

func TestGameHasMoves(t *testing.T) {
	t.Run("has moves", func(t *testing.T) {
		pgn := `1. d4 d5 2. c4 Nf6 3. Nc3 Bf5 4. cxd5 Nxd5 5. Qb3 Nxc3 6. bxc3 b6 
7. d5 e6 8. c4 exd5 9. cxd5 Be4 10. Qa4+ Nd7 11. Qxe4+ Be7 12. Ba3 Nf6 
13. Qa4+ Nd7 14. Bxe7 Qxe7 15. Rc1 1-0`
		opt, err := chess.PGN(strings.NewReader(pgn))
		if err != nil {
			t.Fail()
		}
		game := chess.NewGame(opt)
		moves := getMoves(t)
		has := util.GameHasMoves(game, moves)
		if !has {
			t.Fail()
		}
	})

	t.Run("doesnt have moves", func(t *testing.T) {
		pgn := `1. d4 d5 2. c4 Nf6 3. cxd5 Nxd5 4. e4 Nb4 5. a3 N4a6 6. Bxa6 Nxa6 
7. d5 Nc5 8. Nf3 Nxe4 9. O-O Bg4 10. Qd4 Bxf3 11. gxf3 Nf6 12. Kh1 Nxd5 13. Nc3 0-1`
		opt, err := chess.PGN(strings.NewReader(pgn))
		if err != nil {
			t.Fail()
		}
		game := chess.NewGame(opt)
		moves := getMoves(t)
		has := util.GameHasMoves(game, moves)
		if has {
			t.Fail()
		}
	})
}

func getMoves(t *testing.T) []*chess.Move {
	t.Helper()

	pgn := `1. d4 d5 2. c4 Nf6 3. Nc3 Bf5 4. cxd5 Nxd5 5. Qb3 Nxc3 6. bxc3 b6 
7. d5 e6 8. c4 exd5 9. cxd5 Be4`
	opt, err := chess.PGN(strings.NewReader(pgn))
	if err != nil {
		t.Fail()
	}
	game := chess.NewGame(opt)
	return game.Moves()[4:18]
}
