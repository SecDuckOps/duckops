package platform

import (
	"runtime"
	"time"
)

var defaultCPUSample = time.Now()

func DefaultSystemMetrics() (cpu float64, memoryUsed float64, err error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	memoryUsed = float64(m.Alloc) / 1024 / 1024

	now := time.Now()
	elapsed := now.Sub(defaultCPUSample).Seconds()
	if elapsed > 0 {
		cpu = float64(runtime.NumCPU()) * elapsed * 0.5
	}
	defaultCPUSample = now

	return cpu, memoryUsed, nil
}
