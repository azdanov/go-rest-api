//go:build integration

package db_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/azdanov/go-rest-api/internal/comment"
	"github.com/azdanov/go-rest-api/internal/db"
	uuid "github.com/gofrs/uuid/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type CommentTestSuite struct {
	suite.Suite
	pgContainer *postgres.PostgresContainer
	db          *db.Database
}

func (s *CommentTestSuite) getUUID() string {
	uuid, err := uuid.NewV7()
	if err != nil {
		require.NoError(s.T(), err)
	}

	return uuid.String()
}

func (s *CommentTestSuite) SetupSuite() {
	ctx := context.Background()
	pgContainer, err := postgres.Run(ctx,
		"postgres:17.4-alpine",
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	require.NoError(s.T(), err)
	s.pgContainer = pgContainer

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(s.T(), err)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	newDB, err := db.NewDatabase(logger, connStr)
	require.NoError(s.T(), err)
	s.db = newDB

	// Run migrations
	err = s.db.Migrate("../../migrations")
	require.NoError(s.T(), err)
}

func (s *CommentTestSuite) TearDownSuite() {
	ctx := context.Background()
	err := s.pgContainer.Terminate(ctx)
	require.NoError(s.T(), err)
	err = s.db.Client.Close()
	require.NoError(s.T(), err)
}

func (s *CommentTestSuite) SetupTest() {
	// Optional: Clean up tables before each test if needed
	_, err := s.db.Client.ExecContext(context.Background(), "DELETE FROM comments")
	require.NoError(s.T(), err)
}

func TestCommentTestSuite(t *testing.T) {
	suite.Run(t, new(CommentTestSuite))
}

func (s *CommentTestSuite) TestCreateComment() {
	ctx := context.Background()
	cmt := comment.Comment{
		ID:     s.getUUID(),
		Slug:   "test-slug",
		Body:   "test body",
		Author: "test author",
	}

	createdCmt, err := s.db.CreateComment(ctx, cmt)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), cmt.ID, createdCmt.ID)
	assert.Equal(s.T(), cmt.Slug, createdCmt.Slug)
	assert.Equal(s.T(), cmt.Body, createdCmt.Body)
	assert.Equal(s.T(), cmt.Author, createdCmt.Author)

	// Verify by getting
	fetchedCmt, err := s.db.GetComment(ctx, cmt.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), cmt, fetchedCmt)
}

func (s *CommentTestSuite) TestGetComment() {
	ctx := context.Background()
	cmt := comment.Comment{
		ID:     s.getUUID(),
		Slug:   "get-slug",
		Body:   "get body",
		Author: "get author",
	}
	_, err := s.db.CreateComment(ctx, cmt) // Insert first
	require.NoError(s.T(), err)

	fetchedCmt, err := s.db.GetComment(ctx, cmt.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), cmt, fetchedCmt)
}

func (s *CommentTestSuite) TestGetComment_NotFound() {
	ctx := context.Background()
	nonExistentID := s.getUUID()

	_, err := s.db.GetComment(ctx, nonExistentID)
	require.Error(s.T(), err)
	// Depending on your error handling, you might want a more specific check
	assert.Contains(s.T(), err.Error(), "failed to scan comment row")
}

func (s *CommentTestSuite) TestUpdateComment() {
	ctx := context.Background()
	cmt := comment.Comment{
		ID:     s.getUUID(),
		Slug:   "update-slug-initial",
		Body:   "update body initial",
		Author: "update author initial",
	}
	_, err := s.db.CreateComment(ctx, cmt) // Insert first
	require.NoError(s.T(), err)

	// Update fields
	cmt.Slug = "update-slug-updated"
	cmt.Body = "update body updated"

	err = s.db.UpdateComment(ctx, cmt)
	require.NoError(s.T(), err)

	// Verify by getting
	fetchedCmt, err := s.db.GetComment(ctx, cmt.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), cmt.Slug, fetchedCmt.Slug)
	assert.Equal(s.T(), cmt.Body, fetchedCmt.Body)
	assert.Equal(s.T(), cmt.Author, fetchedCmt.Author) // Author should remain the same
}

func (s *CommentTestSuite) TestDeleteComment() {
	ctx := context.Background()
	cmt := comment.Comment{
		ID:     s.getUUID(),
		Slug:   "delete-slug",
		Body:   "delete body",
		Author: "delete author",
	}
	_, err := s.db.CreateComment(ctx, cmt) // Insert first
	require.NoError(s.T(), err)

	err = s.db.DeleteComment(ctx, cmt.ID)
	require.NoError(s.T(), err)

	// Verify deletion by trying to get
	_, err = s.db.GetComment(ctx, cmt.ID)
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "failed to scan comment row")
}
