package ui

import (
	"strings"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/failosof/cops/puzzle"
	"github.com/failosof/cops/util"
)

type PuzzleList struct {
	editor *widget.Editor
	style  material.EditorStyle
	border widget.Border
}

func NewPuzzleList(th *material.Theme) *PuzzleList {
	editor := &widget.Editor{
		ReadOnly: true,
	}
	return &PuzzleList{
		editor: editor,
		style:  material.Editor(th, editor, "Lichess Puzzle URLs"),
		border: widget.Border{
			Color:        util.BlackColor,
			CornerRadius: unit.Dp(1),
			Width:        unit.Dp(1),
		},
	}
}

func (l *PuzzleList) Layout(gtx layout.Context) layout.Dimensions {
	return l.border.Layout(gtx, Pad(unit.Dp(7), l.style.Layout))
}

func (l *PuzzleList) Clear() {
	l.editor.SetText("")
}

func (l *PuzzleList) Append(puzzle puzzle.Data) {
	var list strings.Builder
	list.WriteString(l.editor.Text())
	list.WriteString(puzzle.URL())
	list.WriteRune('\n')
	l.editor.SetText(list.String())
}
