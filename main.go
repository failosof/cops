package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/failosof/cops/ui"
)

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
