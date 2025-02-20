package ui

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

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

	opening  *OpeningName
	board    *chessboard.Widget
	controls *BoardControls

	moves   *MovesCountSelector
	turn    *TurnSelector
	puzzles *PuzzleList
	search  *IconButton
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
	w.window.Option(app.MinSize(unit.Dp(820), unit.Dp(620)))
	w.window.Option(app.MaxSize(unit.Dp(820), unit.Dp(620)))

	th := material.NewTheme()

	w.opening = NewOpeningName(th)
	w.board = chessboard.NewWidget(th, w.chessBoardConfig)
	w.controls = NewBoardControls(th)

	w.moves = NewMovesNumberSelector(th, 1, 40)
	w.turn = NewTurnSelector(th)
	w.puzzles = NewPuzzleList(th)
	w.search = NewIconButton(th, SearchIcon, util.GreenColor)

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
			w.handleControls(gtx)
			w.handleBoard(gtx)
			w.handleSearch(gtx)
			w.layoutWidgets(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

func (w *Window) handleControls(gtx layout.Context) {
	switch {
	case w.controls.ShouldReset(gtx):
		w.board.Reset()
	case w.controls.ShouldMoveBackward(gtx):
		w.board.MoveBackward()
	//case w.controls.ShouldMoveForward(gtx):
	//	w.board.MoveForward()
	case w.controls.ShouldFlip(gtx):
		w.board.Flip()
	default:
		return // do not refresh the screen
	}
	redraw(gtx)
}

func (w *Window) handleBoard(gtx layout.Context) {
	//if w.board.PositionChanged() {
	game := w.board.Game()
	openingName, _ := w.state.SearchOpening(&game)
	w.opening.Set(openingName)
	redraw(gtx)
	//}
}

func (w *Window) handleSearch(gtx layout.Context) {
	if w.search.button.Clicked(gtx) {
		openingName := w.opening.name
		game := w.board.Game()
		minMoves := uint8(len(game.Moves()))
		maxMoves := w.moves.Selected()
		if !openingName.Empty() && minMoves > 0 && maxMoves > 0 {
			w.puzzles.Clear()
			turn := w.turn.Selected()
			puzzles := w.state.SearchPuzzles(openingName, turn, minMoves, maxMoves)
			w.puzzles.Set(puzzles)
			redraw(gtx)
		}
	}
}

func (w *Window) layoutWidgets(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
		layout.Flexed(1, Pad(w.padding, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Flexed(3, w.layoutBoardPane),
				layout.Flexed(2, w.layoutSearchPane),
			)
		})),
	)
}

func (w *Window) layoutBoardPane(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
		layout.Rigid(Pad(w.padding, w.opening.Layout)),
		layout.Flexed(1, Pad(w.padding, func(gtx layout.Context) layout.Dimensions {
			return widget.Border{
				Color:        util.BlackColor,
				CornerRadius: unit.Dp(1),
				Width:        unit.Dp(1),
			}.Layout(gtx, w.board.Layout)
		})),
		layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
		layout.Rigid(Pad(w.padding, w.controls.Layout)),
	)
}

func (w *Window) layoutSearchPane(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
		layout.Flexed(1, Pad(w.padding, w.puzzles.Layout)),
		layout.Rigid(PadSides(w.padding, w.moves.Layout)),
		layout.Rigid(PadSides(w.padding, w.turn.Layout)),
		layout.Rigid(Pad(w.padding, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Flexed(1, w.search.Layout),
			)
		})),
	)
}

func redraw(gtx layout.Context) {
	gtx.Execute(op.InvalidateCmd{At: gtx.Now.Add(time.Second / 25)})
}
