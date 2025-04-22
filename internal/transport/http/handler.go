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
	handlerTimeout  = 3
	shutdownTimeout = 15
)

type CommentService interface{}

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

	h.Server = &http.Server{
		IdleTimeout:       handlerTimeout * time.Second,
		WriteTimeout:      handlerTimeout * time.Second,
		ReadHeaderTimeout: handlerTimeout * time.Second,
		ReadTimeout:       handlerTimeout * time.Second,
		Addr:              addr,
		Handler:           h.Router,
	}

	return h
}

func (h *Handler) mapRoutes() {
	h.Router.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World! %s \n", r.Host)
	})
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
