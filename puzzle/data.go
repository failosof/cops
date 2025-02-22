package puzzle

import (
	"encoding/binary"
	"fmt"
	"strings"
	"unsafe"

	"github.com/failosof/cops/game"
	"github.com/failosof/cops/util"
	"github.com/notnil/chess"
)

const DatabaseURL = "https://database.lichess.org/lichess_db_puzzle.csv.zst"

type ID [5]uint8

func (id ID) String() string {
	return string(id[:])
}

type Data struct {
	Move   uint8
	Turn   chess.Color
	ID     ID
	GameID game.ID
}

func NewData(id, gameURL, fen string) (d Data, err error) {
	fenParts := strings.Split(fen, " ")
	if len(fenParts) != 6 {
		err = fmt.Errorf("invalid fen format")
		return
	}

	d.Move, err = util.MoveNumber(fenParts)
	d.Turn, err = util.PlayingTurn(fenParts)

	// puzzle saved position is one ply behind
	// thus it has an inverse turn encoded
	d.Turn = d.Turn.Other()

	copy(d.ID[:], id)
	d.GameID = game.IDFromURL(gameURL)

	return
}

func (d Data) URL() (url string) {
	url = "https://lichess.org/training/" + d.ID.String()
	return
}

func (d Data) GobEncode() (out []byte, err error) {
	out = make([]byte, unsafe.Sizeof(d))
	_, err = binary.Encode(out, binary.LittleEndian, d)
	if err != nil {
		err = fmt.Errorf("puzzle encode: %w", err)
	}
	return
}

func (d *Data) GobDecode(data []byte) (err error) {
	_, err = binary.Decode(data, binary.LittleEndian, d)
	if err != nil {
		err = fmt.Errorf("puzzle decode: %w", err)
	}
	return
}
