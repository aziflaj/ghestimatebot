package ghbot

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"
)

type Server struct {
	srv http.Server
}

func NewServer(port string) *Server {
	// Setup routes
	mux := setupMux()

	// Apply middleware
	handler := applyMiddlewares(mux, recoveryMiddleware, loggingMiddleware)

	return &Server{
		srv: http.Server{
			Addr:           port,
			Handler:        handler,
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

func setupMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("GET /_/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	mux.Handle("POST /wh", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error("Failed to read request body", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		slog.Info("Received webhook", "method", r.Method, "path", r.URL.Path, "payload", string(payload))
		w.WriteHeader(http.StatusOK)
	}))

	return mux
}
