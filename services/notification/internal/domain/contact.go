package domain

import "time"

type Contact struct {
	UserID   int64
	Username string
	ChatID   string
	Updated  time.Time
}
