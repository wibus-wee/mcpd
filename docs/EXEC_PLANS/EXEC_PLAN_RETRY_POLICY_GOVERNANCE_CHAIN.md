# Retry Policy Unification + Governance Policy Chain

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

This document follows `.agent/PLANS.md` from the repository root and must be maintained according to it.

## Purpose / Big Picture

统一基础设施层的重试/退避策略，并将治理逻辑抽象为可组合的策略链。完成后，Gateway/Plugin/Lifecycle 会共享同一套重试语义，治理决策会通过统一的 Policy Chain 执行，便于新增策略、统一行为与测试。系统对外行为保持不变，所有相关测试继续通过。

## Progress

- [x] (2026-02-07 17:40) 引入 `internal/infra/retry` 包（Policy/Backoff/Retry/Loop），并迁移 Gateway/Plugin/Lifecycle 的重试逻辑。
- [x] (2026-02-07 17:45) 在 `internal/infra/governance` 增加 Policy Chain 抽象，并改造 Executor 使用 Chain。
- [x] (2026-02-07 17:42) 删除 gateway 旧 backoff 实现，统一使用 retry。
- [x] (2026-02-07 17:48) 运行相关测试与全量 infra 测试（`go test ./internal/infra/...`）。

## Surprises & Discoveries

- Observation: None yet.
  Evidence: N/A.

## Decision Log

- Decision: 新增 `internal/infra/retry` 作为统一重试/退避实现，并替换 gateway/backoff.go。
  Rationale: 统一策略、可配置、可复用，减少重复实现和语义偏差。
  Date/Author: 2026-02-07 / Codex
- Decision: Governance 以 Policy Chain 为核心抽象，Executor 成为 Chain 的适配器。
  Rationale: 统一策略入口，便于未来组合多策略并保持 RPC 层稳定。
  Date/Author: 2026-02-07 / Codex
- Decision: 允许删除旧 backoff 实现（破坏性重构），以避免重复维护。
  Rationale: 单一实现来源，避免逻辑漂移。
  Date/Author: 2026-02-07 / Codex

## Outcomes & Retrospective

完成 retry 统一与治理策略链抽象，Gateway/Plugin/Lifecycle 重试逻辑一致化，Executor 改为 Chain 驱动且对外 API 不变。`go test ./internal/infra/...` 全部通过。后续新增治理策略可直接以 Policy 形式接入，无需改 RPC 层。

## Context and Orientation

当前重试与退避逻辑分散在 gateway、plugin、lifecycle 等模块中，策略语义不一致；治理逻辑通过 `governance.Executor` 与 RPC guard 组合，但缺少可组合的策略链抽象。此计划引入统一 retry 包和 Policy Chain，使行为集中、清晰、可扩展。

关键路径：

- Gateway: `internal/infra/gateway/*`
- Plugin: `internal/infra/plugin/manager.go`
- Lifecycle: `internal/infra/lifecycle/manager.go`
- Governance: `internal/infra/governance/executor.go`
- RPC: `internal/infra/rpc/control_service_governance.go`

## Plan of Work

### Phase 1: Retry Policy Unification

1) 新增 `internal/infra/retry` 包：

- `policy.go`: 定义 `Policy`（BaseDelay/MaxDelay/Factor/Jitter/MaxRetries）。
- `backoff.go`: `Backoff`（Next/Reset/Sleep），支持抖动。
- `retry.go`: `Retry(ctx, policy, fn)` 与 `Loop(ctx, policy, fn)`。

2) 迁移 Gateway 重试：

- `internal/infra/gateway/syncer.go`、`internal/infra/gateway/log_bridge.go` 使用 `retry.Backoff`。
- 删除 `internal/infra/gateway/backoff.go`。

3) 迁移 Plugin 重试：

- `internal/infra/plugin/manager.go` 中 socket 等待重试改用 `retry.Retry`，保持 100ms 固定间隔和 timeout 语义。

4) 迁移 Lifecycle 重试：

- `internal/infra/lifecycle/manager.go` 中 `initializeWithRetry` 使用 `retry.Retry`，保留重试次数与日志语义。

### Phase 2: Governance Policy Chain

1) 在 `internal/infra/governance` 新增 `policy.go`：

- `Policy` 接口：`Request(ctx, req)` / `Response(ctx, req)`。
- `Chain`：按顺序执行 Request，逆序执行 Response。
- `PipelinePolicy`：适配现有 `pipeline.Engine`。

2) 改造 `Executor`：

- `Executor` 内部使用 `Chain`，对外 API 保持一致（Request/Response/Execute）。

3) 保持 RPC 层不变，仅替换 Executor 内部实现。

## Concrete Steps

Phase 1:

- 创建 `internal/infra/retry/*`。
- 替换 gateway/syncer/log_bridge 的 backoff。
- 移除 `internal/infra/gateway/backoff.go`。
- 更新 plugin/lifecycle 重试逻辑。
- 运行：

  gofmt -w internal/infra/retry/*.go
  gofmt -w internal/infra/gateway/*.go
  gofmt -w internal/infra/plugin/*.go
  gofmt -w internal/infra/lifecycle/*.go

  go test ./internal/infra/gateway
  go test ./internal/infra/plugin
  go test ./internal/infra/lifecycle

Phase 2:

- 新增 `internal/infra/governance/policy.go`。
- 改造 `internal/infra/governance/executor.go`。
- 运行：

  gofmt -w internal/infra/governance/*.go
  go test ./internal/infra/governance

Final:

- 运行：

  go test ./internal/infra/...

## Validation and Acceptance

- 所有 targeted tests 通过，`go test ./internal/infra/...` 通过。
- Gateway 仍能稳定 sync tools/resources/prompts（通过现有测试验证）。
- Plugin/Lifecycle 重试次数与超时语义保持一致。
- Governance 决策链行为与旧 Executor 一致（无错误码变化）。

## Idempotence and Recovery

- 步骤可重复执行。
- 若迁移失败，可恢复 gateway/backoff.go 并回退引用；其余变更可通过 git revert 回滚。

## Artifacts and Notes

新增/修改关键文件：

- `internal/infra/retry/*`
- `internal/infra/gateway/syncer.go`
- `internal/infra/gateway/log_bridge.go`
- `internal/infra/plugin/manager.go`
- `internal/infra/lifecycle/manager.go`
- `internal/infra/governance/policy.go`
- `internal/infra/governance/executor.go`

## Interfaces and Dependencies

- Retry:
  - `Policy` with `MaxRetries < 0` meaning infinite retries (bounded by context).
  - `Backoff.Reset/Next/Sleep` to replace old gateway backoff.
  - `Retry` used by plugin/lifecycle.

- Governance:
  - `Policy` interface with request/response stage hooks.
  - `Chain.Execute` executes policies then delegates to next.
  - `Executor` maintains existing public API.

---

Change Log: Initial creation of the ExecPlan for retry policy unification and governance policy chain refactor.

Change Log: Updated Progress and Outcomes after completing implementation and validation.
