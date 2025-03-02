package ui

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/failosof/cops/core"
)

type OpeningName struct {
	padding         unit.Dp
	name            core.OpeningName
	border          *widget.Border
	familyEditor    *widget.Editor
	variationEditor *widget.Editor
	familyStyle     material.EditorStyle
	variationStyle  material.EditorStyle
}

func NewOpeningName(th *material.Theme) *OpeningName {
	familyEditor := widget.Editor{
		SingleLine: true,
		ReadOnly:   true,
	}
	variationEditor := widget.Editor{
		SingleLine: true,
		ReadOnly:   true,
	}

	return &OpeningName{
		padding: unit.Dp(7),
		border: &widget.Border{
			Color:        BlackColor,
			CornerRadius: unit.Dp(1),
			Width:        unit.Dp(1),
		},
		familyEditor:    &familyEditor,
		variationEditor: &variationEditor,
		familyStyle:     material.Editor(th, &familyEditor, "Family"),
		variationStyle:  material.Editor(th, &variationEditor, "Variation"),
	}
}

func (w *OpeningName) Layout(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceEnd}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return w.border.Layout(gtx, Pad(w.padding, w.familyStyle.Layout))
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(3)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return w.border.Layout(gtx, Pad(w.padding, w.variationStyle.Layout))
		}),
	)
}

func (w *OpeningName) Set(name core.OpeningName) {
	w.name = name
	w.familyEditor.SetText(name.Family())
	w.variationEditor.SetText(name.Variation())
}
