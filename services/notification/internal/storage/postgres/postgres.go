package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/s1lentmol/q-flow-backend/services/notification/internal/domain"
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

func (s *Storage) UpsertContact(ctx context.Context, contact domain.Contact) error {
	const query = `
INSERT INTO user_contacts (user_id, telegram_username, chat_id, updated_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (user_id) DO UPDATE SET telegram_username = EXCLUDED.telegram_username, chat_id = EXCLUDED.chat_id, updated_at = NOW()
`
	_, err := s.pool.Exec(ctx, query, contact.UserID, contact.Username, contact.ChatID)
	if err != nil {
		return fmt.Errorf("postgres: upsert contact: %w", err)
	}
	return nil
}

func (s *Storage) GetContact(ctx context.Context, userID int64) (domain.Contact, error) {
	const query = `SELECT user_id, telegram_username, chat_id, updated_at FROM user_contacts WHERE user_id = $1`
	var c domain.Contact
	if err := s.pool.QueryRow(ctx, query, userID).Scan(&c.UserID, &c.Username, &c.ChatID, &c.Updated); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Contact{}, fmt.Errorf("contact not found")
		}
		return domain.Contact{}, fmt.Errorf("postgres: get contact: %w", err)
	}
	return c, nil
}
