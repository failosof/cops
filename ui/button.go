package ui

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/failosof/cops/util"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

type Icon []byte

var (
	ResetIcon    Icon = icons.NavigationRefresh
	FlipIcon     Icon = icons.NotificationSync
	BackwardIcon Icon = icons.NavigationArrowBack
	ForwardIcon  Icon = icons.NavigationArrowForward
	SearchIcon   Icon = icons.ActionSearch
)

type IconButton struct {
	icon   *widget.Icon
	button *widget.Clickable
	style  material.ButtonLayoutStyle
}

func NewIconButton(th *material.Theme, name Icon, color color.NRGBA) *IconButton {
	icon, _ := widget.NewIcon(name)
	button := new(widget.Clickable)
	style := material.ButtonLayout(th, button)
	style.Background = color
	return &IconButton{
		icon:   icon,
		button: button,
		style:  style,
	}
}

func (b *IconButton) Layout(gtx layout.Context) layout.Dimensions {
	return b.style.Layout(gtx, Pad(unit.Dp(5), func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return b.icon.Layout(gtx, util.WhiteColor)
			}),
		)
	}))
}
