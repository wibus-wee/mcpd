package aggregator

import (
	"mcpv/internal/domain"
	core "mcpv/internal/infra/aggregator/core"
	idx "mcpv/internal/infra/aggregator/index"
	"mcpv/internal/infra/telemetry"

	"go.uber.org/zap"
)

type RefreshGate = core.RefreshGate

type ToolIndex = idx.ToolIndex

type ResourceIndex = idx.ResourceIndex

type PromptIndex = idx.PromptIndex

type RuntimeStatusIndex = idx.RuntimeStatusIndex

type ServerInitIndex = idx.ServerInitIndex

type GenericIndex[Snapshot any, Target any, Cache any] = core.GenericIndex[Snapshot, Target, Cache]

type GenericIndexOptions[Snapshot any, Target any, Cache any] = core.GenericIndexOptions[Snapshot, Target, Cache]

type RefreshErrorDecision = core.RefreshErrorDecision

type listChangeSubscriber = core.ListChangeSubscriber

func NewRefreshGate() *RefreshGate {
	return core.NewRefreshGate()
}

func NewToolIndex(rt domain.Router, specs map[string]domain.ServerSpec, specKeys map[string]string, cfg domain.RuntimeConfig, metadataCache *domain.MetadataCache, logger *zap.Logger, health *telemetry.HealthTracker, gate *RefreshGate, listChanges listChangeSubscriber) *ToolIndex {
	return idx.NewToolIndex(rt, specs, specKeys, cfg, metadataCache, logger, health, gate, listChanges)
}

func NewResourceIndex(rt domain.Router, specs map[string]domain.ServerSpec, specKeys map[string]string, cfg domain.RuntimeConfig, metadataCache *domain.MetadataCache, logger *zap.Logger, health *telemetry.HealthTracker, gate *RefreshGate, listChanges listChangeSubscriber) *ResourceIndex {
	return idx.NewResourceIndex(rt, specs, specKeys, cfg, metadataCache, logger, health, gate, listChanges)
}

func NewPromptIndex(rt domain.Router, specs map[string]domain.ServerSpec, specKeys map[string]string, cfg domain.RuntimeConfig, metadataCache *domain.MetadataCache, logger *zap.Logger, health *telemetry.HealthTracker, gate *RefreshGate, listChanges listChangeSubscriber) *PromptIndex {
	return idx.NewPromptIndex(rt, specs, specKeys, cfg, metadataCache, logger, health, gate, listChanges)
}

func NewRuntimeStatusIndex(scheduler domain.Scheduler, logger *zap.Logger) *RuntimeStatusIndex {
	return idx.NewRuntimeStatusIndex(scheduler, logger)
}

func NewServerInitIndex(cp domain.ServerInitStatusReader, logger *zap.Logger) *ServerInitIndex {
	return idx.NewServerInitIndex(cp, logger)
}
