package ui

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/failosof/cops/core"
	"github.com/failosof/cops/ui/board"
	"github.com/failosof/cops/util"
)

type Window struct {
	state            *core.State
	window           *app.Window
	chessBoardConfig chessboard.Config

	padding unit.Dp

	openingFamily    *OpeningNamePart
	openingVariation *OpeningNamePart
	board            *chessboard.Widget
	controls         *BoardControls

	movesNumberSelector *MovesCountSelector
	turnSelector        *TurnSelector
	puzzleList          *PuzzleList
	searchButton        *IconButton
}

func NewWindow(state *core.State, chessRes core.ChessResources) (*Window, error) {
	config, err := chessboard.NewConfig(chessRes.BackgroundFile, chessRes.PiecesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to config chess board widget: %w", err)
	}
	config.ShowHints = true
	config.ShowLastMove = true

	return &Window{
		state:            state,
		window:           new(app.Window),
		chessBoardConfig: config,
		padding:          unit.Dp(3),
	}, nil
}

func (w *Window) Show(ctx context.Context) {
	w.window.Option(app.Title("Chess Opening Puzzle Search"))
	w.window.Option(app.MinSize(unit.Dp(1020), unit.Dp(640)))
	w.window.Option(app.MaxSize(unit.Dp(1020), unit.Dp(640)))

	th := material.NewTheme()

	w.openingFamily = NewOpeningNamePart(th, "Family")
	w.openingVariation = NewOpeningNamePart(th, "Variation")
	w.board = chessboard.NewWidget(th, w.chessBoardConfig)
	w.controls = NewBoardControls(th)

	w.movesNumberSelector = NewMovesNumberSelector(th, 1, 40)
	w.turnSelector = NewTurnSelector(th)
	w.puzzleList = NewPuzzleList(th)
	w.searchButton = NewIconButton(th, SearchIcon, util.GreenColor)

	go func() {
		if err := w.update(ctx); err != nil {
			slog.Error("main window update", "err", err)
			os.Exit(1)
		}
		os.Exit(0)
	}()

	app.Main()
}

func (w *Window) update(ctx context.Context) error {
	var ops op.Ops
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		switch e := w.window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, Pad(w.padding, w.layoutMainPane)),
			)
			w.controls
			e.Frame(gtx.Ops)
		}
	}
}

func (w *Window) layoutMainPane(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Flexed(1, w.layoutBoardPane),
		layout.Flexed(1, w.layoutSearchPane),
	)
}

func (w *Window) layoutBoardPane(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
		layout.Rigid(Pad(w.padding, w.openingFamily.Layout)),
		layout.Rigid(Pad(w.padding, w.openingVariation.Layout)),
		layout.Flexed(1, Pad(w.padding, func(gtx layout.Context) layout.Dimensions {
			return widget.Border{
				Color:        util.BlackColor,
				CornerRadius: unit.Dp(1),
				Width:        unit.Dp(1),
			}.Layout(gtx, w.board.Layout)
		})),
		layout.Rigid(Pad(w.padding, w.controls.Layout)),
	)
}

func (w *Window) layoutSearchPane(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
		layout.Rigid(PadSides(w.padding, w.movesNumberSelector.Layout)),
		layout.Rigid(Pad(w.padding, w.turnSelector.Layout)),
		layout.Flexed(1, Pad(w.padding, w.puzzleList.Layout)),
		layout.Rigid(Pad(w.padding, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Flexed(1, w.searchButton.Layout),
			)
		})),
	)
}
