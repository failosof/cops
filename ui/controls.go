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
		layout.Flexed(1, c.reset.Layout),
		layout.Rigid(layout.Spacer{Width: c.padding}.Layout),
		layout.Flexed(1, c.backward.Layout),
		layout.Rigid(layout.Spacer{Width: c.padding}.Layout),
		layout.Flexed(1, c.forward.Layout),
		layout.Rigid(layout.Spacer{Width: c.padding}.Layout),
		layout.Flexed(1, c.flip.Layout),
	)
}

func (c *BoardControls) ShouldReset(gtx layout.Context) bool {
	return c.reset.button.Clicked(gtx)
}

func (c *BoardControls) ShouldMoveBackward(gtx layout.Context) bool {
	return c.backward.button.Clicked(gtx)
}

func (c *BoardControls) ShouldMoveForward(gtx layout.Context) bool {
	return c.forward.button.Clicked(gtx)
}

func (c *BoardControls) ShouldFlip(gtx layout.Context) bool {
	return c.flip.button.Clicked(gtx)
}

func (c *BoardControls) Brighten() {
	c.reset.Brighten()
	c.backward.Brighten()
	c.forward.Brighten()
	c.flip.Brighten()
}

func (c *BoardControls) Fade() {
	c.reset.Fade()
	c.backward.Fade()
	c.forward.Fade()
	c.flip.Fade()
}
