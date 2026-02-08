# Standardize Concurrency Patterns and Fix Identified Races

This ExecPlan is a living document. The sections Progress, Surprises & Discoveries, Decision Log, and Outcomes & Retrospective must be kept up to date as work proceeds.

This plan is governed by `.agent/PLANS.md` at the repository root. Keep this document aligned with its requirements.

## Purpose / Big Picture

目标是让并发行为更一致、更可推理，并修复已确认的真实竞态条件。在完成后，控制面与聚合器关键路径会有明确的同步边界，已知竞态（例如 Automation 的 `subAgent`）会被消除，同时不改变既有业务行为。可验证方式为：相关包测试与 `go test -race` 通过，且没有引入新的回归。

## Progress

- [x] (2026-02-08 00:00Z) 完成仓库范围并发原语检索与模块级审阅摘要。
- [x] (2026-02-08 00:00Z) 修复 Automation Service 的 `subAgent` 竞态条件。
- [x] (2026-02-08 00:00Z) 文档化 BaseIndex 的锁顺序并调整访问路径以符合锁序。
- [x] (2026-02-08 00:00Z) 统一 Gateway registries 的锁使用策略。
- [x] (2026-02-08 00:00Z) 为 Observability 的 index 指针访问添加同步保护。
- [ ] 运行相关测试与 race 检查并记录结果。

## Surprises & Discoveries

- Observation: BaseIndex 内部使用四把互斥锁，覆盖了相互关联的状态字段，锁顺序没有显式约束，提升了出错与不一致快照的风险。
  Evidence: `internal/infra/aggregator/index/base_index.go` 中的 `specsMu`, `bootstrapMu`, `baseMu`, `serverMu`。
- Observation: Gateway 的 tool/resource registry 同时使用 `applyMu` 与 `mu`，职责重叠，复杂度高。
  Evidence: `internal/infra/gateway/tool_registry.go` 与 `internal/infra/gateway/resource_registry.go`。
- Observation: Observability Service 对 runtime/server init index 的指针赋值未加锁，仅使用 atomic 标记 worker 启动。
  Evidence: `internal/app/controlplane/observability/observability.go`。

## Decision Log

- Decision: 先聚焦在已识别的高风险区域（Automation, BaseIndex, Observability 指针, Gateway registries），其它模块保持现状。
  Rationale: 在不扩大风险的前提下优先修复确定问题并降低复杂度。
  Date/Author: 2026-02-08 / Codex
- Decision: BaseIndex 先采用锁顺序文档化与访问路径收敛，而不是立即合并为单锁。
  Rationale: 在降低认知复杂度的同时避免潜在性能回归，先用最小改动收敛并发约束。
  Date/Author: 2026-02-08 / Codex

## Outcomes & Retrospective

尚未产生代码变更。本计划记录了并发模式盘点与标准化范围，后续将逐步落地并更新此节。

## Context and Orientation

该仓库是 MCP 控制面与运行时的实现，涉及大量并发访问：控制面状态与注册表、聚合器索引、调度与生命周期、Gateway registry、UI 服务与共享缓存。此次工作聚焦于并发模式标准化与已识别竞态修复，避免大规模重构。

关键文件与路径如下：
- `internal/app/controlplane/automation/automation.go`
- `internal/infra/aggregator/index/base_index.go`
- `internal/app/controlplane/observability/observability.go`
- `internal/infra/gateway/tool_registry.go`
- `internal/infra/gateway/resource_registry.go`
- `internal/infra/gateway/prompt_registry.go`

## Module-by-Module Review (Concurrency Inventory)

控制面与应用层：`internal/app/controlplane/state.go` 与 `internal/app/runtime/state.go` 使用单一 `sync.RWMutex` 保护状态，读写边界清晰，保持现状。`internal/app/controlplane/reload.go` 使用 `atomic` 标记启动与版本，风险低。`internal/app/controlplane/observability/observability.go` 对 index 指针未同步，需要增加互斥或原子封装。`internal/app/controlplane/registry/client_registry.go` 使用单一互斥锁保护 map 与 subscriber 集合，保持现状。`internal/app/bootstrap/*` 和 `internal/app/catalog/catalog_provider_dynamic.go` 各自使用单锁或原子值组合，结构清晰，保持现状。

聚合器：`internal/infra/aggregator/index/base_index.go` 使用多把互斥锁保护相互关联字段，缺乏统一锁顺序，属于高风险区域，需要先文档化锁顺序并约束访问路径。`runtime_status_index.go` 与 `server_init_index.go` 使用 `atomic.Value` 保存快照、RWMutex 保护订阅者，属于清晰模式，保持现状。`internal/infra/aggregator/core/index_core.go` 使用 mutex + subsMu + atomic.Value 的组合，职责划分明确，保持现状。

