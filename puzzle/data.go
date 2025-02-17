package puzzle

import (
	"encoding/binary"
	"fmt"
	"strings"
	"unsafe"

	"github.com/failosof/cops/lichess"
	"github.com/failosof/cops/util"
	"github.com/notnil/chess"
)

type Data struct {
	Move   uint8
	Turn   chess.Color
	ID     [5]uint8
	GameID [8]uint8
}

func NewData(id, gameURL, fen string) (d Data, err error) {
	fenParts := strings.Split(fen, " ")
	if len(fenParts) != 6 {
		err = fmt.Errorf("invalid fen format")
		return
	}

	d.Move, err = util.MoveNumber(fenParts)
	d.Turn, err = util.PlayingTurn(fenParts)

	copy(d.ID[:], id)
	copy(d.GameID[:], lichess.GameID(gameURL))

	return
}

func (d Data) URL() (url string) {
	id := string(d.ID[:])
	url = lichess.Puzzle(id)
	return
}

func (d Data) GameURL() (url string) {
	id := string(d.GameID[:])
	url = lichess.Game(id)
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
