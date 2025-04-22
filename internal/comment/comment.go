package comment

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/gofrs/uuid/v5"
)

var (
	ErrCommentNotFound = errors.New("comment not found")
	ErrNotImplemented  = errors.New("not implemented")
)

type Comment struct {
	ID     string
	Slug   string
	Body   string
	Author string
}

type Store interface {
	GetComment(context.Context, string) (Comment, error)
	CreateComment(context.Context, Comment) (Comment, error)
	UpdateComment(context.Context, Comment) error
	DeleteComment(context.Context, string) error
}

type Service struct {
	Store Store
}

func NewService(store Store) *Service {
	return &Service{
		Store: store,
	}
}

func (s *Service) CreateComment(ctx context.Context, c Comment) (Comment, error) {
	uuid, err := uuid.NewV7()
	if err != nil {
		log.Println(err)
		return Comment{}, fmt.Errorf("failed to generate UUID: %w", err)
	}

	c.ID = uuid.String()

	c, err = s.Store.CreateComment(ctx, c)
	if err != nil {
		log.Println(err)
		return Comment{}, fmt.Errorf("failed to create comment: %w", err)
	}
	return c, nil
}

func (s *Service) GetComment(ctx context.Context, id string) (Comment, error) {
	comment, err := s.Store.GetComment(ctx, id)
	if err != nil {
		log.Println(err)
		return Comment{}, ErrCommentNotFound
	}
	return comment, nil
}

func (s *Service) UpdateComment(ctx context.Context, c Comment) error {
	err := s.Store.UpdateComment(ctx, c)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("failed to update comment: %w", err)
	}
	return nil
}

func (s *Service) DeleteComment(ctx context.Context, id string) error {
	err := s.Store.DeleteComment(ctx, id)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("failed to delete comment: %w", err)
	}
	return nil
}
