package ui

import (
	"context"
	"fmt"
	"image/color"
	"log/slog"
	"os"

	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/failosof/cops/core"
	"github.com/failosof/cops/ui/board"
	"github.com/failosof/cops/ui/board/util"
)

func DrawMainWindow(ctx context.Context, state *core.State, chessBoardRes core.ChessBoardResources) {
	go func() {
		if err := draw(ctx, state, chessBoardRes, new(app.Window)); err != nil {
			slog.Error("main window", "err", err)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	app.Main()
}

func draw(ctx context.Context, state *core.State, chessBoardRes core.ChessBoardResources, window *app.Window) error {
	window.Option(app.Title("Chess Opening Puzzle Search"))
	window.Option(app.MinSize(unit.Dp(1000), unit.Dp(600)))

	config, err := chessboard.NewConfig(chessBoardRes.BackgroundFile, chessBoardRes.PiecesDir)
	if err != nil {
		return fmt.Errorf("failed to config chess board widget: %w", err)
	}
	config.ShowHints = true
	config.ShowLastMove = true

	th := material.NewTheme()

	openingFamilyEditor := &widget.Editor{
		SingleLine: true,
		ReadOnly:   true,
	}
	openingVariationEditor := &widget.Editor{
		SingleLine: true,
		ReadOnly:   true,
	}
	resetPositionBtn := new(widget.Clickable)
	moveBackwardBtn := new(widget.Clickable)
	moveForwardBtn := new(widget.Clickable)
	flipBoardBtn := new(widget.Clickable)
	board := chessboard.NewWidget(th, config)

	float := new(widget.Float)
	radioButtonsGroup := new(widget.Enum)
	editor := new(widget.Editor)

	var ops op.Ops
	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			layout.UniformInset(unit.Dp(3)).Layout(gtx,
				func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return layout.UniformInset(unit.Dp(3)).Layout(gtx,
										func(gtx layout.Context) layout.Dimensions {
											e := material.Editor(th, openingFamilyEditor, "Sicilian Defense")
											border := widget.Border{
												Color:        util.BlackColor,
												CornerRadius: unit.Dp(1),
												Width:        unit.Dp(1),
											}
											return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
												return layout.UniformInset(unit.Dp(5)).Layout(gtx, e.Layout)
											})
										},
									)
								}),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return layout.UniformInset(unit.Dp(3)).Layout(gtx,
										func(gtx layout.Context) layout.Dimensions {
											e := material.Editor(th, openingVariationEditor, "Najdorf Variation")
											border := widget.Border{
												Color:        color.NRGBA{A: 0xff},
												CornerRadius: unit.Dp(1),
												Width:        unit.Dp(1),
											}
											return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
												return layout.UniformInset(unit.Dp(5)).Layout(gtx, e.Layout)
											})
										},
									)
								}),
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									return layout.UniformInset(unit.Dp(3)).Layout(gtx,
										func(gtx layout.Context) layout.Dimensions {
											return board.Layout(gtx)
										},
									)
								}),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
										layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
											btn := material.Button(th, resetPositionBtn, "RESET")
											btn.Background = util.RedColor
											btn.Font.Weight = font.Bold
											return layout.UniformInset(unit.Dp(3)).Layout(gtx, btn.Layout)
										}),
										layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
											btn := material.Button(th, moveBackwardBtn, "<")
											btn.Background = util.GrayColor
											btn.Font.Weight = font.Bold
											return layout.UniformInset(unit.Dp(3)).Layout(gtx, btn.Layout)
										}),
										layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
											btn := material.Button(th, moveForwardBtn, ">")
											btn.Background = util.GrayColor
											btn.Font.Weight = font.Bold
											return layout.UniformInset(unit.Dp(3)).Layout(gtx, btn.Layout)
										}),
										layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
											btn := material.Button(th, flipBoardBtn, "FLIP")
											btn.Background = util.BlueColor
											btn.Font.Weight = font.Bold
											return layout.UniformInset(unit.Dp(3)).Layout(gtx, btn.Layout)
										}),
									)
								}),
							)
						}),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
										layout.Flexed(1, material.Slider(th, float).Layout),
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											return layout.UniformInset(unit.Dp(8)).Layout(gtx,
												material.Body1(th, fmt.Sprintf("%.2f", float.Value)).Layout,
											)
										}),
									)
								}),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
										layout.Flexed(1, material.RadioButton(th, radioButtonsGroup, "r1", "WHITE").Layout),
										layout.Flexed(1, material.RadioButton(th, radioButtonsGroup, "r2", "W|B").Layout),
										layout.Flexed(1, material.RadioButton(th, radioButtonsGroup, "r3", "BLACK").Layout),
									)
								}),
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									e := material.Editor(th, editor, "Lichess Puzzle URLs")
									border := widget.Border{
										Color:        util.BlackColor,
										CornerRadius: unit.Dp(1),
										Width:        unit.Dp(1),
									}
									e.Editor.SetText("https://lichess.org/training/sfv0s2")
									return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
										return layout.UniformInset(unit.Dp(5)).Layout(gtx, e.Layout)
									})
								}),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
										layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
											btn := material.Button(th, resetPositionBtn, "SEARCH")
											btn.Background = util.GreenColor
											btn.Font.Weight = font.Bold
											return layout.UniformInset(unit.Dp(3)).Layout(gtx, btn.Layout)
										}),
									)
								}),
							)
						}),
					)
				},
			)
			e.Frame(gtx.Ops)
		}
	}
}
