package ui

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/failosof/cops/core"
	"github.com/notnil/chess"
	"golang.org/x/exp/shiny/materialdesign/icons"
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

type Icon []byte

var (
	ResetIcon    Icon = icons.NavigationRefresh
	FlipIcon     Icon = icons.NotificationSync
	BackwardIcon Icon = icons.NavigationArrowBack
	ForwardIcon  Icon = icons.NavigationArrowForward
	SearchIcon   Icon = icons.ActionSearch
)

type IconButton struct {
	color  color.NRGBA
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
		color:  color,
		icon:   icon,
		button: button,
		style:  style,
	}
}

func (b *IconButton) Layout(gtx layout.Context) layout.Dimensions {
	if gtx.Enabled() {
		b.style.Background = b.color
	} else {
		b.style.Background = Disabled(b.color)
	}
	return b.style.Layout(gtx, Pad(unit.Dp(5), func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return b.icon.Layout(gtx, WhiteColor)
			}),
		)
	}))
}

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
		reset:    NewIconButton(th, ResetIcon, RedColor),
		backward: NewIconButton(th, BackwardIcon, GrayColor),
		forward:  NewIconButton(th, ForwardIcon, GrayColor),
		flip:     NewIconButton(th, FlipIcon, BlueColor),
	}
}

func (c *BoardControls) Layout(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Flexed(1, c.reset.Layout),
		layout.Rigid(layout.Spacer{Width: c.padding}.Layout),
		layout.Flexed(1, c.backward.Layout),
		//layout.Rigid(layout.Spacer{Width: c.padding}.Layout),
		//layout.Flexed(1, c.forward.Layout),
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
			Color:        BlackColor,
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

func (l *PuzzleList) Add(puzzles []core.PuzzleData) {
	var list strings.Builder
	for _, puzzle := range puzzles {
		list.WriteString(puzzle.URL())
		list.WriteRune('\n')
	}
	l.editor.SetText(list.String())
}

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
	slider.Color = GrayColor
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
	if 1 <= moves && moves <= 40 {
		percent := float32(moves) / float32(s.max-s.min)
		s.slider.Float.Value = percent
	} else {
		s.slider.Float.Value = 0
	}
}

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
