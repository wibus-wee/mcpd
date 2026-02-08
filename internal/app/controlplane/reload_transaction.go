package controlplane

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"mcpv/internal/domain"
)

type reloadTransaction struct {
	observer *reloadObserver
	logger   *zap.Logger
}

func newReloadTransaction(observer *reloadObserver, logger *zap.Logger) *reloadTransaction {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &reloadTransaction{
		observer: observer,
		logger:   logger,
	}
}

func (t *reloadTransaction) apply(ctx context.Context, steps []reloadStep, mode domain.ReloadMode) error {
	applied := make([]reloadStep, 0, len(steps))
	for _, step := range steps {
		if err := step.apply(ctx); err != nil {
			applyErr := wrapReloadStage(step.name, err)
			rollbackStart := time.Now()
			if rollbackErr := t.rollbackSteps(ctx, applied); rollbackErr != nil {
				rollbackDuration := time.Since(rollbackStart)
				if t.observer != nil {
					t.observer.observeReloadRollback(mode, domain.ReloadRollbackResultFailure, step.name, rollbackDuration)
				}
				t.logger.Warn("config reload rollback failed",
					zap.String("failure_stage", step.name),
					zap.Duration("latency", rollbackDuration),
					zap.Error(rollbackErr),
				)
				return errors.Join(applyErr, wrapReloadStage("rollback", rollbackErr))
			}
			rollbackDuration := time.Since(rollbackStart)
			if t.observer != nil {
				t.observer.observeReloadRollback(mode, domain.ReloadRollbackResultSuccess, step.name, rollbackDuration)
			}
			t.logger.Info("config reload rolled back",
				zap.String("failure_stage", step.name),
				zap.Duration("latency", rollbackDuration),
			)
			return applyErr
		}
		applied = append(applied, step)
	}
	return nil
}

func (t *reloadTransaction) rollbackSteps(ctx context.Context, steps []reloadStep) error {
	if len(steps) == 0 {
		return nil
	}
	var rollbackErr error
	for i := len(steps) - 1; i >= 0; i-- {
		step := steps[i]
		if step.rollback == nil {
			continue
		}
		if err := step.rollback(ctx); err != nil {
			rollbackErr = errors.Join(rollbackErr, err)
		}
	}
	return rollbackErr
}

func wrapReloadStage(stage string, err error) error {
	if err == nil {
		return nil
	}
	var applyErr reloadApplyError
	if errors.As(err, &applyErr) {
		return err
	}
	return reloadApplyError{stage: stage, err: err}
}
