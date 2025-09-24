package testutils

import (
	"testing"
	"time"
)

func AddRPSMetricToBenchmark(b *testing.B, fn func()) {
	start := time.Now()
	fn()
	elapsed := time.Since(start)
	rps := float64(b.N) / elapsed.Seconds()
	b.ReportMetric(rps, "rps")
}
