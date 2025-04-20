package main

import (
	"fmt"
	"log"

	"github.com/azdanov/go-rest-api/internal/db"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func Run() error {
	log.Println("running server...")

	db, err := db.NewDatabase()
	if err != nil {
		log.Println("error creating database:", err)
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	if err = db.Migrate(); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := Run(); err != nil {
		log.Println("error running server:", err)
	}
}
