package ui

import (
	"fmt"
	"math"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/failosof/cops/util"
)

type MovesCountSelector struct {
	min, max uint8
	padding  unit.Dp
	slider   material.SliderStyle
	hint     material.LabelStyle
	count    material.LabelStyle
}

func NewMovesNumberSelector(th *material.Theme, min, max uint8) *MovesCountSelector {
	float := new(widget.Float)
	slider := material.Slider(th, float)
	slider.Color = util.GrayColor
	return &MovesCountSelector{
		min:     min,
		max:     max,
		padding: unit.Dp(8),
		slider:  slider,
		hint:    material.Body1(th, "Moves"),
		count:   material.Body1(th, "1"),
	}
}

func (s *MovesCountSelector) Layout(gtx layout.Context) layout.Dimensions {
	s.count.Text = fmt.Sprint(s.Selected())
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(Pad(s.padding, s.hint.Layout)),
		layout.Flexed(1, s.slider.Layout),
		layout.Rigid(Pad(s.padding, s.count.Layout)),
	)
}

func (s *MovesCountSelector) Selected() (res uint8) {
	percent := s.slider.Float.Value
	from := float32(s.max - s.min + 1)
	count := from * percent
	res = uint8(math.Ceil(float64(count)))
	return
}

func (s *MovesCountSelector) Set(moves uint8) {
	util.Assert(1 <= moves && moves <= 40, "moves number must be in [1, 40]")

	percent := float32(moves) / float32(s.max-s.min)
	s.slider.Float.Value = percent
}
