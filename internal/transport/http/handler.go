package http

import (
	"context"
	"fmt"
	"log"
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
}

func NewHandler(service CommentService) *Handler {
	h := &Handler{
		Service: service,
	}

	h.Router = mux.NewRouter()

	h.mapRoutes()

	h.Router.Use(
		LoggingMiddleware,
		JSONMiddleware,
		TimeoutMiddleware,
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
	h.Router.HandleFunc("/api/v1/comments", h.PostComment).Methods(http.MethodPost)
	h.Router.HandleFunc("/api/v1/comments/{id}", h.GetComment).Methods(http.MethodGet)
	h.Router.HandleFunc("/api/v1/comments/{id}", h.UpdateComment).Methods(http.MethodPut)
	h.Router.HandleFunc("/api/v1/comments/{id}", h.DeleteComment).Methods(http.MethodDelete)
}

func (h *Handler) Serve() error {
	go func() {
		if err := h.Server.ListenAndServe(); err != nil {
			log.Printf("failed to start server: %v", err)
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

	log.Println("Server shutdown gracefully")

	return nil
}
