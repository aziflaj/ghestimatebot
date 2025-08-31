package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/aziflaj/ghbot/internal/ghbot"
	"github.com/joho/godotenv"
)

func main() {
	// setup logger
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(h))

	err := godotenv.Load()
	if err != nil {
		slog.Error("Failed to load environment variables", "error", err)
		return
	}

	// handle signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		slog.Info("Received signal, shutting down", "signal", sig)
		cancel()
	}()

	// read port from env
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}

	ghbot.Run(ctx, ":"+port)
}
