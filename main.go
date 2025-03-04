package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/failosof/cops/ui"
)

// todo: pagination

// todo: search options: by moves sequence or by position
// todo: cache games with position hashes

const MemLimit = 4 * 1024 * 1024 * 1024 // 4 Gb

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// turn off gc until MemLimit reached
	debug.SetMemoryLimit(MemLimit)
	debug.SetGCPercent(-1)

	window, err := ui.NewWindow()
	if err != nil {
		slog.Error("failed to create main window", "err", err)
		os.Exit(1)
	}
	window.Show(ctx)
}
