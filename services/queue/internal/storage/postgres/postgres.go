package postgres

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/s1lentmol/q-flow-backend/services/queue/internal/domain/models"
	"github.com/s1lentmol/q-flow-backend/services/queue/internal/storage"
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

func (s *Storage) ListQueues(ctx context.Context, group string) ([]models.Queue, error) {
	const query = `SELECT id, title, description, mode, status, group_code, owner_id, created_at, updated_at 
FROM queues WHERE group_code = $1 AND status = 'active' ORDER BY created_at DESC`

	rows, err := s.pool.Query(ctx, query, group)
	if err != nil {
		return nil, fmt.Errorf("postgres: list queues: %w", err)
	}
	defer rows.Close()

	var queues []models.Queue
	for rows.Next() {
		var q models.Queue
		if err := rows.Scan(&q.ID, &q.Title, &q.Description, &q.Mode, &q.Status, &q.GroupCode, &q.OwnerID, &q.CreatedAt, &q.UpdatedAt); err != nil {
			return nil, fmt.Errorf("postgres: scan queue: %w", err)
		}
		queues = append(queues, q)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: rows error: %w", err)
	}

	return queues, nil
}

func (s *Storage) CreateQueue(ctx context.Context, q models.Queue) (models.Queue, error) {
	const query = `INSERT INTO queues (title, description, mode, status, group_code, owner_id)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at, updated_at`

	err := s.pool.QueryRow(ctx, query, q.Title, q.Description, q.Mode, q.Status, q.GroupCode, q.OwnerID).
		Scan(&q.ID, &q.CreatedAt, &q.UpdatedAt)
	if err != nil {
		return models.Queue{}, fmt.Errorf("postgres: create queue: %w", err)
	}

	return q, nil
}

func (s *Storage) GetQueue(ctx context.Context, queueID int64) (models.Queue, []models.Participant, error) {
	const queueQuery = `SELECT id, title, description, mode, status, group_code, owner_id, created_at, updated_at FROM queues WHERE id = $1`
	const participantQuery = `SELECT id, queue_id, user_id, position, slot_time, created_at 
FROM queue_participants WHERE queue_id = $1 ORDER BY position ASC`

	var q models.Queue
	if err := s.pool.QueryRow(ctx, queueQuery, queueID).
		Scan(&q.ID, &q.Title, &q.Description, &q.Mode, &q.Status, &q.GroupCode, &q.OwnerID, &q.CreatedAt, &q.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Queue{}, nil, storage.ErrQueueNotFound
		}
		return models.Queue{}, nil, fmt.Errorf("postgres: get queue: %w", err)
	}

	rows, err := s.pool.Query(ctx, participantQuery, queueID)
	if err != nil {
		return models.Queue{}, nil, fmt.Errorf("postgres: list participants: %w", err)
	}
	defer rows.Close()

	var participants []models.Participant
	for rows.Next() {
		var p models.Participant
		if err := rows.Scan(&p.ID, &p.QueueID, &p.UserID, &p.Position, &p.SlotTime, &p.CreatedAt); err != nil {
			return models.Queue{}, nil, fmt.Errorf("postgres: scan participant: %w", err)
		}
		participants = append(participants, p)
	}
	if err := rows.Err(); err != nil {
		return models.Queue{}, nil, fmt.Errorf("postgres: participants rows error: %w", err)
	}

	return q, participants, nil
}

