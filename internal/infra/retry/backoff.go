package retry

import (
	"context"
	"math/rand"
	"time"
)

type Backoff struct {
	policy  Policy
	current time.Duration
	rng     *rand.Rand
}

func NewBackoff(policy Policy) *Backoff {
	normalized := policy.normalized()
	return &Backoff{
		policy:  normalized,
		current: normalized.BaseDelay,
		rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (b *Backoff) Reset() {
	if b == nil {
		return
	}
	b.current = b.policy.BaseDelay
}

func (b *Backoff) Next() time.Duration {
	if b == nil {
		return 0
	}
	current := b.current
	next := time.Duration(float64(b.current) * b.policy.Factor)
	if next <= 0 {
		next = b.policy.BaseDelay
	}
	if next > b.policy.MaxDelay {
		next = b.policy.MaxDelay
	}
	b.current = next
	return b.jitter(current)
}

func (b *Backoff) Sleep(ctx context.Context) bool {
	if b == nil {
		return false
	}
	delay := b.Next()
	if delay <= 0 {
		return true
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func (b *Backoff) jitter(base time.Duration) time.Duration {
	if b.policy.Jitter == 0 {
		return base
	}
	delta := float64(base) * b.policy.Jitter
	variation := (b.rng.Float64()*2 - 1) * delta
	result := time.Duration(float64(base) + variation)
	if result < 0 {
		return 0
	}
	return result
}
