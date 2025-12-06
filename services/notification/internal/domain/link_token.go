package domain

import "time"

type LinkToken struct {
	Token    string
	UserID   int64
	Username string
	Created  time.Time
	UsedAt   *time.Time
}
