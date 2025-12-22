package telemetry

import (
	"time"

	"mcpd/internal/domain"
)

type NoopMetrics struct{}

func NewNoopMetrics() *NoopMetrics {
	return &NoopMetrics{}
}

func (n *NoopMetrics) ObserveRoute(serverType string, duration time.Duration, err error) {}

func (n *NoopMetrics) ObserveInstanceStart(serverType string, duration time.Duration, err error) {}

func (n *NoopMetrics) ObserveInstanceStop(serverType string, err error) {}

func (n *NoopMetrics) SetActiveInstances(serverType string, count int) {}

var _ domain.Metrics = (*NoopMetrics)(nil)
