package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
)

const (
	addr            = "0.0.0.0:8080"
	requestTimeout  = 10
	serverTimeout   = 15
	shutdownTimeout = 15
)

type Handler struct {
	Router  *mux.Router
	Service CommentService
	Server  *http.Server
	logger  *slog.Logger
}

func NewHandler(service CommentService, logger *slog.Logger) *Handler {
	h := &Handler{
		Service: service,
		logger:  logger,
	}

	h.Router = mux.NewRouter()

	h.mapRoutes()

	h.Router.Use(
		h.LoggingMiddleware,
		h.JSONMiddleware,
		h.TimeoutMiddleware,
	)

	h.Server = &http.Server{
		IdleTimeout:       serverTimeout * time.Second,
		WriteTimeout:      serverTimeout * time.Second,
		ReadHeaderTimeout: serverTimeout * time.Second,
		ReadTimeout:       serverTimeout * time.Second,
		Addr:              addr,
		Handler:           h.Router,
	}

	return h
}

func (h *Handler) mapRoutes() {
	h.Router.HandleFunc("/api/v1/comments", h.JWTAuth(h.PostComment)).Methods(http.MethodPost)
	h.Router.HandleFunc("/api/v1/comments/{id}", h.JWTAuth(h.UpdateComment)).Methods(http.MethodPut)
	h.Router.HandleFunc("/api/v1/comments/{id}", h.JWTAuth(h.DeleteComment)).Methods(http.MethodDelete)
	h.Router.HandleFunc("/api/v1/comments/{id}", h.GetComment).Methods(http.MethodGet)
}

func (h *Handler) Serve() error {
	go func() {
		if err := h.Server.ListenAndServe(); err != nil {
			h.logger.Error("failed to start server", slog.Any("error", err))
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout*time.Second)
	defer cancel()
	if err := h.Server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	h.logger.Info("server shutdown gracefully")

	return nil
}
