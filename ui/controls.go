package ui

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/failosof/cops/util"
)

type BoardControls struct {
	padding  unit.Dp
	reset    *IconButton
	backward *IconButton
	forward  *IconButton
	flip     *IconButton
}

func NewBoardControls(th *material.Theme) *BoardControls {
	return &BoardControls{
		padding:  unit.Dp(5),
		reset:    NewIconButton(th, ResetIcon, util.RedColor),
		backward: NewIconButton(th, BackwardIcon, util.GrayColor),
		forward:  NewIconButton(th, ForwardIcon, util.GrayColor),
		flip:     NewIconButton(th, FlipIcon, util.BlueColor),
	}
}

func (c *BoardControls) Layout(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Flexed(1, PadRight(c.padding, c.reset.Layout)),
		layout.Flexed(1, PadRight(c.padding, c.backward.Layout)),
		layout.Flexed(1, PadRight(c.padding, c.forward.Layout)),
		layout.Flexed(1, c.flip.Layout),
	)
}
