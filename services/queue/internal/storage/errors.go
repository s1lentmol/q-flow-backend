package storage

import "errors"

var (
	ErrQueueNotFound      = errors.New("queue not found")
	ErrParticipantExists  = errors.New("participant already in queue")
	ErrParticipantMissing = errors.New("participant not found in queue")
	ErrNotOwner           = errors.New("not a queue owner")
)