Gateway：`tool_registry.go` 与 `resource_registry.go` 同时使用 `applyMu` 与 `mu`，职责重叠，建议合并为单锁以简化同步；`prompt_registry.go` 仅使用单锁，保持现状。`client_manager.go` 与 `core.go` 使用单锁或 atomic + channel 的组合，保持现状。

调度与生命周期：`internal/infra/scheduler/*` 使用分层锁（全局 + per-pool）与通道协作，虽然复杂但符合调度场景需求，暂不动。`internal/infra/lifecycle/manager.go` 单锁保护连接与 stop 函数，保持现状。

UI 与其他基础设施：`internal/ui/*` 与 `internal/infra/telemetry/*` 的锁模式清晰，保持现状。`internal/domain/session_cache.go` 与 `internal/domain/metadata_cache.go` 使用锁以满足 LRU/TTL 语义，保持现状。

## Plan of Work

第一步修复 Automation Service 的 `subAgent` 竞态，新增专用 `sync.RWMutex` 并提供读取快照方法，所有外部调用在锁外完成。第二步对 BaseIndex 进行锁顺序文档化与访问路径收敛，明确锁职责与锁序，确保所有路径遵循同一顺序并避免跨锁调用外部依赖。第三步简化 Gateway 的 tool/resource registry，将 `applyMu` 与 `mu` 合并为单锁，保证 etag 检查、registered 快照与更新在同一锁域内完成。第四步对 Observability 的 index 指针读写加锁，确保 `Watch` 与 worker 访问使用一致的快照指针。最后执行相关测试与 race 检查。

## Concrete Steps

在仓库根目录执行。

1) 修复 Automation Service 竞态：
   - 编辑 `internal/app/controlplane/automation/automation.go`，新增 `subAgentMu sync.RWMutex` 与 `getSubAgent()`。
   - 更新 `SetSubAgent`、`IsSubAgentEnabled`、`IsSubAgentEnabledForClient`、`AutomaticMCP` 使用读写锁并在锁外调用 SubAgent 方法。

2) 文档化 BaseIndex 锁顺序并收敛访问路径：
   - 编辑 `internal/infra/aggregator/index/base_index.go`。
   - 在结构体与关键方法处写明锁职责与锁序（例如 `specsMu` → `bootstrapMu` → `baseMu` → `serverMu`），并在读写路径统一使用该顺序。
   - 对外部调用保持锁外执行，必要时先取快照再调用。

3) 简化 Gateway registries：
   - 编辑 `internal/infra/gateway/tool_registry.go` 与 `internal/infra/gateway/resource_registry.go`。
   - 合并 `applyMu` 与 `mu` 为单锁，调整 ApplySnapshot 的锁范围。

4) Observability index 指针同步：
   - 编辑 `internal/app/controlplane/observability/observability.go`。
   - 为 `runtimeStatusIdx` 与 `serverInitIdx` 加锁读写或用 `atomic.Value` 封装，读取时先复制指针再使用。

5) 运行测试与 race：

   go test ./internal/app/controlplane/...
   go test ./internal/infra/aggregator/...
   go test ./internal/infra/gateway/...
   go test -race ./internal/app/controlplane/...
   go test -race ./internal/infra/aggregator/...

## Validation and Acceptance

以下结果同时满足即为通过：
- `go test ./internal/app/controlplane/...` 通过。
- `go test ./internal/infra/aggregator/...` 通过。
- `go test ./internal/infra/gateway/...` 通过。
- `go test -race` 针对 controlplane 与 aggregator 包无 data race 报告。
- Automation 的 `subAgent` 在并发 Set/Read 场景下不再触发竞态。

## Idempotence and Recovery

所有修改均可重复应用。如 BaseIndex 锁顺序文档化导致回归，可回退到现有锁实现，但仍保留 Automation 修复（这是明确的正确性修复）。

## Artifacts and Notes

并发检索命令（可复现审查范围）：

   rg -n "sync\\.(Mutex|RWMutex|Once|WaitGroup)|atomic\\.|chan\\s|make\\(chan" internal

## Interfaces and Dependencies

需要在代码中引入的同步元素包括：
- `internal/app/controlplane/automation/automation.go` 增加 `subAgentMu sync.RWMutex` 与 `getSubAgent()`。
- `internal/infra/aggregator/index/base_index.go` 添加锁顺序与职责说明，并将访问路径收敛到文档化的锁序。
- `internal/app/controlplane/observability/observability.go` 增加 index 指针同步保护。
- `internal/infra/gateway/tool_registry.go` 与 `internal/infra/gateway/resource_registry.go` 简化为单锁模型。

Plan Change Note: Updated BaseIndex approach from single-lock consolidation to lock-order documentation and access-path convergence, based on risk assessment and to minimize performance regression risk.
