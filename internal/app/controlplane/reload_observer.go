package controlplane

import (
	"errors"
	"time"

	"go.uber.org/zap"

	"mcpv/internal/domain"
)

type reloadObserver struct {
	metrics    domain.Metrics
	coreLogger *zap.Logger
	logger     *zap.Logger
}

func newReloadObserver(metrics domain.Metrics, coreLogger, logger *zap.Logger) *reloadObserver {
	if coreLogger == nil {
		coreLogger = zap.NewNop()
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &reloadObserver{
		metrics:    metrics,
		coreLogger: coreLogger,
		logger:     logger,
	}
}

func (o *reloadObserver) recordReloadSuccess(source domain.CatalogUpdateSource, action domain.ReloadAction) {
	if o.metrics == nil {
		return
	}
	o.metrics.RecordReloadSuccess(source, action)
}

func (o *reloadObserver) recordReloadFailure(source domain.CatalogUpdateSource, action domain.ReloadAction) {
	if o.metrics == nil {
		return
	}
	o.metrics.RecordReloadFailure(source, action)
}

func (o *reloadObserver) recordReloadRestart(source domain.CatalogUpdateSource, action domain.ReloadAction) {
	if o.metrics == nil {
		return
	}
	o.metrics.RecordReloadRestart(source, action)
}

func (o *reloadObserver) recordReloadActionFailures(source domain.CatalogUpdateSource, diff domain.CatalogDiff) {
	for range diff.AddedSpecKeys {
		o.recordReloadFailure(source, domain.ReloadActionServerAdd)
	}
	for range diff.RemovedSpecKeys {
		o.recordReloadFailure(source, domain.ReloadActionServerRemove)
	}
	for range diff.UpdatedSpecKeys {
		o.recordReloadFailure(source, domain.ReloadActionServerUpdate)
	}
	for range diff.ReplacedSpecKeys {
		o.recordReloadFailure(source, domain.ReloadActionServerReplace)
	}
}

func (o *reloadObserver) handleApplyError(update domain.CatalogUpdate, err error, duration time.Duration) {
	reloadMode := resolveReloadMode(update.Snapshot.Summary.Runtime.ReloadMode)
	stage := reloadFailureStage(err)
	fields := []zap.Field{
		zap.Uint64("revision", update.Snapshot.Revision),
		zap.Int("servers", update.Snapshot.Summary.TotalServers),
		zap.Int("added", len(update.Diff.AddedSpecKeys)),
		zap.Int("removed", len(update.Diff.RemovedSpecKeys)),
		zap.Int("updated", len(update.Diff.UpdatedSpecKeys)),
		zap.String("reload_mode", string(reloadMode)),
		zap.String("failure_stage", stage),
		zap.String("failure_summary", err.Error()),
		zap.Duration("latency", duration),
		zap.Error(err),
	}
	o.observeReloadApply(reloadMode, domain.ReloadApplyResultFailure, stage, duration)
	if reloadMode == domain.ReloadModeStrict {
		o.coreLogger.Fatal("config reload apply failed; shutting down", fields...)
	}
	o.logger.Warn("config reload apply failed", fields...)
}

func (o *reloadObserver) observeReloadApply(mode domain.ReloadMode, result domain.ReloadApplyResult, summary string, duration time.Duration) {
	if o.metrics == nil {
		return
	}
	o.metrics.ObserveReloadApply(domain.ReloadApplyMetric{
		Mode:     mode,
		Result:   result,
		Summary:  summary,
		Duration: duration,
	})
}

func (o *reloadObserver) observeReloadRollback(mode domain.ReloadMode, result domain.ReloadRollbackResult, summary string, duration time.Duration) {
	if o.metrics == nil {
		return
	}
	o.metrics.ObserveReloadRollback(domain.ReloadRollbackMetric{
		Mode:     mode,
		Result:   result,
		Summary:  summary,
		Duration: duration,
	})
}

func reloadFailureStage(err error) string {
	var applyErr reloadApplyError
	if errors.As(err, &applyErr) && applyErr.stage != "" {
		return applyErr.stage
	}
	return "unknown"
}
