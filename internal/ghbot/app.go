package ghbot

import (
	"context"
	"log/slog"
	"time"
)

func Run(ctx context.Context, port string) {
	server := NewServer(port)
	go server.Start(ctx)

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Stop(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}
}
