package ghestimatebot

import (
	"context"
	"log/slog"
	"time"
)

func Run(ctx context.Context, cfg *Config) {
	server := NewServer(cfg)
	go server.Start(ctx)

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Stop(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}
}
