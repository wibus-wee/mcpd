# Infra Modular Refactor (Gateway, Catalog Loader, Aggregator)

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

This document follows `.agent/PLANS.md` from the repository root and must be maintained according to it.

## Purpose / Big Picture

This plan restructures three large infrastructure areas into coherent modules with clear ownership, while preserving behavior. The outcome is a control plane stack that is easier to reason about, test, and extend. A user should still be able to start the system, use MCP tools/resources/prompts, and observe identical runtime behavior; tests for gateway, catalog loading, and aggregation should continue to pass. The work is split into phases so each change is small enough to validate but the final architecture is unified.

## Progress

- [x] (2026-02-07 15:30) Catalog loader refactor: split raw types, decode, normalize, validate, defaults; keep Load API stable. Tests: `go test ./internal/infra/catalog`.
- [x] (2026-02-07 15:55) Gateway refactor: introduced core/sync/handlers/rpc_client/list_all modules and shared snapshot syncer. Tests: `go test ./internal/infra/gateway`.
- [x] (2026-02-07 16:25) Aggregator refactor: split into core/index packages with facade in aggregator. Tests: `go test ./internal/infra/aggregator/...`.
- [x] (2026-02-07 16:30) Run gofmt and targeted tests after each phase; run broader tests at the end. Tests: `go test ./internal/infra/...`.

## Surprises & Discoveries

- Observation: Refresh timeout semantics had to stay on `RuntimeConfig.RouteTimeout()` to preserve refresh blocking behavior.
  Evidence: `TestToolIndex_RefreshConcurrentFetches` failed when refresh timeout defaulted to tool refresh interval; fixing restored test.

## Decision Log

- Decision: Use a module-first refactor with clear dependency direction and facade wrappers to preserve behavior during migration.
  Rationale: This provides the best long-term architecture while keeping incremental validation possible.
  Date/Author: 2026-02-07 / Codex
- Decision: Execute phases in order: catalog loader → gateway → aggregator.
  Rationale: This reduces risk by starting with a pure data-processing module, then the gateway, then the highest-coupling aggregator.
  Date/Author: 2026-02-07 / Codex

## Outcomes & Retrospective

Completed the three-phase refactor with behavior preserved. Catalog loader, gateway, and aggregator are now modularized with a facade for compatibility. Tests across `internal/infra` pass. No user-facing behavior change observed; future work can safely iterate within each module boundary.

## Context and Orientation

The repository currently has several oversized files and cross-cutting responsibilities:

- `internal/infra/gateway/gateway.go` mixes RPC lifecycle, list/watch loops, tool/resource/prompt handlers, and readiness gating.
- `internal/infra/catalog/loader.go` contains raw type definitions, decoding, normalization, validation, and runtime defaults.
- `internal/infra/aggregator/aggregator.go` implements tool index, caching, refresh, request building, and list-change coordination.

The goal is to split these into smaller modules with single responsibilities, while preserving the public API and behavior. Each refactor must be testable and should not alter semantics.

## Plan of Work

### Phase 1: Catalog Loader Modularization

Scope: `internal/infra/catalog/loader.go`.

Actions:

1) Create file-scoped modules within `internal/infra/catalog/`:

- `raw_types.go` for raw config structs (rawCatalog, rawServerSpec, rawPluginSpec, rawStreamableHTTPConfig, rawRuntimeConfig, rawSubAgentConfig, rawObservabilityConfig, rawRPCConfig, rawRPCTLSConfig).
- `defaults.go` for `newRuntimeViper` and `setRuntimeDefaults`.
- `decoder.go` for config decoding (viper read, decodeRuntimeConfig), keeping error messages unchanged.
- `normalizer_server.go` for normalizeServerSpec, normalizeStreamableHTTPConfig, normalizeHTTPHeaders, normalizeTags.
- `normalizer_plugin.go` for normalizePluginSpecs, normalizePluginSpec.
- `normalizer_runtime.go` for normalizeRuntimeConfig, normalizeObservabilityConfig, normalizeRPCConfig.
- `validator_server.go` for validateServerSpec and validateStreamableHTTPSpec.
- `loader.go` should keep only Loader, NewLoader, LoadRuntimeConfig, Load, and their orchestration.

2) Ensure no behavior change: error messages and validation rules must remain identical.

3) Update any tests if paths/visibility change. Ensure `internal/infra/catalog/loader_test.go` passes.

