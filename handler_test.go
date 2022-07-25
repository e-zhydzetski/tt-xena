package xena_test

import (
	"errors"
	"testing"
	"time"

	v2 "github.com/e-zhydzetski/tt-xena/internal/v2"

	v1 "github.com/e-zhydzetski/tt-xena/internal/v1"

	xena "github.com/e-zhydzetski/tt-xena"

	v0 "github.com/e-zhydzetski/tt-xena/internal/v0"
)

func BenchmarkHandlers(b *testing.B) {
	const (
		deduplicateWindow    = time.Second * 5
		jpsMin               = 10
		jpsMax               = 1000
		duplicateProbability = 5
	)

	t := xena.SetupTestSuite(b, deduplicateWindow, jpsMin, jpsMax, duplicateProbability)

	// to build test suite cache, to minimize heap affection on real tests
	b.Run("warm up", t.PerformTest(func() xena.Handler {
		return NoopHandler{}
	}))

	b.Run("v0", t.PerformTest(func() xena.Handler {
		return v0.NewHandler(deduplicateWindow, jpsMax)
	}))

	b.Run("v1", t.PerformTest(func() xena.Handler {
		return v1.NewHandler(deduplicateWindow)
	}))

	b.Run("v2", t.PerformTest(func() xena.Handler {
		return v2.NewHandler(deduplicateWindow)
	}))
}

type NoopHandler struct {
}

var ErrNoop = errors.New("noop")

func (n NoopHandler) Handle(_ xena.Job) error {
	return ErrNoop
}
