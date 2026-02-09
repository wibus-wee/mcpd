package retry

import "time"

type Policy struct {
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	Factor     float64
	Jitter     float64
	MaxRetries int
}

func (p Policy) normalized() Policy {
	if p.BaseDelay <= 0 {
		p.BaseDelay = 100 * time.Millisecond
	}
	if p.MaxDelay <= 0 {
		p.MaxDelay = p.BaseDelay
	}
	if p.Factor <= 0 {
		p.Factor = 2
	}
	if p.Factor < 1 {
		p.Factor = 1
	}
	if p.Jitter < 0 {
		p.Jitter = 0
	}
	if p.Jitter > 1 {
		p.Jitter = 1
	}
	return p
}
