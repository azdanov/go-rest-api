package main

import (
	"context"
	"fmt"
	"log"

	"github.com/azdanov/go-rest-api/internal/db"
	_ "github.com/lib/pq"
)

func Run() error {
	log.Println("running server...")

	ctx := context.Background()

	db, err := db.NewDatabase()
	if err != nil {
		log.Println("error creating database:", err)
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	if err = db.Ping(ctx); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := Run(); err != nil {
		log.Println("error running server:", err)
	}
}
