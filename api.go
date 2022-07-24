package xena

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Job struct {
	ID        uuid.UUID
	Timestamp time.Time
}

type CompressedJob struct {
	ID        [16]byte
	Timestamp int64
}

var ErrDuplicate = errors.New("duplicate detected")

type Handler interface {
	Handle(job CompressedJob) error
}
