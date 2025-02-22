package util_test

import (
	"strings"
	"testing"

	"github.com/failosof/cops/util"
	"github.com/notnil/chess"
)

func TestGameHasMoves(t *testing.T) {
	pgn := `[Event "Rated rapid game"]
[Site "https://lichess.org/dfUljrOi"]
[Date "2019.11.13"]
[White "HANG_LOOSE"]
[Black "lanetica"]
[Result "1-0"]
[GameId "dfUljrOi"]
[UTCDate "2019.11.13"]
[UTCTime "15:03:37"]
[WhiteElo "1632"]
[BlackElo "1648"]
[WhiteRatingDiff "+8"]
[BlackRatingDiff "-7"]
[Variant "Standard"]
[TimeControl "300+5"]
[ECO "D06"]
[Opening "Queen's Gambit Declined: Marshall Defense"]
[Termination "Time forfeit"]

1. d4 d5 2. c4 Nf6 3. Nc3 Bf5 4. cxd5 Nxd5 5. Qb3 Nxc3 6. bxc3 b6 
7. d5 e6 8. c4 exd5 9. cxd5 Be4 10. Qa4+ Nd7 11. Qxe4+ Be7 12. Ba3 Nf6 
13. Qa4+ Nd7 14. Bxe7 Qxe7 15. Rc1 1-0`
	opt, err := chess.PGN(strings.NewReader(pgn))
	if err != nil {
		t.Fail()
	}
	game := chess.NewGame(opt)
	moves := game.Moves()[4:18]
	has := util.GameHasMoves(game, moves)
	if !has {
		t.Fail()
	}
}
