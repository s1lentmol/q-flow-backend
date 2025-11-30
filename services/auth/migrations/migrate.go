package migrator

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"

	// pgx stdlib driver for database/sql.
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var migrations embed.FS

// Migrate applies embedded migrations to the provided Postgres URL.
func Migrate(url string) error {
	db, err := sql.Open("pgx", url)
	if err != nil {
		return fmt.Errorf("cannot connect to db: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("failed to close database connection", "error", err)
		}
	}()

	if err := db.Ping(); err != nil {
		return err
	}

	goose.SetBaseFS(migrations)
	if err = goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("cannot set migrations dialect: %w", err)
	}

	version, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("cannot get migration version: %w", err)
	}

	if err = goose.Up(db, "."); err != nil {
		if err := goose.DownTo(db, ".", version); err != nil {
			slog.Error(
				"cannot rollback migrations",
				slog.Any("error", err),
				slog.Any("try rollback to version", version),
			)
		}
		return fmt.Errorf("cannot up migrations: %w", err)
	}

	return nil
}
