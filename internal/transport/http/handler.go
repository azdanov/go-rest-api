package http

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

const (
	addr    = "0.0.0.0:8080"
	timeout = 3 * time.Second
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
		IdleTimeout:       timeout,
		WriteTimeout:      timeout,
		ReadHeaderTimeout: timeout,
		ReadTimeout:       timeout,
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
	log.Printf("Starting server on http://%s\n", addr)
	if err := h.Server.ListenAndServe(); err != nil {
		return err
	}
	return nil
}
