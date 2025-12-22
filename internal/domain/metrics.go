package domain

import "time"

type Metrics interface {
	ObserveRoute(serverType string, duration time.Duration, err error)
	ObserveInstanceStart(serverType string, duration time.Duration, err error)
	ObserveInstanceStop(serverType string, err error)
	SetActiveInstances(serverType string, count int)
}
