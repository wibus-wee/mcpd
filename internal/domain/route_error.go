package domain

import (
	"errors"
	"fmt"
)

// RouteStage labels which routing phase failed.
type RouteStage string

const (
	// RouteStageDecode indicates decode failure.
	RouteStageDecode RouteStage = "decode"
	// RouteStageValidate indicates validation failure.
	RouteStageValidate RouteStage = "validate"
	// RouteStageAcquire indicates instance acquisition failure.
	RouteStageAcquire RouteStage = "acquire"
	// RouteStageCall indicates execution failure.
	RouteStageCall RouteStage = "call"
)

// RouteError wraps an error with routing stage information.
type RouteError struct {
	Stage RouteStage
	Err   error
}

// Error implements the error interface.
func (e *RouteError) Error() string {
	return fmt.Sprintf("%s: %v", e.Stage, e.Err)
}

// Unwrap returns the underlying error.
func (e *RouteError) Unwrap() error {
	return e.Err
}

// NewRouteError wraps an error with a routing stage when needed.
func NewRouteError(stage RouteStage, err error) error {
	if err == nil {
		return nil
	}
	var routeErr *RouteError
	if errors.As(err, &routeErr) {
		return err
	}
	return &RouteError{Stage: stage, Err: err}
}

// RouteStageFrom extracts a route stage from an error when present.
func RouteStageFrom(err error) (RouteStage, bool) {
	var routeErr *RouteError
	if errors.As(err, &routeErr) {
		return routeErr.Stage, true
	}
	return "", false
}
