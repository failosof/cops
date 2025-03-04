package chessboard

import (
	"image"
	"image/color"
	"image/draw"
	_ "image/png"
	"io"
	"time"

	"github.com/failosof/cops/ui/board/union"
)

type Coordinates int8

const (
	NoCoordinates Coordinates = iota
	InsideCoordinates
	OutsideCoordinates
	EachSquare
)

type Colors struct {
	Empty     color.NRGBA
	Hint      color.NRGBA
	Highlight color.NRGBA
	LastMove  color.NRGBA
	Primary   color.NRGBA
	Info      color.NRGBA
	Warning   color.NRGBA
	Danger    color.NRGBA
}

type Texture struct {
	Image *image.NRGBA
	Size  union.Size
}

func LoadTexture(file io.Reader) (t Texture, err error) {
	src, _, err := image.Decode(file)
	if err != nil {
		return
	}
	t.Image = image.NewNRGBA(image.Rect(0, 0, src.Bounds().Dx(), src.Bounds().Dy()))
	draw.Draw(t.Image, t.Image.Bounds(), src, src.Bounds().Min, draw.Src)
	t.Size = union.SizeFromMinPt(t.Image.Bounds().Max)
	return
}

type Config struct {
	ShowHints      bool
	ShowLastMove   bool
	Colors         Colors
	AnimationSpeed time.Duration
	Coordinates    Coordinates
	BoardTexture   Texture
	PieceTextures  []Texture
}

func NewConfig(board Texture, pieces []Texture, colors Colors) (c *Config) {
	c = new(Config)
	c.BoardTexture = board
	c.PieceTextures = pieces
	c.Colors = colors
	return
}
