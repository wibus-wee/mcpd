package domain

import "time"

type RouteStatus string

const (
	RouteStatusSuccess RouteStatus = "success"
	RouteStatusError   RouteStatus = "error"
)

type RouteReason string

const (
	RouteReasonSuccess          RouteReason = "success"
	RouteReasonInvalidRequest   RouteReason = "invalid_request"
	RouteReasonMethodNotAllowed RouteReason = "method_not_allowed"
	RouteReasonTimeoutColdStart RouteReason = "timeout_cold_start"
	RouteReasonTimeoutExecution RouteReason = "timeout_execution"
	RouteReasonConnClosed       RouteReason = "conn_closed"
	RouteReasonAcquireFailed    RouteReason = "acquire_failed"
	RouteReasonExecutionFailed  RouteReason = "execution_failed"
	RouteReasonUnknown          RouteReason = "unknown"
)

type RouteMetric struct {
	ServerType string
	Caller     string
	Profile    string
	Status     RouteStatus
	Reason     RouteReason
	Duration   time.Duration
}

type Metrics interface {
	ObserveRoute(metric RouteMetric)
	ObserveInstanceStart(serverType string, duration time.Duration, err error)
	ObserveInstanceStop(serverType string, err error)
	SetActiveInstances(serverType string, count int)
	SetPoolCapacityRatio(serverType string, ratio float64)
	ObserveSubAgentTokens(provider string, model string, tokens int)
	ObserveSubAgentLatency(provider string, model string, duration time.Duration)
	ObserveSubAgentFilterPrecision(provider string, model string, ratio float64)
}
