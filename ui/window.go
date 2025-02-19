package ui

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/failosof/cops/core"
	"github.com/failosof/cops/ui/board"
	"github.com/failosof/cops/ui/board/util"
	"github.com/notnil/chess"
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
	config, err := chessboard.NewConfig(chessBoardRes.BackgroundFile, chessBoardRes.PiecesDir)
	if err != nil {
		return fmt.Errorf("failed to config chess board widget: %w", err)
	}

	config.ShowHints = true
	config.ShowLastMove = true

	game := chess.NewGame()

	th := material.NewTheme()
	board := chessboard.NewWidget(th, config)
	board.SetGame(game)

	var ops op.Ops
	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			layout.Background{}.Layout(gtx,
				func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(unit.Dp(10)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						defer clip.Rect{Max: gtx.Constraints.Min}.Push(gtx.Ops).Pop()
						paint.Fill(gtx.Ops, util.GrayColor)
						return layout.Dimensions{Size: gtx.Constraints.Min}
					})
				},
				func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{}.Layout(gtx)
				},
			)
		}
	}
}
