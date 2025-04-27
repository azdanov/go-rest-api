//go:build e2e

package http_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/azdanov/go-rest-api/internal/comment"
	"github.com/azdanov/go-rest-api/internal/db"
	transportHttp "github.com/azdanov/go-rest-api/internal/transport/http"
	uuid "github.com/gofrs/uuid/v5"
	jwt "github.com/golang-jwt/jwt/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"resty.dev/v3"
)

const (
	testServerHost = "localhost"
	startupTimeout = 10 * time.Second
	jwtSigningKey  = "default"
)

type HandlerE2ETestSuite struct {
	suite.Suite
	pgContainer  *postgres.PostgresContainer
	db           *db.Database
	handler      *transportHttp.Handler
	client       *resty.Client
	serverCtx    context.Context
	serverCancel context.CancelFunc
	jwtToken     string
}

func (s *HandlerE2ETestSuite) getUUID() string {
	id, err := uuid.NewV7()
	s.Require().NoError(err)
	return id.String()
}

func findFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func (s *HandlerE2ETestSuite) SetupSuite() {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:17.4-alpine",
		postgres.WithDatabase("test-e2e-db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	s.Require().NoError(err)
	s.pgContainer = pgContainer

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	s.Require().NoError(err)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	newDB, err := db.NewDatabase(logger, connStr)
	s.Require().NoError(err)
	s.db = newDB

	err = s.db.Migrate("../../../migrations")
	s.Require().NoError(err)

	commentService := comment.NewService(s.db, logger)

	freePort, err := findFreePort()
	s.Require().NoError(err, "Failed to find free port")
	testAddr := fmt.Sprintf("%s:%d", testServerHost, freePort)

	s.handler = transportHttp.NewHandler(commentService, logger)
	s.handler.Server.Addr = testAddr

	s.serverCtx, s.serverCancel = context.WithCancel(context.Background())
	go func() {
		logger.Info("Starting test server", "address", s.handler.Server.Addr)
		if err = s.handler.Server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Test server failed", slog.Any("error", err))
			s.serverCancel()
		}
		logger.Info("Test server stopped")
	}()

	s.Require().Eventually(func() bool {
		conn, err := net.DialTimeout("tcp", s.handler.Server.Addr, 1*time.Second)
		if err != nil {
			logger.Debug("Server not ready yet", "address", s.handler.Server.Addr, "error", err)
			return false
		}
		conn.Close()
		logger.Info("Test server is ready", "address", s.handler.Server.Addr)
		return true
	}, startupTimeout, 100*time.Millisecond, "Server did not start within timeout")

	s.client = resty.New().SetBaseURL("http://" + s.handler.Server.Addr + "/api/v1")
	s.setupTestJWT()
}

func (s *HandlerE2ETestSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if s.handler != nil && s.handler.Server != nil {
		logger := slog.Default()
		logger.Info("Shutting down test server", "address", s.handler.Server.Addr)
		err := s.handler.Server.Shutdown(ctx)
		s.NoError(err, "Server shutdown failed")
	}
	if s.serverCancel != nil {
		s.serverCancel()
	}

	if s.pgContainer != nil {
		err := s.pgContainer.Terminate(ctx)
		s.NoError(err, "Postgres container termination failed")
	}

	if s.db != nil && s.db.Client != nil {
		err := s.db.Client.Close()
		s.NoError(err, "Database client close failed")
	}
}

func (s *HandlerE2ETestSuite) SetupTest() {
	_, err := s.db.Client.ExecContext(context.Background(), "DELETE FROM comments")
	s.Require().NoError(err)
	s.client.SetAuthToken(s.jwtToken)
	os.Setenv("JWT_SIGNING_KEY", jwtSigningKey)
}

func TestHandlerE2ETestSuite(t *testing.T) {
	suite.Run(t, new(HandlerE2ETestSuite))
}

func (s *HandlerE2ETestSuite) setupTestJWT() {
	jwtSecret := []byte(jwtSigningKey)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   "test-user-id",
		"email": "test@example.com",
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	s.Require().NoError(err, "Failed to sign JWT token")

	s.jwtToken = tokenString
	s.client.SetAuthToken(tokenString)
}