Acceptance: `go test ./internal/infra/catalog` passes, and Load/LoadRuntimeConfig return the same behavior as before.

### Phase 2: Gateway Modularization

Scope: `internal/infra/gateway/gateway.go` and gateway package.

Actions:

1) Introduce submodules (still inside `internal/infra/gateway`):

- `core.go`: Gateway struct, NewGateway, Run, run, toolsReady gating, heartbeat, register/unregister.
- `syncer.go`: snapshotSyncer abstraction that handles list/watch loop with backoff and re-registration, reusable for tools/resources/prompts.
- `handlers.go`: toolHandler, promptHandler, resourceHandler, plus automatic_mcp/automatic_eval handlers.
- `rpc_client.go`: helper wrappers for RPC calls (callTool/getPrompt/readResource/automaticMCP/automaticEval) that manage FailedPrecondition + Unavailable resets.
- `list_all.go`: listAllResources / listAllPrompts pagination logic.

2) Replace the three near-duplicate sync loops with the shared syncer, preserving ETag and error behavior. Use closures for:

- list function (ListTools/ListResources/ListPrompts)
- watch function (WatchTools/WatchResources/WatchPrompts)
- apply snapshot callback (registry.ApplySnapshot)

3) Keep logging messages and error behaviors consistent.

Acceptance: `go test ./internal/infra/gateway` passes. Gateway behavior unchanged.

### Phase 3: Aggregator Package Split

Scope: `internal/infra/aggregator/*`.

Actions:

1) Create `internal/infra/aggregator/core/` containing:

- GenericIndex, RefreshGate, refresh workers, request builder, list-change hub/subscriber.
- Supporting helpers (context_helpers, snapshot_keys) that are shared.

2) Create `internal/infra/aggregator/index/` containing:

- tool_index, resource_index, prompt_index, runtime_status_index, server_init_index.

3) Keep `internal/infra/aggregator` as facade:

- Provide type aliases for the core/index types, plus constructors/wrappers to minimize call site changes.
- Move tests to appropriate packages or keep them in facade if they verify public API.

4) Update imports across the repo to reflect the new package structure, keeping stable top-level types if possible.

Acceptance: `go test ./internal/infra/aggregator/...` passes and any callers compile without behavior changes.

## Concrete Steps

All commands are run from repository root.

Phase 1:

- Split files as described.
- Run:

  gofmt -w internal/infra/catalog/*.go
  go test ./internal/infra/catalog

Phase 2:

- Create new gateway modules and refactor gateway.go.
- Run:

  gofmt -w internal/infra/gateway/*.go
  go test ./internal/infra/gateway

Phase 3:

- Create new aggregator/core and aggregator/index packages and move files.
- Update imports.
- Run:

  gofmt -w internal/infra/aggregator/**/*.go
  go test ./internal/infra/aggregator/...

Final:

- Run a broader test pass:

  go test ./internal/infra/...

## Validation and Acceptance

- Each phase must pass its package tests.
- Behavior should be unchanged; any changes in errors/logs must be consciously documented.
- For gateway, ensure tools/resources/prompts sync paths remain stable by passing existing tests.
- For aggregator, ensure indexes still refresh and resolve, and tests pass.

## Idempotence and Recovery

- The refactor is file/module oriented; rerunning the steps is safe.
- If a phase fails midway, `git status` shows moved files; move them back or reset the phase and re-run gofmt/tests.

## Artifacts and Notes

Key files after refactor:

- `internal/infra/catalog/loader.go`, `raw_types.go`, `decoder.go`, `normalizer_*.go`, `validator_*.go`, `defaults.go`.
- `internal/infra/gateway/core.go`, `syncer.go`, `handlers.go`, `rpc_client.go`, `list_all.go`.
- `internal/infra/aggregator/core/*`, `internal/infra/aggregator/index/*`, with `internal/infra/aggregator` as facade.

## Interfaces and Dependencies

- Catalog loader must preserve `Loader.Load` and `Loader.LoadRuntimeConfig` signatures.
- Gateway must preserve `NewGateway(cfg, caller, tags, serverName, logger)` and `Run(ctx)` behavior.
- Aggregator facade must preserve exported constructors used by the rest of the codebase (type aliases and wrappers where needed).

---

Change Log: Created initial ExecPlan for infra modular refactor to provide a detailed, unified plan prior to implementation.
