//go:build unit
// +build unit

package comment_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/azdanov/go-rest-api/internal/comment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockStore struct {
	mock.Mock
}

func (m *MockStore) GetComment(ctx context.Context, id string) (comment.Comment, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(comment.Comment), args.Error(1)
}

func (m *MockStore) CreateComment(ctx context.Context, c comment.Comment) (comment.Comment, error) {
	args := m.Called(ctx, c)

	cmt := args.Get(0).(comment.Comment)
	cmt.ID = "00000000-0000-0000-0000-000000000000" // Simulate ID generation
	return cmt, args.Error(1)
}

func (m *MockStore) UpdateComment(ctx context.Context, c comment.Comment) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *MockStore) DeleteComment(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestCreateComment_Success(t *testing.T) {
	mockStore := new(MockStore)
	logger := slog.Default()
	service := comment.NewService(mockStore, logger)

	ctx := t.Context()
	comment := comment.Comment{
		Slug:   "test-slug",
		Body:   "This is a test comment",
		Author: "test-author",
	}

	mockStore.On("CreateComment", ctx, mock.AnythingOfType("comment.Comment")).Return(comment, nil)

	createdComment, err := service.CreateComment(ctx, comment)

	require.NoError(t, err)
	assert.NotEmpty(t, createdComment.ID)
	assert.Equal(t, comment.Slug, createdComment.Slug)
	assert.Equal(t, comment.Body, createdComment.Body)
	assert.Equal(t, comment.Author, createdComment.Author)

	mockStore.AssertExpectations(t)
}

func TestGetComment_NotFound(t *testing.T) {
	mockStore := new(MockStore)
	logger := slog.Default()
	service := comment.NewService(mockStore, logger)

	ctx := t.Context()
	mockStore.On("GetComment", ctx, "non-existent-id").Return(comment.Comment{}, comment.ErrCommentNotFound)

	_, err := service.GetComment(ctx, "non-existent-id")
	require.ErrorIs(t, err, comment.ErrCommentNotFound)

	mockStore.AssertExpectations(t)
}

func TestGetComment_Success(t *testing.T) {
	mockStore := new(MockStore)
	logger := slog.Default()
	service := comment.NewService(mockStore, logger)

	ctx := t.Context()
	expectedComment := comment.Comment{
		ID:     "test-id",
		Slug:   "test-slug",
		Body:   "This is a test comment",
		Author: "test-author",
	}

	mockStore.On("GetComment", ctx, "test-id").Return(expectedComment, nil)

	retrievedComment, err := service.GetComment(ctx, "test-id")

	require.NoError(t, err)
	assert.Equal(t, expectedComment.ID, retrievedComment.ID)
	assert.Equal(t, expectedComment.Slug, retrievedComment.Slug)
	assert.Equal(t, expectedComment.Body, retrievedComment.Body)
	assert.Equal(t, expectedComment.Author, retrievedComment.Author)

	mockStore.AssertExpectations(t)
}

func TestUpdateComment_Success(t *testing.T) {
	mockStore := new(MockStore)
	logger := slog.Default()
	service := comment.NewService(mockStore, logger)

	ctx := t.Context()
	commentToUpdate := comment.Comment{
		ID:     "test-id",
		Slug:   "updated-slug",
		Body:   "This is an updated comment",
		Author: "test-author",
	}

	mockStore.On("UpdateComment", ctx, commentToUpdate).Return(nil)

	err := service.UpdateComment(ctx, commentToUpdate)

	require.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestUpdateComment_Error(t *testing.T) {
	mockStore := new(MockStore)
	logger := slog.Default()
	service := comment.NewService(mockStore, logger)

	ctx := t.Context()
	commentToUpdate := comment.Comment{
		ID:     "test-id",
		Slug:   "updated-slug",
		Body:   "This is an updated comment",
		Author: "test-author",
	}

	mockError := errors.New("update failed")
	mockStore.On("UpdateComment", ctx, commentToUpdate).Return(mockError)

	err := service.UpdateComment(ctx, commentToUpdate)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update comment")
	mockStore.AssertExpectations(t)
}

func TestDeleteComment_Success(t *testing.T) {
	mockStore := new(MockStore)
	logger := slog.Default()
	service := comment.NewService(mockStore, logger)

	ctx := t.Context()
	mockStore.On("DeleteComment", ctx, "test-id").Return(nil)

	err := service.DeleteComment(ctx, "test-id")

	require.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestDeleteComment_Error(t *testing.T) {
	mockStore := new(MockStore)
	logger := slog.Default()
	service := comment.NewService(mockStore, logger)

	ctx := t.Context()
	mockError := errors.New("delete failed")
	mockStore.On("DeleteComment", ctx, "test-id").Return(mockError)

	err := service.DeleteComment(ctx, "test-id")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete comment")
	mockStore.AssertExpectations(t)
}
