package ui

import (
	"fmt"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/failosof/cops/core"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"image/color"
	"math"
	"strconv"
)

type TextFieldOption int8

const (
	ReadOnly TextFieldOption = 1 << iota
	SingleLine
)

type TextField struct {
	padding unit.Dp
	border  *widget.Border
	editor  *widget.Editor
	style   material.EditorStyle
}

func NewTextField(th *material.Theme, hint string, options TextFieldOption) *TextField {
	editor := widget.Editor{
		ReadOnly:   options&ReadOnly != 0,
		SingleLine: options&SingleLine != 0,
	}
	return &TextField{
		padding: unit.Dp(7),
		border: &widget.Border{
			Color:        BlackColor,
			CornerRadius: unit.Dp(1),
			Width:        unit.Dp(1),
		},
		editor: &editor,
		style:  material.Editor(th, &editor, hint),
	}
}

func (w *TextField) SetText(text string) {
	w.editor.SetText(text)
}

func (w *TextField) Layout(gtx layout.Context) layout.Dimensions {
	return w.border.Layout(gtx, Pad(w.padding, w.style.Layout))
}

type OpeningName struct {
	family    *TextField
	variation *TextField
}

func NewOpeningName(th *material.Theme) *OpeningName {
	return &OpeningName{
		family:    NewTextField(th, "Family", ReadOnly|SingleLine),
		variation: NewTextField(th, "Variation", ReadOnly|SingleLine),
	}
}

func (w *OpeningName) Layout(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceEnd}.Layout(gtx,
		layout.Rigid(w.family.Layout),
		layout.Rigid(layout.Spacer{Height: unit.Dp(3)}.Layout),
		layout.Rigid(w.variation.Layout),
	)
}

func (w *OpeningName) Set(name core.OpeningName) {
	w.family.SetText(name.Family())
	w.variation.SetText(name.Variation())
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

type RangeSlider struct {
	min, max uint8
	padding  unit.Dp
	slider   material.SliderStyle
	hint     material.LabelStyle
	count    material.LabelStyle
}

func NewRangeSlider(th *material.Theme, hint string, min, max uint8) *RangeSlider {
	float := new(widget.Float)
	slider := material.Slider(th, float)
	slider.Color = GrayColor
	return &RangeSlider{
		min:     min,
		max:     max,
		padding: unit.Dp(8),
		slider:  slider,
		hint:    material.Body1(th, hint),
		count:   material.Body1(th, "1"),
	}
}

func (s *RangeSlider) Layout(gtx layout.Context) layout.Dimensions {
	s.count.Text = fmt.Sprint(s.Selected())
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(Pad(s.padding, s.hint.Layout)),
		layout.Flexed(1, s.slider.Layout),
		layout.Rigid(Pad(s.padding, s.count.Layout)),
	)
}

func (s *RangeSlider) Selected() (res uint8) {
	percent := s.slider.Float.Value
	from := float32(s.max - s.min + 1)
	count := from * percent
	res = uint8(math.Ceil(float64(count)))
	return
}

func (s *RangeSlider) Set(value uint8) {
	if s.min <= value && value <= s.max {
		percent := float32(value) / float32(s.max-s.min)
		s.slider.Float.Value = percent
	} else {
		s.slider.Float.Value = 0
	}
}

type OptionSelector[Option fmt.Stringer] struct {
	group   *widget.Enum
	options []Option
	styles  []material.RadioButtonStyle
	layouts []layout.FlexChild
}

func NewOptionSelector[Option fmt.Stringer](th *material.Theme, options []Option) *OptionSelector[Option] {
	group := new(widget.Enum)
	group.Value = "0"
	styles := make([]material.RadioButtonStyle, len(options))
	layouts := make([]layout.FlexChild, len(options))
	for i, option := range options {
		styles[i] = material.RadioButton(th, group, strconv.Itoa(i), option.String())
		layouts[i] = layout.Flexed(1, styles[i].Layout)
	}
	return &OptionSelector[Option]{
		group:   group,
		options: options,
		styles:  styles,
		layouts: layouts,
	}
}

func (w *OptionSelector[Option]) Layout(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, w.layouts...)
}

func (w *OptionSelector[Option]) Selected() (o Option) {
	if i, err := strconv.Atoi(w.group.Value); err == nil {
		if 0 <= i && i < len(w.options) {
			o = w.options[i]
		}
	}
	return
}
