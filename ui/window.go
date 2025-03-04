package ui

import (
	"context"
	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/failosof/cops/core"
	"github.com/failosof/cops/ui/board"
	"github.com/notnil/chess"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"time"
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

	game *chess.Game

	resourcesLoaded  atomic.Bool
	loadingStatus    string
	index            *core.Index
	chessBoardConfig *chessboard.Config

	searching     atomic.Bool
	resultsLoaded atomic.Bool
	resultsMu     sync.RWMutex
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
		w.window.Invalidate()
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
				} else {
					gtx = gtx.Disabled()
				}
				w.layoutWidgets(gtx)
			} else {
				w.layoutLoading(gtx)
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
	w.window.Invalidate()
}

func (w *Window) handleBoard(gtx layout.Context) {
	//if w.board.PositionChanged() {
	openingName, _ := w.index.SearchOpening(w.board.Game())
	w.opening.Set(openingName)
	//gtx.Execute(op.InvalidateCmd{})
	//}
}

func (w *Window) handleSearch(gtx layout.Context) {
	if w.search.button.Clicked(gtx) {
		w.searching.Store(true)

		maxMoves := w.moves.Selected()
		turn := w.turn.Selected()
		go func() {
			w.resultsMu.Lock()
			defer w.resultsMu.Unlock()

			start := time.Now()
			results := w.index.SearchPuzzles(w.board.Game(), turn, maxMoves)
			took := time.Since(start)
			slog.Info("puzzle search", "found", len(results), "took", took)

			w.results = make([]core.PuzzleData, len(results))
			copy(w.results, results)

			w.searching.Store(false)
			w.resultsLoaded.Store(false)
			w.window.Invalidate()
		}()

		gtx.Execute(op.InvalidateCmd{})
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
	if !w.resultsLoaded.Load() {
		w.resultsLoaded.Store(true)
		w.resultsMu.Lock()
		results := w.results
		if len(results) > PageSize {
			results = results[:PageSize]
		}
		w.puzzles.Add(results)
		w.resultsMu.Unlock()
		gtx.Execute(op.InvalidateCmd{})
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
