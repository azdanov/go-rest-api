package main

import (
	"fmt"
	"log"

	"github.com/azdanov/go-rest-api/internal/comment"
	"github.com/azdanov/go-rest-api/internal/db"
	transportHttp "github.com/azdanov/go-rest-api/internal/transport/http"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func Run() error {
	log.Println("running server...")

	db, err := db.NewDatabase()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	if err = db.Migrate(); err != nil {
		return err
	}

	commentService := comment.NewService(db)

	httpHandler := transportHttp.NewHandler(commentService)
	if err = httpHandler.Serve(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

func main() {
	if err := Run(); err != nil {
		log.Println("error running server:", err)
	}
}
