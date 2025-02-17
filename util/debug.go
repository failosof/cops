package util

import (
	"log/slog"
	"os"
)

func Assert(fact bool, msg string) {
	if !fact {
		slog.Error("assertion failed", "msg", msg)
		os.Exit(1)
	}
}
