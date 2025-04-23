package comment

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gofrs/uuid/v5"
)

var (
	ErrCommentNotFound = errors.New("comment not found")
	ErrNotImplemented  = errors.New("not implemented")
)

type Comment struct {
	ID     string `json:"id"`
	Slug   string `json:"slug"`
	Body   string `json:"body"`
	Author string `json:"author"`
}

type Store interface {
	GetComment(context.Context, string) (Comment, error)
	CreateComment(context.Context, Comment) (Comment, error)
	UpdateComment(context.Context, Comment) error
	DeleteComment(context.Context, string) error
}

type Service struct {
	Store  Store
	logger *slog.Logger
}

func NewService(store Store, logger *slog.Logger) *Service {
	return &Service{
		Store:  store,
		logger: logger,
	}
}

func (s *Service) CreateComment(ctx context.Context, c Comment) (Comment, error) {
	uuid, err := uuid.NewV7()
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to generate UUID", slog.Any("error", err))
		return Comment{}, fmt.Errorf("failed to generate UUID: %w", err)
	}

	c.ID = uuid.String()

	c, err = s.Store.CreateComment(ctx, c)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create comment", slog.Any("error", err))
		return Comment{}, fmt.Errorf("failed to create comment: %w", err)
	}

	return c, nil
}

func (s *Service) GetComment(ctx context.Context, id string) (Comment, error) {
	comment, err := s.Store.GetComment(ctx, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get comment", slog.Any("error", err))
		return Comment{}, ErrCommentNotFound
	}
	return comment, nil
}

func (s *Service) UpdateComment(ctx context.Context, c Comment) error {
	if err := s.Store.UpdateComment(ctx, c); err != nil {
		s.logger.ErrorContext(ctx, "failed to update comment", slog.Any("error", err))
		return fmt.Errorf("failed to update comment: %w", err)
	}
	return nil
}

func (s *Service) DeleteComment(ctx context.Context, id string) error {
	if err := s.Store.DeleteComment(ctx, id); err != nil {
		s.logger.ErrorContext(ctx, "failed to delete comment", slog.Any("error", err))
		return fmt.Errorf("failed to delete comment: %w", err)
	}
	return nil
}
