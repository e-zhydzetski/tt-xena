package v1

import (
	"github.com/google/uuid"
	"time"

	xena "github.com/e-zhydzetski/tt-xena"
)

func NewHandler(dedupWindow time.Duration) *Handler {
	return &Handler{
		dedupWindowNanos: dedupWindow.Nanoseconds(),
		ids:              map[uuid.UUID]struct{}{},
	}
}

type Handler struct {
	dedupWindowNanos int64
	dedupQueue       Queue[xena.Job]
	ids              map[uuid.UUID]struct{}
}

func (h *Handler) Handle(job xena.Job) error {
	now := job.Timestamp
	leftTimeBorder := now - h.dedupWindowNanos

	t, exists := h.dedupQueue.TailValue()
	for exists {
		if t.Timestamp >= leftTimeBorder {
			break
		}
		delete(h.ids, t.ID)
		h.dedupQueue.CutTail()
		t, exists = h.dedupQueue.TailValue()
	}

	if _, exists := h.ids[job.ID]; exists {
		return xena.ErrDuplicate
	}

	h.dedupQueue.PushHead(job)
	h.ids[job.ID] = struct{}{}

	return nil
}
