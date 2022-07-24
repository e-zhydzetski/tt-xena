package xena_test

import (
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"

	xena "github.com/e-zhydzetski/tt-xena"
	v0 "github.com/e-zhydzetski/tt-xena/internal/v0"
)

func BenchmarkHandlers(b *testing.B) {
	t := SetupTest(b, 10, 1000)

	h0 := v0.NewHandler()
	b.Run("v0", t.PerformTest(h0))
}

func SetupTest(b *testing.B, jpsMin, jpsMax int) *Test {
	seed := time.Now().UnixNano()
	if s, ok := os.LookupEnv("RAND_SEED"); ok {
		seed, _ = strconv.ParseInt(s, 10, 64)
	}
	b.Log("Seed:", seed)
	rand.Seed(seed)

	if jpsMin >= jpsMax {
		b.Fatal("Jobs rate invalid")
	}

	return &Test{
		jpsMin: int64(jpsMin),
		jpsMax: int64(jpsMax),
	}
}

type Test struct {
	curTimeNanos int64 // start from 0
	jpsMin       int64
	jpsMax       int64

	activeJobIdx int
	jobs         []xena.Job
	results      []error

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

func (t *Test) nextJob() xena.Job {
	t.activeJobIdx++
	if len(t.jobs) <= t.activeJobIdx {
		ts := t.randomTimeShift()
		t.curTimeNanos += ts
		j := xena.Job{
			ID:        uuid.New(), // TODO add random duplicates
			Timestamp: t.curTimeNanos,
		}
		t.jobs = append(t.jobs, j)
		t.results = append(t.results, nil)
	}
	return t.jobs[t.activeJobIdx]
}

//nolint:gosec // unsecure rand is ok here
func (t *Test) randomTimeShift() int64 {
	jps := rand.Int63n(t.jpsMax-t.jpsMin) + t.jpsMin
	return int64(time.Second) / jps
}

func (t *Test) recordJobResult(res error) {
	expected := t.results[t.activeJobIdx]
	if expected != res {
		t.errorsTotal++
	}
}

func (t *Test) reset() {
	t.activeJobIdx = -1
	t.errorsTotal = 0
}

func (t *Test) report(logf func(format string, args ...any)) {
	logf("Jobs: %d, errors: %d\n", t.activeJobIdx+1, t.errorsTotal)
	logf("Avg rate: %f j/s", float64(t.activeJobIdx+1)/time.Duration(t.curTimeNanos).Seconds())
}
