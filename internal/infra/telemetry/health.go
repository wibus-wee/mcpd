package telemetry

import (
	"sort"
	"sync"
	"time"
)

type HealthTracker struct {
	mu     sync.RWMutex
	checks map[string]*healthCheck
}

type healthCheck struct {
	name       string
	staleAfter time.Duration
	lastBeat   time.Time
}

type Heartbeat struct {
	name    string
	tracker *HealthTracker
}

type HealthReport struct {
	Status string              `json:"status"`
	Checks []HealthCheckStatus `json:"checks"`
}

type HealthCheckStatus struct {
	Name         string    `json:"name"`
	Healthy      bool      `json:"healthy"`
	LastBeat     time.Time `json:"lastBeat"`
	StaleAfterMs int64     `json:"staleAfterMs"`
}

func NewHealthTracker() *HealthTracker {
	return &HealthTracker{
		checks: make(map[string]*healthCheck),
	}
}

func (t *HealthTracker) Register(name string, staleAfter time.Duration) *Heartbeat {
	if staleAfter <= 0 {
		staleAfter = time.Second
	}

	check := &healthCheck{
		name:       name,
		staleAfter: staleAfter,
		lastBeat:   time.Now(),
	}

	t.mu.Lock()
	t.checks[name] = check
	t.mu.Unlock()

	return &Heartbeat{name: name, tracker: t}
}

func (t *HealthTracker) Report() HealthReport {
	now := time.Now()

	report := HealthReport{Status: "ok"}

	t.mu.RLock()
	report.Checks = make([]HealthCheckStatus, 0, len(t.checks))
	for _, check := range t.checks {
		healthy := now.Sub(check.lastBeat) <= check.staleAfter
		if !healthy {
			report.Status = "stale"
		}

		report.Checks = append(report.Checks, HealthCheckStatus{
			Name:         check.name,
			Healthy:      healthy,
			LastBeat:     check.lastBeat,
			StaleAfterMs: check.staleAfter.Milliseconds(),
		})
	}
	t.mu.RUnlock()

	sort.Slice(report.Checks, func(i, j int) bool {
		return report.Checks[i].Name < report.Checks[j].Name
	})

	return report
}

func (h *Heartbeat) Beat() {
	if h == nil || h.tracker == nil {
		return
	}
	tr := h.tracker
	tr.mu.Lock()
	if check := tr.checks[h.name]; check != nil {
		check.lastBeat = time.Now()
	}
	tr.mu.Unlock()
}

func (h *Heartbeat) Stop() {
	if h == nil || h.tracker == nil {
		return
	}
	tr := h.tracker
	tr.mu.Lock()
	delete(tr.checks, h.name)
	tr.mu.Unlock()
}
