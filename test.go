package xena

import (
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func SetupTestSuite(b *testing.B, dedupWindow time.Duration, jpsMin, jpsMax, dupProb int) *TestSuite {
	seed := time.Now().UnixNano()
	if s, ok := os.LookupEnv("RAND_SEED"); ok {
		seed, _ = strconv.ParseInt(s, 10, 64)
	}
	b.Log("Seed:", seed)
	rand.Seed(seed)

	if jpsMin >= jpsMax {
		b.Fatal("Jobs rate invalid")
	}

	if dupProb < 0 || dupProb > 100 {
		b.Fatal("Duplicate probability should be in [0;100]")
	}

	return &TestSuite{
		dedupWindowNanos: dedupWindow.Nanoseconds(),
		jpsMin:           int64(jpsMin),
		jpsMax:           int64(jpsMax),
		dupProb:          dupProb,
	}
}

type TestSuite struct {
	curTimeNanos int64 // virtual now() in nanos, start from 0, only relative value make difference

	dedupWindowNanos int64 // deduplication windows in nanos
	jpsMin           int64 // min jobs per second
	jpsMax           int64 // max jobs per second
	dupProb          int   // desired duplicate probability in percents

	activeJobIdx int
	jobs         []duplicateAwareJob // jobs cache, should be the same for all tests

	leftWindowBorderIdx int // index of latest non-duplicate job in deduplication interval, not reset, its ID used when duplicate needed

	errorsTotal uint
}

type duplicateAwareJob struct {
	job       Job
	duplicate bool
}

func (t *TestSuite) PerformTest(hf func() Handler) func(b *testing.B) {
	t.reset()

	initAlloc := getTotalAllocBytes()
	handler := hf()
	constructionAlloc := getTotalAllocBytes() - initAlloc

	once := sync.Once{}
	return func(b *testing.B) {
		once.Do(func() {
			b.Logf("Construction heap size: %d bytes", constructionAlloc)
		})
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			j := t.nextJob()
			b.StartTimer()

			res := handler.Handle(j)

			b.StopTimer()
			t.recordJobResult(res)
			b.StartTimer()
		}
		b.StopTimer()
		if b.N > 1 { // skip first launch report
			t.printSessionReport(b.Logf)
		}
	}
}

func getTotalAllocBytes() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.TotalAlloc
}

func (t *TestSuite) nextJob() Job {
	t.activeJobIdx++
	if len(t.jobs) <= t.activeJobIdx { // out of cache
		ts := t.randomTimeShift()
		t.curTimeNanos += ts

		j := duplicateAwareJob{ // make unique job by default
			job: Job{
				ID:        uuid.New(),
				Timestamp: t.curTimeNanos,
			},
			duplicate: false,
		}

		if t.randomDuplicate() { // transform to duplicate job if possible
			leftBorder := t.curTimeNanos - t.dedupWindowNanos
			for t.leftWindowBorderIdx < t.activeJobIdx {
				// move window, skip duplicates
				// duplicate can't be used to create another duplicate, as it should be ignored by handler
				if leftBorder > t.jobs[t.leftWindowBorderIdx].job.Timestamp || t.jobs[t.leftWindowBorderIdx].duplicate {
					t.leftWindowBorderIdx++
					continue
				}
				break
			}
			if t.leftWindowBorderIdx < t.activeJobIdx {
				j.duplicate = true
				j.job.ID = t.jobs[t.leftWindowBorderIdx].job.ID // TODO maybe get random non-duplicate job ID in [leftWindowBorderIdx;activeJobIdx)
			}
		}

		t.jobs = append(t.jobs, j)
	}
	return t.jobs[t.activeJobIdx].job
}

//nolint:gosec // unsecure rand is ok here
func (t *TestSuite) randomTimeShift() int64 {
	jps := rand.Int63n(t.jpsMax-t.jpsMin) + t.jpsMin
	return int64(time.Second) / jps
}

//nolint:gosec // unsecure rand is ok here
func (t *TestSuite) randomDuplicate() bool {
	return rand.Intn(100) < t.dupProb
}

func (t *TestSuite) recordJobResult(res error) {
	var expected error
	if t.jobs[t.activeJobIdx].duplicate {
		expected = ErrDuplicate
	}
	if expected != res {
		t.errorsTotal++
	}
}

func (t *TestSuite) reset() {
	t.activeJobIdx = -1
	t.errorsTotal = 0
}

func (t *TestSuite) printSessionReport(logf func(format string, args ...any)) {
	logf("Jobs: %d, errors: %d", t.activeJobIdx+1, t.errorsTotal)
	logf("Avg rate: %f j/s", float64(t.activeJobIdx+1)/time.Duration(t.curTimeNanos).Seconds())
}
