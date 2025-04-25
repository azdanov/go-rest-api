package main

import (
	"fmt"
	"log/slog"

	"github.com/azdanov/go-rest-api/internal/comment"
	"github.com/azdanov/go-rest-api/internal/db"
	transportHttp "github.com/azdanov/go-rest-api/internal/transport/http"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func Run(logger *slog.Logger) error {
	logger.Info("starting server")

	db, err := db.NewDatabase(logger, db.CreateConnectionString())
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	if err = db.Migrate("./migrations"); err != nil {
		return err
	}

	commentService := comment.NewService(db, logger)

	httpHandler := transportHttp.NewHandler(commentService, logger)
	if err = httpHandler.Serve(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

func main() {
	logger := slog.Default()

	if err := Run(logger); err != nil {
		logger.Error("failed to run server", slog.Any("error", err))
	}
}
