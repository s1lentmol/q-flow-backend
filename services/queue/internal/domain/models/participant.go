package models

import "time"

type Participant struct {
	ID        int64
	QueueID   int64
	UserID    int64
	Position  int32
	SlotTime  *time.Time
	CreatedAt time.Time
}
