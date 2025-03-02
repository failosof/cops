package ui

import (
	"context"
	"log/slog"
	"os"
	"sync/atomic"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/failosof/cops/core"
	"github.com/failosof/cops/ui/board"
)

// todo: puzzle search pagination

const PageSize = 25 // puzzles on one page

type Window struct {
	window  *app.Window
	padding unit.Dp
	theme   *material.Theme

	opening       *OpeningName
	board         *chessboard.Widget
	boardControls *BoardControls

	moves   *MovesCountSelector
	turn    *TurnSelector
	puzzles *PuzzleList
	search  *IconButton

	resourcesLoaded  atomic.Bool
	loadingStatus    string
	index            *core.Index
	chessBoardConfig *chessboard.Config

	searching     atomic.Bool
	puzzlesLoaded atomic.Bool
	results       []core.PuzzleData
}

func NewWindow() (*Window, error) {
	return &Window{
		window:  new(app.Window),
		padding: unit.Dp(3),
	}, nil
}

func (w *Window) Show(ctx context.Context) {
	w.window.Option(app.Title("Chess Opening Puzzle Search"))
	w.window.Option(app.MinSize(unit.Dp(820), unit.Dp(620)))
	w.window.Option(app.MaxSize(unit.Dp(820), unit.Dp(620)))

	w.theme = material.NewTheme()

	w.loadingStatus = "Loading..."
	w.opening = NewOpeningName(w.theme)
	w.boardControls = NewBoardControls(w.theme)

	w.moves = NewMovesNumberSelector(w.theme, 1, 40)
	w.turn = NewTurnSelector(w.theme)
	w.puzzles = NewPuzzleList(w.theme)
	w.search = NewIconButton(w.theme, SearchIcon, GreenColor)

	go func() {
		if err := w.update(ctx); err != nil {
			slog.Error("main window update", "err", err)
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}()

	go func() {
		var err error

		w.index, err = core.LoadIndex()
		if err != nil {
			slog.Error("failed to load index", "err", err)
			w.loadingStatus = "Index load error"
			return
		}

		w.chessBoardConfig, err = LoadChessBoardConfig()
		if err != nil {
			slog.Error("failed to load chess board config", "err", err)
			w.loadingStatus = "Assets load error"
			return
		}
		w.chessBoardConfig.ShowHints = true
		w.chessBoardConfig.ShowLastMove = true

		w.board = chessboard.NewWidget(w.theme, w.chessBoardConfig)

		w.resourcesLoaded.Store(true)
	}()

	app.Main()
}

func (w *Window) update(ctx context.Context) error {
	var ops op.Ops
	for {
		switch e := w.window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			select {
			case <-ctx.Done():
				return nil
			default:
			}

			gtx := app.NewContext(&ops, e)
			if w.resourcesLoaded.Load() {
				if !w.searching.Load() {
					w.handleControls(gtx)
					w.handleBoard(gtx)
					w.handleSearch(gtx)
				}
				w.layoutWidgets(gtx)
			} else {
				w.layoutLoading(gtx)
				redraw(gtx)
			}
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
		game := w.board.Game()
		turn := w.turn.Selected()
		w.searching.Store(true)
		w.boardControls.Fade()
		w.results = make([]core.PuzzleData, 0, 1000)
		redraw(gtx)
		go func() {
			for puzzle := range w.index.SearchPuzzles(&game, turn, maxMoves) {
				w.results = append(w.results, puzzle)
			}
			w.boardControls.Brighten()
			w.searching.Store(false)
			w.puzzlesLoaded.Store(false)
		}()
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
				Color:        BlackColor,
				CornerRadius: unit.Dp(1),
				Width:        unit.Dp(1),
			}.Layout(gtx, w.board.Layout)
		})),
		layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
		layout.Rigid(Pad(w.padding, w.boardControls.Layout)),
	)
}

func (w *Window) layoutSearchPane(gtx layout.Context) layout.Dimensions {
	if !w.puzzlesLoaded.Load() {
		w.puzzles.Add(w.results)
		w.puzzlesLoaded.Store(true)
	}

	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
		layout.Flexed(1, Pad(w.padding, w.puzzles.Layout)),
		layout.Rigid(PadSides(w.padding, w.moves.Layout)),
		layout.Rigid(PadSides(w.padding, w.turn.Layout)),
		layout.Rigid(Pad(w.padding, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, layout.Flexed(1, w.search.Layout))
		})),
	)
}

func (w *Window) layoutLoading(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle, Spacing: layout.SpaceAround}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return material.H3(w.theme, w.loadingStatus).Layout(gtx)
			})
		}),
	)
}

func redraw(gtx layout.Context) {
	gtx.Execute(op.InvalidateCmd{})
}
