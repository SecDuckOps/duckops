package compression

import (
	"sync"
	"sync/atomic"
	"time"
)

type MetricsTracker struct {
	mu              sync.Mutex
	totalOriginal   int64
	totalCompressed int64
	totalDuration   int64
	totalCalls      int64
	totalDuplicates int64
	totalSummaries  int64
}

func NewMetricsTracker() *MetricsTracker {
	return &MetricsTracker{}
}

func (m *MetricsTracker) Record(original, compressed int, mode Mode, duration time.Duration) CompressionMetrics {
	atomic.AddInt64(&m.totalOriginal, int64(original))
	atomic.AddInt64(&m.totalCompressed, int64(compressed))
	atomic.AddInt64(&m.totalDuration, duration.Microseconds())
	atomic.AddInt64(&m.totalCalls, 1)

	ratio := 0.0
	if original > 0 {
		ratio = float64(original-compressed) / float64(original)
		if ratio < 0 {
			ratio = 0
		}
	}

	costSaved := float64(original-compressed) / 1000 * CostPerThousandTokens

	return CompressionMetrics{
		OriginalTokens:   original,
		CompressedTokens: compressed,
		Ratio:            ratio,
		Mode:             mode,
		Duration:         duration,
		CostSaved:        costSaved,
	}
}

func (m *MetricsTracker) RecordDuplicates(count int) {
	atomic.AddInt64(&m.totalDuplicates, int64(count))
}

func (m *MetricsTracker) RecordSummaries(count int) {
	atomic.AddInt64(&m.totalSummaries, int64(count))
}

func (m *MetricsTracker) Snapshot() CompressionMetrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	totalCalls := atomic.LoadInt64(&m.totalCalls)
	totalOrig := atomic.LoadInt64(&m.totalOriginal)
	totalComp := atomic.LoadInt64(&m.totalCompressed)
	totalDur := atomic.LoadInt64(&m.totalDuration)

	avgDuration := time.Duration(0)
	if totalCalls > 0 {
		avgDuration = time.Duration(totalDur/totalCalls) * time.Microsecond
	}

	ratio := 0.0
	if totalOrig > 0 {
		ratio = float64(totalOrig-totalComp) / float64(totalOrig)
		if ratio < 0 {
			ratio = 0
		}
	}

	costSaved := float64(totalOrig-totalComp) / 1000 * CostPerThousandTokens

	return CompressionMetrics{
		OriginalTokens:   int(totalOrig),
		CompressedTokens: int(totalComp),
		Ratio:            ratio,
		Duration:         avgDuration,
		CostSaved:        costSaved,
	}
}

func (m *MetricsTracker) Reset() {
	atomic.StoreInt64(&m.totalOriginal, 0)
	atomic.StoreInt64(&m.totalCompressed, 0)
	atomic.StoreInt64(&m.totalDuration, 0)
	atomic.StoreInt64(&m.totalCalls, 0)
	atomic.StoreInt64(&m.totalDuplicates, 0)
	atomic.StoreInt64(&m.totalSummaries, 0)
}
