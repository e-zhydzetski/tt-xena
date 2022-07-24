package xena_test

import (
	"testing"
	"time"

	xena "github.com/e-zhydzetski/tt-xena"

	v0 "github.com/e-zhydzetski/tt-xena/internal/v0"
)

func BenchmarkHandlers(b *testing.B) {
	const (
		deduplicateWindow    = time.Second * 5
		jpsMin               = 10
		jpsMax               = 100
		duplicateProbability = 5
	)

	t := xena.SetupTestSuite(b, deduplicateWindow, jpsMin, jpsMax, duplicateProbability)

	h0 := v0.NewHandler(deduplicateWindow, jpsMax)
	b.Run("v0", t.PerformTest(h0))
}
