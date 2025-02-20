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
	style  material.EditorStyle
	border widget.Border
}

func NewPuzzleList(th *material.Theme) *PuzzleList {
	editor := &widget.Editor{
		ReadOnly: true,
	}
	return &PuzzleList{
		style: material.Editor(th, editor, "Lichess Puzzle URLs"),
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

func (l *PuzzleList) Set(puzzles []puzzle.Data) {
	var list strings.Builder
	for _, puzzleData := range puzzles {
		list.WriteString(puzzleData.URL())
		list.WriteRune('\n')
	}
	l.style.Editor.SetText(list.String())
}
