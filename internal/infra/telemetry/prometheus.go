package telemetry

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"mcpd/internal/domain"
)

type PrometheusMetrics struct {
	routeDuration   *prometheus.HistogramVec
	instanceStarts  *prometheus.CounterVec
	instanceStops   *prometheus.CounterVec
	activeInstances *prometheus.GaugeVec
}

func NewPrometheusMetrics() *PrometheusMetrics {
	return &PrometheusMetrics{
		routeDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "mcpd_route_duration_seconds",
				Help:    "Duration of route requests in seconds",
				Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"server_type", "status"},
		),
		instanceStarts: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "mcpd_instance_starts_total",
				Help: "Total number of instance start attempts",
			},
			[]string{"server_type"},
		),
		instanceStops: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "mcpd_instance_stops_total",
				Help: "Total number of instance stops",
			},
			[]string{"server_type"},
		),
		activeInstances: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "mcpd_active_instances",
				Help: "Current number of active instances",
			},
			[]string{"server_type"},
		),
	}
}

func (p *PrometheusMetrics) ObserveRoute(serverType string, duration time.Duration, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}
	p.routeDuration.WithLabelValues(serverType, status).Observe(duration.Seconds())
}

func (p *PrometheusMetrics) ObserveInstanceStart(serverType string, duration time.Duration, err error) {
	p.instanceStarts.WithLabelValues(serverType).Inc()
}

func (p *PrometheusMetrics) ObserveInstanceStop(serverType string, err error) {
	p.instanceStops.WithLabelValues(serverType).Inc()
}

func (p *PrometheusMetrics) SetActiveInstances(serverType string, count int) {
	p.activeInstances.WithLabelValues(serverType).Set(float64(count))
}

var _ domain.Metrics = (*PrometheusMetrics)(nil)
