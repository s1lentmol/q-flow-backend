package models

import "time"

type QueueMode string

const (
	ModeLive    QueueMode = "live"
	ModeManaged QueueMode = "managed"
	ModeRandom  QueueMode = "random"
	ModeSlots   QueueMode = "slots"
)

type QueueStatus string

const (
	StatusActive   QueueStatus = "active"
	StatusArchived QueueStatus = "archived"
)

type Queue struct {
	ID          int64
	Title       string
	Description string
	Mode        QueueMode
	Status      QueueStatus
	GroupCode   string
	OwnerID     int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
