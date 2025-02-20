package ui

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/notnil/chess"
)

type TurnSelector struct {
	group  *widget.Enum
	white  material.RadioButtonStyle
	black  material.RadioButtonStyle
	either material.RadioButtonStyle
}

func NewTurnSelector(th *material.Theme) *TurnSelector {
	group := new(widget.Enum)
	group.Value = "e"
	return &TurnSelector{
		group:  group,
		white:  material.RadioButton(th, group, "w", "White"),
		black:  material.RadioButton(th, group, "b", "Black"),
		either: material.RadioButton(th, group, "e", "Either"),
	}
}

func (s *TurnSelector) Layout(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Rigid(layout.Spacer{Width: unit.Dp(15)}.Layout),
		layout.Flexed(1, s.white.Layout),
		layout.Flexed(1, s.black.Layout),
		layout.Flexed(1, s.either.Layout),
	)
}

func (s *TurnSelector) Selected() (t chess.Color) {
	switch s.group.Value {
	case "w":
		t = chess.White
	case "b":
		t = chess.Black
	}
	return
}
