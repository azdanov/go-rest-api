package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
)

type Database struct {
	Client *sqlx.DB
}

func NewDatabase() (*Database, error) {
	connectionString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSL_MODE"),
	)

	db, err := sqlx.Connect("postgres", connectionString)
	if err != nil {
		return &Database{}, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &Database{
		Client: db,
	}, nil
}

func (db *Database) Ping(ctx context.Context) error {
	if err := db.Client.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	return nil
}
