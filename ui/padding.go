package ui

import (
	"gioui.org/layout"
	"gioui.org/unit"
)

func Pad(padding unit.Dp, widget layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(padding).Layout(gtx, widget)
	}
}

func PadRight(padding unit.Dp, widget layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Right: padding}.Layout(gtx, widget)
	}
}

func PadSides(padding unit.Dp, widget layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Left: padding, Right: padding}.Layout(gtx, widget)
	}
}
