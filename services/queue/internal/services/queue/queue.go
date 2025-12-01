package queue

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/s1lentmol/q-flow-backend/services/queue/internal/domain/models"
)

var (
	ErrGroupMismatch = errors.New("queue does not belong to group")
	ErrForbidden     = errors.New("forbidden")
	ErrSlotRequired  = errors.New("slot time required for slots mode")
	ErrQueueInactive = errors.New("queue is not active")
)

type Storage interface {
	ListQueues(ctx context.Context, group string) ([]models.Queue, error)
	CreateQueue(ctx context.Context, q models.Queue) (models.Queue, error)
	GetQueue(ctx context.Context, queueID int64) (models.Queue, []models.Participant, error)
	UpdateStatus(ctx context.Context, queueID int64, status models.QueueStatus) error
	DeleteQueue(ctx context.Context, queueID int64) error
	AddParticipant(ctx context.Context, queue models.Queue, userID int64, slotTime *time.Time) (int32, error)
	RemoveParticipant(ctx context.Context, queue models.Queue, userID int64) error
	Advance(ctx context.Context, queue models.Queue) (models.Participant, error)
}

type Notifier interface {
	NotifyPositionSoon(ctx context.Context, userID int64, queueTitle string, position int32) error
}

type Service struct {
	log     *slog.Logger
	storage Storage
	notif   Notifier
}

func New(log *slog.Logger, storage Storage, notif Notifier) *Service {
	return &Service{log: log, storage: storage, notif: notif}
}

func (s *Service) ListQueues(ctx context.Context, group string) ([]models.Queue, error) {
	return s.storage.ListQueues(ctx, group)
}

func (s *Service) CreateQueue(ctx context.Context, title, description, group string, mode models.QueueMode, ownerID int64) (models.Queue, error) {
	q := models.Queue{
		Title:       title,
		Description: description,
		Mode:        mode,
		Status:      models.StatusActive,
		GroupCode:   group,
		OwnerID:     ownerID,
	}
	return s.storage.CreateQueue(ctx, q)
}

func (s *Service) GetQueue(ctx context.Context, queueID int64, group string) (models.Queue, []models.Participant, error) {
	q, parts, err := s.storage.GetQueue(ctx, queueID)
	if err != nil {
		return models.Queue{}, nil, err
	}
	if q.GroupCode != group {
		return models.Queue{}, nil, ErrGroupMismatch
	}
	return q, parts, nil
}

func (s *Service) JoinQueue(ctx context.Context, queueID, userID int64, group string, slotTimeStr string) (int32, error) {
	queue, _, err := s.storage.GetQueue(ctx, queueID)
	if err != nil {
		return 0, err
	}
	if queue.GroupCode != group {
		return 0, ErrGroupMismatch
	}
	if queue.Status != models.StatusActive {
		return 0, ErrQueueInactive
	}

	var slotTimePtr *time.Time
	if queue.Mode == models.ModeSlots {
		if slotTimeStr == "" {
			return 0, ErrSlotRequired
		}
		t, err := time.Parse(time.RFC3339, slotTimeStr)
		if err != nil {
			return 0, fmt.Errorf("invalid slot_time: %w", err)
		}
		slotTimePtr = &t
	}

	position, err := s.storage.AddParticipant(ctx, queue, userID, slotTimePtr)
	if err != nil {
		return 0, err
	}
	if position <= 3 {
		if err := s.notif.NotifyPositionSoon(ctx, userID, queue.Title, position); err != nil {
			s.log.Warn("failed to send notification", slog.Any("err", err))
		}
	}
	return position, nil
}

func (s *Service) LeaveQueue(ctx context.Context, queueID, userID int64, group string) error {
	queue, _, err := s.storage.GetQueue(ctx, queueID)
	if err != nil {
		return err
	}
	if queue.GroupCode != group {
		return ErrGroupMismatch
	}

	return s.storage.RemoveParticipant(ctx, queue, userID)
}

func (s *Service) AdvanceQueue(ctx context.Context, queueID int64, actorID int64, group string) (models.Participant, error) {
	queue, _, err := s.storage.GetQueue(ctx, queueID)
	if err != nil {
		return models.Participant{}, err
	}
	if queue.GroupCode != group {
		return models.Participant{}, ErrGroupMismatch
	}
	if queue.OwnerID != actorID {
		return models.Participant{}, ErrForbidden
	}
	if queue.Status != models.StatusActive {
		return models.Participant{}, ErrQueueInactive
	}

	removed, err := s.storage.Advance(ctx, queue)
	if err != nil {
		return models.Participant{}, err
	}
	// notify new first participant if exists
	_, participants, err := s.storage.GetQueue(ctx, queueID)
	if err == nil && len(participants) > 0 {
		p := participants[0]
		if err := s.notif.NotifyPositionSoon(ctx, p.UserID, queue.Title, p.Position); err != nil {
			s.log.Warn("failed to send notification", slog.Any("err", err))
		}
	}
	return removed, nil
}

func (s *Service) RemoveParticipant(ctx context.Context, queueID int64, userID int64, actorID int64, group string) error {
	queue, _, err := s.storage.GetQueue(ctx, queueID)
	if err != nil {
		return err
	}
	if queue.GroupCode != group {
		return ErrGroupMismatch
	}
	if queue.OwnerID != actorID {
		return ErrForbidden
	}
	if err := s.storage.RemoveParticipant(ctx, queue, userID); err != nil {
		return err
	}
	_, participants, err := s.storage.GetQueue(ctx, queueID)
	if err == nil && len(participants) > 0 {
		p := participants[0]
		if err := s.notif.NotifyPositionSoon(ctx, p.UserID, queue.Title, p.Position); err != nil {
			s.log.Warn("failed to send notification", slog.Any("err", err))
		}
	}
	return nil
}

func (s *Service) ArchiveQueue(ctx context.Context, queueID int64, actorID int64, group string) error {
	queue, _, err := s.storage.GetQueue(ctx, queueID)
	if err != nil {
		return err
	}
	if queue.GroupCode != group {
		return ErrGroupMismatch
	}
	if queue.OwnerID != actorID {
		return ErrForbidden
	}
	return s.storage.UpdateStatus(ctx, queueID, models.StatusArchived)
}

func (s *Service) DeleteQueue(ctx context.Context, queueID int64, actorID int64, group string) error {
	queue, _, err := s.storage.GetQueue(ctx, queueID)
	if err != nil {
		return err
	}
	if queue.GroupCode != group {
		return ErrGroupMismatch
	}
	if queue.OwnerID != actorID {
		return ErrForbidden
	}
	return s.storage.DeleteQueue(ctx, queueID)
}
