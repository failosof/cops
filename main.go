package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/failosof/cops/ui"
)

/*
todo:
    ? 2. make index creation tool
    7. puzzle search pagination
    8. move out board state update
    9. fix captures on the board
    10. finish board widgets todos
*/

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	window, err := ui.NewWindow()
	if err != nil {
		slog.Error("failed to create main window", "err", err)
		os.Exit(1)
	}
	window.Show(ctx)
}
