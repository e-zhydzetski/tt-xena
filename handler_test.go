package xena_test

import (
	"testing"
	"time"

	xena "github.com/e-zhydzetski/tt-xena"

	v0 "github.com/e-zhydzetski/tt-xena/internal/v0"
)

func BenchmarkHandlers(b *testing.B) {
	t := xena.SetupTestSuite(b, time.Second*5, 10, 100, 30)

	h0 := v0.NewHandler()
	b.Run("v0", t.PerformTest(h0))
}
