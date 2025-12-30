package domain

import (
	"sync/atomic"
	"time"
)

type InstanceCallStats struct {
	TotalCalls    int64
	TotalErrors   int64
	TotalDuration time.Duration
	LastCallAt    time.Time
}

func (i *Instance) RecordCall(duration time.Duration, err error) {
	atomic.AddInt64(&i.callCount, 1)
	atomic.AddInt64(&i.totalDurationNs, duration.Nanoseconds())
	atomic.StoreInt64(&i.lastCallUnixNano, time.Now().UnixNano())
	if err != nil {
		atomic.AddInt64(&i.errorCount, 1)
	}
}

func (i *Instance) CallStats() InstanceCallStats {
	totalCalls := atomic.LoadInt64(&i.callCount)
	totalErrors := atomic.LoadInt64(&i.errorCount)
	totalDuration := time.Duration(atomic.LoadInt64(&i.totalDurationNs))
	lastCallUnixNano := atomic.LoadInt64(&i.lastCallUnixNano)
	var lastCallAt time.Time
	if lastCallUnixNano > 0 {
		lastCallAt = time.Unix(0, lastCallUnixNano)
	}
	return InstanceCallStats{
		TotalCalls:    totalCalls,
		TotalErrors:   totalErrors,
		TotalDuration: totalDuration,
		LastCallAt:    lastCallAt,
	}
}
