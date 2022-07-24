package v0

import (
	"fmt"
	xena "github.com/e-zhydzetski/tt-xena"
	"github.com/google/uuid"
	"math"
	"time"
)

func NewHandler(dedupWindow time.Duration, jpsMax int) *Handler {
	cleanerBufferSize := int(math.Ceil(dedupWindow.Seconds() * float64(jpsMax)))
	return &Handler{
		dedupWindowNanos: dedupWindow.Nanoseconds(),
		jobTimestampByID: map[uuid.UUID]int64{},
		cleanerBuffer:    make([]uuid.UUID, cleanerBufferSize),
	}
}

type Handler struct {
	dedupWindowNanos int64
	jobTimestampByID map[uuid.UUID]int64
	cleanerBuffer    []uuid.UUID
	cleanIdx         int
}

var zeroUUID = uuid.UUID{}

func (h *Handler) Handle(job xena.Job) error {
	now := job.Timestamp

	t, exists := h.jobTimestampByID[job.ID]
	if exists && now-h.dedupWindowNanos <= t { // duplicate, just ignore and error
		return xena.ErrDuplicate
	}

	h.jobTimestampByID[job.ID] = job.Timestamp

	h.cleanIdx = (h.cleanIdx + 1) % len(h.cleanerBuffer)
	oldId := h.cleanerBuffer[h.cleanIdx]
	if oldId != zeroUUID {
		oldTimestamp, exists := h.jobTimestampByID[oldId]
		if exists && oldTimestamp > now-h.dedupWindowNanos {
			fmt.Println("WARNING! Job removed within deduplication window")
			// we can't do anything as cleanerBuffer has fixed size
		}
		delete(h.jobTimestampByID, oldId)
	}
	h.cleanerBuffer[h.cleanIdx] = job.ID

	return nil
}
