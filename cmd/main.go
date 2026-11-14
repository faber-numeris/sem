package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/faber-numeris/sem/internal/app/bootstrap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := bootstrap.Run(ctx); err != nil {
		slog.Error("Service terminated with error", "error", err)
		os.Exit(1)
	}
}
