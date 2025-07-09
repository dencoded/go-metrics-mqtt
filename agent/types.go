package agent

import (
	"runtime/metrics"
)

const (
	defaultBufferSize = 4096
)

type timestampedSample struct {
	sample    metrics.Sample
	timestamp int64
}
