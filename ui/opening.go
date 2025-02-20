package ui

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/failosof/cops/util"
)

type OpeningNamePart struct {
	border widget.Border
	style  material.EditorStyle
}

func NewOpeningNamePart(th *material.Theme, hint string) *OpeningNamePart {
	editor := widget.Editor{
		SingleLine: true,
		ReadOnly:   true,
	}
	return &OpeningNamePart{
		border: widget.Border{
			Color:        util.BlackColor,
			CornerRadius: unit.Dp(1),
			Width:        unit.Dp(1),
		},
		style: material.Editor(th, &editor, hint),
	}
}

func (p *OpeningNamePart) Layout(gtx layout.Context) layout.Dimensions {
	return p.border.Layout(gtx, Pad(unit.Dp(7), p.style.Layout))
}

func (p *OpeningNamePart) Set(text string) {
	p.style.Editor.SetText(text)
}
