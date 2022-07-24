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
	t := SetupTest(b, 100, 500)

	h0 := v0.NewHandler()
	b.Run("v0", t.PerformTest(h0))
}

func SetupTest(b *testing.B, jpsMin int, jpsMax int) *Test {
	seed := time.Now().UnixNano()
	if s, ok := os.LookupEnv("RAND_SEED"); ok {
		seed, _ = strconv.ParseInt(s, 10, 64)
	}
	b.Log("Seed:", seed)
	r := rand.New(rand.NewSource(seed))

	if jpsMin >= jpsMax {
		b.Fatal("Jobs rate invalid")
	}
	timeShiftMinNanos := int64(time.Second) / int64(jpsMax)
	timeShiftMaxNanos := int64(time.Second) / int64(jpsMin)

	return &Test{
		random:            r,
		timeShiftMinNanos: timeShiftMinNanos,
		timeShiftMaxNanos: timeShiftMaxNanos,
	}
}

type Test struct {
	random *rand.Rand

	curTimeNanos      int64 // start from 0
	timeShiftMaxNanos int64
	timeShiftMinNanos int64

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
		t.curTimeNanos = t.curTimeNanos + t.randomTimeShift()
		j := xena.Job{
			ID:        uuid.New(), // TODO add random duplicates
			Timestamp: t.curTimeNanos,
		}
		t.jobs = append(t.jobs, j)
		t.results = append(t.results, nil)
	}
	return t.jobs[t.activeJobIdx]
}

func (t *Test) randomTimeShift() int64 {
	return t.random.Int63n(t.timeShiftMaxNanos-t.timeShiftMinNanos) + t.timeShiftMinNanos
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
