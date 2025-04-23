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

type PostCommentRequest struct {
	Slug   string `json:"slug"   validate:"required"`
	Body   string `json:"body"   validate:"required"`
	Author string `json:"author" validate:"required"`
}

func convertToComment(pcr PostCommentRequest) comment.Comment {
	return comment.Comment{
		Slug:   pcr.Slug,
		Body:   pcr.Body,
		Author: pcr.Author,
	}
}

func (h *Handler) PostComment(w http.ResponseWriter, r *http.Request) {
	var pcr PostCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&pcr); err != nil {
		h.logger.ErrorContext(r.Context(), "failed to decode request body", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(pcr); err != nil {
		h.logger.ErrorContext(r.Context(), "validation failed", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	comment := convertToComment(pcr)
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

type UpdateCommentRequest struct {
	ID     string `json:"id"     validate:"required,uuid"`
	Slug   string `json:"slug"   validate:"required"`
	Body   string `json:"body"   validate:"required"`
	Author string `json:"author" validate:"required"`
}

func convertToUpdateComment(ucr UpdateCommentRequest) comment.Comment {
	return comment.Comment{
		ID:     ucr.ID,
		Slug:   ucr.Slug,
		Body:   ucr.Body,
		Author: ucr.Author,
	}
}

func (h *Handler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["id"]
	if commentID == "" {
		http.Error(w, "comment ID is required", http.StatusBadRequest)
		return
	}

	var ucr UpdateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&ucr); err != nil {
		h.logger.ErrorContext(r.Context(), "failed to decode request body", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if ucr.ID != commentID {
		http.Error(w, "comment ID in URL and body must match", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(ucr); err != nil {
		h.logger.ErrorContext(r.Context(), "validation failed", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	comment := convertToUpdateComment(ucr)

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
