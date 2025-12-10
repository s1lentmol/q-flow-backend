package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/s1lentmol/q-flow-backend/services/auth/internal/domain/models"
	"github.com/s1lentmol/q-flow-backend/services/auth/internal/storage"
)

type Storage struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*Storage, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: connect: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}

	return &Storage{pool: pool}, nil
}

func (s *Storage) Close() {
	s.pool.Close()
}

func (s *Storage) SaveUser(ctx context.Context, email string, fullName string, passHash []byte) (int64, error) {
	const query = `INSERT INTO users (email, full_name, pass_hash) VALUES ($1, $2, $3) RETURNING id`

	var id int64
	if err := s.pool.QueryRow(ctx, query, email, fullName, passHash).Scan(&id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return 0, storage.ErrUserExists
		}
		return 0, fmt.Errorf("postgres: save user: %w", err)
	}

	return id, nil
}

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const query = `SELECT id, email, full_name, pass_hash FROM users WHERE email = $1`

	var user models.User
	if err := s.pool.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.FullName, &user.PassHash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, storage.ErrUserNotFound
		}
		return models.User{}, fmt.Errorf("postgres: get user: %w", err)
	}

	return user, nil
}

func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const query = `SELECT is_admin FROM users WHERE id = $1`

	var isAdmin bool
	if err := s.pool.QueryRow(ctx, query, userID).Scan(&isAdmin); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, storage.ErrUserNotFound
		}
		return false, fmt.Errorf("postgres: check admin: %w", err)
	}

	return isAdmin, nil
}

func (s *Storage) App(ctx context.Context, appID int) (models.App, error) {
	const query = `SELECT id, name, secret FROM apps WHERE id = $1`

	var app models.App
	if err := s.pool.QueryRow(ctx, query, appID).Scan(&app.ID, &app.Name, &app.Secret); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.App{}, storage.ErrAppNotFound
		}
		return models.App{}, fmt.Errorf("postgres: get app: %w", err)
	}

	return app, nil
}
