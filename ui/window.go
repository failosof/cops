package ui

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync/atomic"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/failosof/cops/core"
	"github.com/failosof/cops/puzzle"
	"github.com/failosof/cops/ui/board"
	"github.com/failosof/cops/util"
	"github.com/notnil/chess"
)

type Window struct {
	index            *core.State
	window           *app.Window
	chessBoardConfig chessboard.Config

	padding unit.Dp

	opening       *OpeningName
	board         *chessboard.Widget
	boardControls *BoardControls

	moves   *MovesCountSelector
	turn    *TurnSelector
	puzzles *PuzzleList
	search  *IconButton
	cancel  *IconButton

	searching atomic.Bool
	loaded    atomic.Bool
	results   []puzzle.Data
}

func NewWindow(state *core.State, chessRes core.ChessResources) (*Window, error) {
	config, err := chessboard.NewConfig(chessRes.BackgroundFile, chessRes.PiecesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to config chess board widget: %w", err)
	}
	config.ShowHints = true
	config.ShowLastMove = true

	return &Window{
		index:            state,
		window:           new(app.Window),
		chessBoardConfig: config,
		padding:          unit.Dp(3),
	}, nil
}

func (w *Window) Show(ctx context.Context) {
	pgn := `1. d4 d5 2. c4 Nf6 3. Nc3 Bf5 4. cxd5 Nxd5 5. Qb3 Nxc3 6. bxc3 b6 7. d5 e6 8. c4 exd5 9. cxd5 Be4`
	opt, _ := chess.PGN(strings.NewReader(pgn))
	game := chess.NewGame(opt, chess.UseNotation(chess.UCINotation{}))

	w.window.Option(app.Title("Chess Opening Puzzle Search"))
	w.window.Option(app.MinSize(unit.Dp(820), unit.Dp(620)))
	w.window.Option(app.MaxSize(unit.Dp(820), unit.Dp(620)))

	th := material.NewTheme()

	w.opening = NewOpeningName(th)
	w.board = chessboard.NewWidget(th, w.chessBoardConfig)
	w.board.SetGame(game)
	w.boardControls = NewBoardControls(th)

	w.moves = NewMovesNumberSelector(th, 1, 40)
	w.moves.Set(12)
	w.turn = NewTurnSelector(th)
	w.turn.group.Value = "w"
	w.puzzles = NewPuzzleList(th)
	w.search = NewIconButton(th, SearchIcon, util.GreenColor)
	w.cancel = NewIconButton(th, CancelIcon, util.RedColor)

	go func() {
		if err := w.update(ctx); err != nil {
			slog.Error("main window update", "err", err)
			os.Exit(1)
		} else {
			os.Exit(0)
		}
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
			if w.searching.Load() {
				w.handleCancel(gtx)
			} else {
				w.handleControls(gtx)
				w.handleBoard(gtx)
				w.handleSearch(gtx)
			}
			w.layoutWidgets(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

func (w *Window) handleControls(gtx layout.Context) {
	switch {
	case w.boardControls.ShouldReset(gtx):
		w.board.Reset()
	case w.boardControls.ShouldMoveBackward(gtx):
		w.board.MoveBackward()
	//case w.boardControls.ShouldMoveForward(gtx):
	//	w.board.MoveForward()
	case w.boardControls.ShouldFlip(gtx):
		w.board.Flip()
	default:
		return // do not refresh the screen
	}
	redraw(gtx)
}

func (w *Window) handleBoard(gtx layout.Context) {
	//if w.board.PositionChanged() {
	game := w.board.Game()
	openingName, _ := w.index.SearchOpening(&game)
	w.opening.Set(openingName)
	redraw(gtx)
	//}
}

func (w *Window) handleSearch(gtx layout.Context) {
	if w.search.button.Clicked(gtx) {
		maxMoves := w.moves.Selected()
		if maxMoves > 0 {
			game := w.board.Game()
			turn := w.turn.Selected()
			w.searching.Store(true)
			w.boardControls.Fade()
			w.results = make([]puzzle.Data, 0, 1000)
			redraw(gtx)
			go func() {
				for puzzle := range w.index.SearchPuzzles(&game, turn, maxMoves) {
					w.results = append(w.results, puzzle)
				}
				w.boardControls.Brighten()
				w.searching.Store(false)
				w.loaded.Store(false)
			}()
		}
	}
}

func (w *Window) handleCancel(gtx layout.Context) {

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
		layout.Rigid(Pad(w.padding, w.boardControls.Layout)),
	)
}

func (w *Window) layoutSearchPane(gtx layout.Context) layout.Dimensions {
	var button *IconButton
	if w.searching.Load() {
		button = w.cancel
	} else {
		button = w.search
	}

	if !w.loaded.Load() {
		w.puzzles.Add(w.results)
		w.loaded.Store(true)
	}

	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
		layout.Flexed(1, Pad(w.padding, w.puzzles.Layout)),
		layout.Rigid(PadSides(w.padding, w.moves.Layout)),
		layout.Rigid(PadSides(w.padding, w.turn.Layout)),
		layout.Rigid(Pad(w.padding, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, layout.Flexed(1, button.Layout))
		})),
	)
}

func redraw(gtx layout.Context) {
	gtx.Execute(op.InvalidateCmd{
		//At: gtx.Now.Add(time.Second / 25),
	})
}
