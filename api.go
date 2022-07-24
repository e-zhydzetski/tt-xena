package xena

import (
	"errors"
	"github.com/google/uuid"
)

type Job struct {
	ID        uuid.UUID
	Timestamp int64
}

var ErrDuplicate = errors.New("duplicate detected")

type Handler interface {
	Handle(job Job) error
}
