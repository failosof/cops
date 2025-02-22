package game

import (
	"fmt"

	"github.com/failosof/cops/util"
	"github.com/notnil/chess"
)

const IndexURL = "https://github.com/failosof/cops/raw/refs/heads/main/assets/index/games.index"

type Index map[ID]Moves

func (i *Index) Save(to string) error {
	if err := util.SaveBinary(to, &i); err != nil {
		return fmt.Errorf("failed to save games index: %w", err)
	}
	return nil
}

func (i *Index) Insert(id, moves string) error {
	parsedID := IDFromString(id)
	parsedMoves, err := ParseMoves(moves)
	if err != nil {
		return fmt.Errorf("failed to parse moves: %w", err)
	}
	(*i)[parsedID] = parsedMoves
	return nil
}

func (i *Index) InsertFromChess(id ID, game *chess.Game) {
	moves := make(Moves, len(game.Moves()))
	for i, move := range game.Moves() {
		moves[i] = MovesFromChess(move)
	}
	(*i)[id] = moves
}

func (i *Index) Search(id ID) Moves {
	return (*i)[id]
}

func (i *Index) Size() int {
	return len(*i)
}
