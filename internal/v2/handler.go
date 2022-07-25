package v2

import (
	"time"

	"github.com/google/uuid"

	xena "github.com/e-zhydzetski/tt-xena"
)

func NewHandler(dedupWindow time.Duration) *Handler {
	return &Handler{
		dedupWindowNanos: dedupWindow.Nanoseconds(),
		jobTimestampByID: map[uuid.UUID]int64{},
	}
}

/*
Handler is modified version of v0.Handler with dynamic ring buffer that allows insert element instead of rewriting
*/
type Handler struct {
	dedupWindowNanos int64
	jobTimestampByID map[uuid.UUID]int64
	cleanRingBuffer  RingBuffer[uuid.UUID]
}

func (h *Handler) Handle(job xena.Job) error {
	now := job.Timestamp

	t, exists := h.jobTimestampByID[job.ID]
	if exists && now-h.dedupWindowNanos <= t { // duplicate, just ignore and error
		return xena.ErrDuplicate
	}

	h.jobTimestampByID[job.ID] = job.Timestamp

	defer h.cleanRingBuffer.Next()

	old := h.cleanRingBuffer.GetNext()
	if old == nil {
		h.cleanRingBuffer.InsertNext(job.ID)
		return nil
	}

	oldID := old.Value
	oldTimestamp, exists := h.jobTimestampByID[oldID]
	if exists && oldTimestamp > now-h.dedupWindowNanos { // next element is within deduplicate window, nothing to remove
		h.cleanRingBuffer.InsertNext(job.ID)
		return nil
	}
	delete(h.jobTimestampByID, oldID)
	old.Value = job.ID

	return nil
}