func (s *HandlerE2ETestSuite) TestPostComment_Success() {
	commentInput := map[string]string{
		"slug":   "e2e-test-slug",
		"body":   "This is the e2e test body",
		"author": "e2e-tester",
	}

	var createdComment comment.Comment
	resp, err := s.client.R().
		SetBody(commentInput).
		SetResult(&createdComment).
		Post("/comments")

	s.Require().NoError(err)
	s.Equal(http.StatusCreated, resp.StatusCode())
	s.Equal("/api/v1/comments/"+createdComment.ID, resp.Header().Get("Location"))
	s.NotEmpty(createdComment.ID)
	s.Equal(commentInput["slug"], createdComment.Slug)
	s.Equal(commentInput["body"], createdComment.Body)
	s.Equal(commentInput["author"], createdComment.Author)

	dbComment, dbErr := s.db.GetComment(context.Background(), createdComment.ID)
	s.Require().NoError(dbErr)
	s.Equal(createdComment, dbComment)
}

func (s *HandlerE2ETestSuite) TestGetComment_Success() {
	seedComment := comment.Comment{
		ID:     s.getUUID(),
		Slug:   "get-slug",
		Body:   "get body",
		Author: "get author",
	}
	_, err := s.db.CreateComment(context.Background(), seedComment)
	s.Require().NoError(err)

	var fetchedComment comment.Comment
	resp, err := s.client.R().
		SetResult(&fetchedComment).
		Get("/comments/" + seedComment.ID)

	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode(), "Response body: %s", resp.String())
	s.Equal(seedComment.ID, fetchedComment.ID)
	s.Equal(seedComment.Slug, fetchedComment.Slug)
	s.Equal(seedComment.Body, fetchedComment.Body)
	s.Equal(seedComment.Author, fetchedComment.Author)
}

func (s *HandlerE2ETestSuite) TestGetComment_NotFound() {
	nonExistentID := s.getUUID()

	resp, err := s.client.R().
		Get("/comments/" + nonExistentID)

	s.Require().NoError(err)
	s.Equal(http.StatusNotFound, resp.StatusCode(), "Response body: %s", resp.String())
}

func (s *HandlerE2ETestSuite) TestUpdateComment_Success() {
	seedComment := comment.Comment{
		ID:     s.getUUID(),
		Slug:   "update-slug-initial",
		Body:   "update body initial",
		Author: "update author",
	}
	_, err := s.db.CreateComment(context.Background(), seedComment)
	s.Require().NoError(err)

	updateInput := map[string]string{
		"id":     seedComment.ID,
		"slug":   "update-slug-final",
		"body":   "update body final",
		"author": seedComment.Author,
	}

	var updatedComment any
	resp, err := s.client.R().
		SetBody(updateInput).
		Put("/comments/" + seedComment.ID)

	s.Require().NoError(err)
	s.Equal(http.StatusNoContent, resp.StatusCode())
	s.Nil(updatedComment)

	dbComment, dbErr := s.db.GetComment(context.Background(), seedComment.ID)
	s.Require().NoError(dbErr)
	s.Equal(seedComment.ID, dbComment.ID)
	s.Equal(updateInput["slug"], dbComment.Slug)
	s.Equal(updateInput["body"], dbComment.Body)
	s.Equal(seedComment.Author, dbComment.Author)
}

func (s *HandlerE2ETestSuite) TestDeleteComment_Success() {
	seedComment := comment.Comment{
		ID:     s.getUUID(),
		Slug:   "delete-slug",
		Body:   "delete body",
		Author: "delete author",
	}
	_, err := s.db.CreateComment(context.Background(), seedComment)
	s.Require().NoError(err)

	resp, err := s.client.R().
		Delete("/comments/" + seedComment.ID)

	s.Require().NoError(err)
	s.Equal(http.StatusNoContent, resp.StatusCode())

	getResp, err := s.client.R().
		Get("/comments/" + seedComment.ID)
	s.Require().NoError(err)
	s.Equal(http.StatusNotFound, getResp.StatusCode())

	_, dbErr := s.db.GetComment(context.Background(), seedComment.ID)
	s.Require().Error(dbErr)
}

func (s *HandlerE2ETestSuite) TestPostComment_ValidationError() {
	commentInput := map[string]string{
		"slug":   "e2e-validation-slug",
		"author": "e2e-validator",
		// "body": "missing",
	}

	var errorResponse map[string]any
	resp, err := s.client.R().
		SetBody(commentInput).
		SetError(&errorResponse).
		Post("/comments")

	s.Require().NoError(err)
	s.Equal(http.StatusBadRequest, resp.StatusCode())
	s.Nil(errorResponse)
}

func (s *HandlerE2ETestSuite) TestGetComment_InvalidIDFormat() {
	invalidID := "not-a-uuid"

	resp, err := s.client.R().
		Get("/comments/" + invalidID)

	s.Require().NoError(err)

	s.Equal(http.StatusBadRequest, resp.StatusCode())
}
