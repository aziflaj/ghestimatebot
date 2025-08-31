package ghestimatebot

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"
)

type Server struct {
	srv http.Server
}

func NewServer(cfg *Config) *Server {
	// Setup routes
	mux := http.NewServeMux()

	mux.HandleFunc("GET /_/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	eventHandler := NewEventHandler(cfg)
	mux.HandleFunc("POST /wh", eventHandler.HandleWebhook)

	// Apply middleware
	httpHandler := applyMiddlewares(mux, recoveryMiddleware, loggingMiddleware)

	return &Server{
		srv: http.Server{
			Addr:           ":" + cfg.Port,
			Handler:        httpHandler,
			ReadTimeout:    15 * time.Second,
			WriteTimeout:   15 * time.Second,
			IdleTimeout:    60 * time.Second,
			MaxHeaderBytes: 1 << 20, // 1 MB
		},
	}
}

func (s *Server) Start(ctx context.Context) {
	s.srv.BaseContext = func(net.Listener) context.Context {
		return ctx
	}

	slog.Info("Starting HTTP server", "addr", s.srv.Addr)
	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Failed to start server", "error", err)
	}
}

func (s *Server) Stop(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