func (s *Storage) UpdateStatus(ctx context.Context, queueID int64, status models.QueueStatus) error {
	const query = `UPDATE queues SET status = $1, updated_at = NOW() WHERE id = $2`
	cmd, err := s.pool.Exec(ctx, query, status, queueID)
	if err != nil {
		return fmt.Errorf("postgres: update status: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return storage.ErrQueueNotFound
	}
	return nil
}

func (s *Storage) DeleteQueue(ctx context.Context, queueID int64) error {
	const query = `DELETE FROM queues WHERE id = $1`
	cmd, err := s.pool.Exec(ctx, query, queueID)
	if err != nil {
		return fmt.Errorf("postgres: delete queue: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return storage.ErrQueueNotFound
	}
	return nil
}

func (s *Storage) AddParticipant(ctx context.Context, queue models.Queue, userID int64, slotTime *time.Time) (int32, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, fmt.Errorf("postgres: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM queue_participants WHERE queue_id=$1 AND user_id=$2)`, queue.ID, userID).Scan(&exists); err != nil {
		return 0, fmt.Errorf("postgres: check participant: %w", err)
	}
	if exists {
		return 0, storage.ErrParticipantExists
	}

	var position int32

	switch queue.Mode {
	case models.ModeLive, models.ModeManaged:
		if err := tx.QueryRow(ctx, `SELECT COALESCE(MAX(position),0)+1 FROM queue_participants WHERE queue_id=$1`, queue.ID).Scan(&position); err != nil {
			return 0, fmt.Errorf("postgres: calc position: %w", err)
		}
	case models.ModeRandom:
		var count int32
		if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM queue_participants WHERE queue_id=$1`, queue.ID).Scan(&count); err != nil {
			return 0, fmt.Errorf("postgres: count participants: %w", err)
		}
		position = randomPosition(count + 1)
		if _, err := tx.Exec(ctx, `UPDATE queue_participants SET position = position + 1 WHERE queue_id=$1 AND position >= $2`, queue.ID, position); err != nil {
			return 0, fmt.Errorf("postgres: shift positions: %w", err)
		}
	case models.ModeSlots:
		if slotTime == nil {
			return 0, fmt.Errorf("slot_time is required for slots mode")
		}
		if err := tx.QueryRow(ctx, `SELECT COALESCE(COUNT(*),0)+1 FROM queue_participants WHERE queue_id=$1 AND (slot_time <= $2 OR slot_time IS NULL)`, queue.ID, slotTime).Scan(&position); err != nil {
			return 0, fmt.Errorf("postgres: calc slot position: %w", err)
		}
		if _, err := tx.Exec(ctx, `UPDATE queue_participants SET position = position + 1 WHERE queue_id=$1 AND (slot_time >= $2 OR slot_time IS NULL)`, queue.ID, slotTime); err != nil {
			return 0, fmt.Errorf("postgres: shift slot positions: %w", err)
		}
	default:
		return 0, fmt.Errorf("unsupported queue mode: %s", queue.Mode)
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO queue_participants (queue_id, user_id, position, slot_time) VALUES ($1, $2, $3, $4)`,
		queue.ID, userID, position, slotTime); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, storage.ErrParticipantExists
		}
		return 0, fmt.Errorf("postgres: insert participant: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("postgres: commit: %w", err)
	}

	return position, nil
}

func (s *Storage) RemoveParticipant(ctx context.Context, queue models.Queue, userID int64) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("postgres: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var position int32
	if err := tx.QueryRow(ctx, `DELETE FROM queue_participants WHERE queue_id=$1 AND user_id=$2 RETURNING position`, queue.ID, userID).Scan(&position); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return storage.ErrParticipantMissing
		}
		return fmt.Errorf("postgres: delete participant: %w", err)
	}

	if _, err := tx.Exec(ctx, `UPDATE queue_participants SET position = position - 1 WHERE queue_id=$1 AND position > $2`, queue.ID, position); err != nil {
		return fmt.Errorf("postgres: shift positions after delete: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("postgres: commit: %w", err)
	}

	return nil
}

func (s *Storage) Advance(ctx context.Context, queue models.Queue) (models.Participant, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return models.Participant{}, fmt.Errorf("postgres: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	orderClause := "position ASC"
	if queue.Mode == models.ModeSlots {
		orderClause = "slot_time ASC NULLS LAST"
	}

	var p models.Participant
	query := fmt.Sprintf(`DELETE FROM queue_participants WHERE id = (
		SELECT id FROM queue_participants WHERE queue_id=$1 ORDER BY %s LIMIT 1
	) RETURNING id, queue_id, user_id, position, slot_time, created_at`, orderClause)

	if err := tx.QueryRow(ctx, query, queue.ID).Scan(&p.ID, &p.QueueID, &p.UserID, &p.Position, &p.SlotTime, &p.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Participant{}, storage.ErrParticipantMissing
		}
		return models.Participant{}, fmt.Errorf("postgres: advance delete: %w", err)
	}

	if _, err := tx.Exec(ctx, `UPDATE queue_participants SET position = position - 1 WHERE queue_id=$1 AND position > $2`, queue.ID, p.Position); err != nil {
		return models.Participant{}, fmt.Errorf("postgres: shift positions after advance: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Participant{}, fmt.Errorf("postgres: commit: %w", err)
	}

	return p, nil
}

func randomPosition(max int32) int32 {
	if max <= 1 {
		return 1
	}
	rand.Seed(time.Now().UnixNano())
	return int32(rand.Intn(int(max)) + 1)
}
