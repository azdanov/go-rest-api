package db

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
)

func (db *Database) Migrate(path string) error {
	db.logger.Info("migrating database...")

	d, err := postgres.WithInstance(db.Client.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+path, "postgres", d)
	if err != nil {
		return fmt.Errorf("failed to create new migration instance: %w", err)
	}

	if err = m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			db.logger.Info("no new migrations to apply")
			return nil
		}
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}
