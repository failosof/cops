package ui

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/failosof/cops/core"
	"github.com/failosof/cops/resources"
	"github.com/failosof/giochess/board"
	"github.com/notnil/chess"
)

const PageSize = 30 // puzzles on one page

type Window struct {
	window  *app.Window
	padding unit.Dp
	theme   *material.Theme

	// left pane
	opening       *OpeningName
	board         *chessboard.Widget
	fen           *TextField
	pgn           *TextField
	boardControls *BoardControls

	// right pane
	movesCount     *RangeSlider
	turn           *OptionSelector[core.Turn]
	searchStrategy *OptionSelector[core.SearchType]
	puzzles        *TextField

	search *IconButton

	// state
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
	w.window.Option(app.MinSize(unit.Dp(920), unit.Dp(825)))

	w.theme = material.NewTheme()

	w.loadingStatus = "Loading..."
	w.opening = NewOpeningName(w.theme)
	w.fen = NewTextField(w.theme, "FEN", ReadOnly|SingleLine)
	w.pgn = NewTextField(w.theme, "PGN", ReadOnly)
	w.boardControls = NewBoardControls(w.theme)

	w.movesCount = NewRangeSlider(w.theme, "Moves", 1, 40)
	w.turn = NewOptionSelector(w.theme, []core.Turn{core.WhiteTurn, core.BlackTurn, core.EitherTurn})
	w.searchStrategy = NewOptionSelector(w.theme, []core.SearchType{core.MoveSequenceSearch, core.PositionSearch})
	w.puzzles = NewTextField(w.theme, "Lichess puzzle links", ReadOnly)
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

		textures, err := resources.LoadChessBoardTextures()
		if err != nil {
			slog.Error("failed to load chess board textures", "err", err)
			w.loadingStatus = "Assets load error"
			return
		}

		w.chessBoardConfig = chessboard.NewConfig(textures.Board, textures.Pieces, chessboard.Colors{
			Hint:     Transparentize(GrayColor, 0.7),
			LastMove: Transparentize(YellowColor, 0.5),
			Primary:  Transparentize(GreenColor, 0.7),
			Info:     Transparentize(BlueColor, 0.7),
			Warning:  Transparentize(YellowColor, 0.7),
			Danger:   Transparentize(RedColor, 0.7),
		})
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

			slog.Info("update", "w", gtx.Constraints.Max.X, "h", gtx.Constraints.Max.Y)
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
	game := w.board.Game()

	openingName, _ := w.index.SearchOpening(game)
	w.opening.Set(openingName)

	w.fen.SetText(game.Position().String())

	positions := game.Positions()
	var pgn strings.Builder
	var notation chess.AlgebraicNotation
	for i, move := range game.Moves() {
		if i&1 == 0 {
			pgn.WriteString(strconv.Itoa(i/2 + 1))
			pgn.WriteString(". ")
		}
		pgn.WriteString(notation.Encode(positions[i], move))
		pgn.WriteString(" ")
	}
	w.pgn.SetText(pgn.String())
}

func (w *Window) handleSearch(gtx layout.Context) {
	if w.search.button.Clicked(gtx) {
		w.searching.Store(true)

		maxMoves := w.movesCount.Selected()
		turn := w.turn.Selected()
		strategy := w.searchStrategy.Selected()

		go func() {
			w.resultsMu.Lock()
			defer w.resultsMu.Unlock()

			start := time.Now()
			results := w.index.SearchPuzzles(w.board.Game(), strategy, turn.ToChess(), maxMoves)
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
		layout.Flexed(6, Pad(w.padding, func(gtx layout.Context) layout.Dimensions {
			return widget.Border{
				Color:        BlackColor,
				CornerRadius: unit.Dp(1),
				Width:        unit.Dp(1),
			}.Layout(gtx, w.board.Layout)
		})),
		layout.Rigid(Pad(w.padding, w.fen.Layout)),
		layout.Flexed(1, Pad(w.padding, w.pgn.Layout)),
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
		var text strings.Builder
		for _, puzzle := range results {
			text.WriteString(puzzle.URL())
			text.WriteRune('\n')
		}
		w.puzzles.SetText(text.String())
		w.resultsMu.Unlock()
		gtx.Execute(op.InvalidateCmd{})
	}

	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
		layout.Flexed(1, Pad(w.padding, w.puzzles.Layout)),
		layout.Rigid(PadSides(w.padding, w.movesCount.Layout)),
		layout.Rigid(PadSides(w.padding, w.turn.Layout)),
		layout.Rigid(PadSides(w.padding, w.searchStrategy.Layout)),
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
