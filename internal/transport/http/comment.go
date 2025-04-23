package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/azdanov/go-rest-api/internal/comment"
	"github.com/gorilla/mux"
)

type CommentService interface {
	CreateComment(context.Context, comment.Comment) (comment.Comment, error)
	GetComment(context.Context, string) (comment.Comment, error)
	UpdateComment(context.Context, comment.Comment) error
	DeleteComment(context.Context, string) error
}

func (h *Handler) PostComment(w http.ResponseWriter, r *http.Request) {
	var comment comment.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		h.logger.ErrorContext(r.Context(), "failed to decode request body", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	comment, err := h.Service.CreateComment(r.Context(), comment)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to create comment", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(comment); err != nil {
		h.logger.ErrorContext(r.Context(), "failed to encode response", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["id"]
	if commentID == "" {
		http.Error(w, "comment ID is required", http.StatusBadRequest)
		return
	}

	comment, err := h.Service.GetComment(r.Context(), commentID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to get comment", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(comment); err != nil {
		h.logger.ErrorContext(r.Context(), "failed to encode response", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["id"]
	if commentID == "" {
		http.Error(w, "comment ID is required", http.StatusBadRequest)
		return
	}

	var comment comment.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		h.logger.ErrorContext(r.Context(), "failed to decode request body", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if comment.ID != commentID {
		http.Error(w, "comment ID in URL and body must match", http.StatusBadRequest)
		return
	}

	if err := h.Service.UpdateComment(r.Context(), comment); err != nil {
		h.logger.ErrorContext(r.Context(), "failed to update comment", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["id"]
	if commentID == "" {
		http.Error(w, "comment ID is required", http.StatusBadRequest)
		return
	}

	if err := h.Service.DeleteComment(r.Context(), commentID); err != nil {
		h.logger.ErrorContext(r.Context(), "failed to delete comment", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
