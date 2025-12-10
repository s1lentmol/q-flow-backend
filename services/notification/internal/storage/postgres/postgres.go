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
	// COALESCE avoids scan errors when columns are NULL.
	const query = `SELECT user_id, COALESCE(telegram_username,''), COALESCE(chat_id,''), updated_at FROM user_contacts WHERE user_id = $1`
	var c domain.Contact
	if err := s.pool.QueryRow(ctx, query, userID).Scan(&c.UserID, &c.Username, &c.ChatID, &c.Updated); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Contact{}, fmt.Errorf("contact not found")
		}
		return domain.Contact{}, fmt.Errorf("postgres: get contact: %w", err)
	}
	return c, nil
}

func (s *Storage) CreateLinkToken(ctx context.Context, token domain.LinkToken) error {
	const query = `
INSERT INTO telegram_link_tokens (token, user_id, telegram_username)
VALUES ($1, $2, $3)
`
	_, err := s.pool.Exec(ctx, query, token.Token, token.UserID, token.Username)
	if err != nil {
		return fmt.Errorf("postgres: create link token: %w", err)
	}
	return nil
}

func (s *Storage) ConsumeLinkToken(ctx context.Context, token string) (domain.LinkToken, error) {
	const query = `
UPDATE telegram_link_tokens
SET used_at = NOW()
WHERE token = $1 AND used_at IS NULL
RETURNING token, user_id, telegram_username, created_at, used_at
`
	var lt domain.LinkToken
	if err := s.pool.QueryRow(ctx, query, token).
		Scan(&lt.Token, &lt.UserID, &lt.Username, &lt.Created, &lt.UsedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.LinkToken{}, fmt.Errorf("token not found or already used")
		}
		return domain.LinkToken{}, fmt.Errorf("postgres: consume link token: %w", err)
	}
	return lt, nil
}
