package xena_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"

	xena "github.com/e-zhydzetski/tt-xena"
	v0 "github.com/e-zhydzetski/tt-xena/internal/v0"
)

func BenchmarkHandlers(b *testing.B) {
	t := SetupTest(b.Log, 200)

	h0 := v0.NewHandler()
	b.Run("v0", t.PerformTest(h0))
}

func SetupTest(log func(args ...any), jpsAvg int) *Test {
	seed := time.Now().UnixNano()
	log("Seed:", seed)
	rand.Seed(seed)

	return &Test{
		jpsAvg: jpsAvg,
	}
}

type Test struct {
	jpsAvg int

	cur     int
	jobs    []xena.CompressedJob
	results []error

	errorsTotal uint
}

func (t *Test) PerformTest(handler xena.Handler) func(b *testing.B) {
	t.reset()
	return func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			j := t.nextJob()
			b.StartTimer()

			res := handler.Handle(j)

			b.StopTimer()
			t.recordJobResult(res)
			b.StartTimer()
		}
		t.report(b.Logf)
	}
}

func (t *Test) nextJob() xena.CompressedJob {
	t.cur++
	if len(t.jobs) <= t.cur {
		j := xena.Job{
			ID:        uuid.New(),
			Timestamp: time.Now(),
		}
		t.jobs = append(t.jobs, j.Compress())
		t.results = append(t.results, nil)
	}
	return t.jobs[t.cur]
}

func (t *Test) recordJobResult(res error) {
	expected := t.results[t.cur]
	if expected != res {
		t.errorsTotal++
	}
}

func (t *Test) reset() {
	t.cur = -1
	t.errorsTotal = 0
}

func (t *Test) report(logf func(format string, args ...any)) {
	logf("Jobs: %d, errors: %d\n", t.cur+1, t.errorsTotal)
}
