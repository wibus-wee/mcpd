# ControlPlane Registry 拆分与职责收敛

这是一个可执行计划（ExecPlan），它是一个活文档。必须在执行过程中持续更新 `Progress`、`Surprises & Discoveries`、`Decision Log` 和 `Outcomes & Retrospective`。

仓库内存在 `.agent/PLANS.md`，本计划必须符合其要求并持续维护。

## Purpose / Big Picture

完成本计划后，`ClientRegistry` 不再承担可见性解析与 caller 探测的全部职责。可见性解析被收敛到 `VisibilityResolver`，caller 探测被封装为 `CallerProbe`，注册与监控逻辑留在 `ClientRegistry`。这样能让 registry 的核心职责更加清晰，便于测试和演进，同时对外 API 行为保持不变。验证方式是运行 `go test ./internal/app/controlplane` 并观察全部通过。

## Progress

- [x] (2026-02-07 23:45Z) 阅读 `internal/app/controlplane/registry.go` 与调用点，确认职责拆分边界与依赖关系。
- [x] (2026-02-07 23:55Z) 新增 `VisibilityResolver` 与 `CallerProbe`，并迁移可见性与探测逻辑。
- [x] (2026-02-07 23:58Z) 拆分 `ClientRegistry` 实现文件并删除原 `registry.go`。
- [x] (2026-02-08 00:00Z) 运行 `go test ./internal/app/controlplane` 验证行为稳定。

## Surprises & Discoveries

暂无。

## Decision Log

- Decision: 以 `ClientRegistry` 为核心服务，新增 `VisibilityResolver` 与 `CallerProbe` 作为协作组件，并保持现有公开方法签名不变。
  Rationale: 在不破坏外部调用的前提下完成职责收敛，降低风险。
  Date/Author: 2026-02-07 / Codex

## Outcomes & Retrospective

- `ClientRegistry` 已拆分为多文件结构，注册/监控/激活逻辑与可见性解析分离。
- 新增 `VisibilityResolver` 与 `CallerProbe`，`ClientRegistry` 通过协作组件完成职责收敛。
- `go test ./internal/app/controlplane` 已通过，行为保持稳定。

## Context and Orientation

`ClientRegistry` 位于 `internal/app/controlplane/registry.go`，负责客户端注册、监控、可见性解析、spec 激活/停用与 catalog 变更处理。caller 探测实现位于 `internal/app/controlplane/caller_probe_*.go` 中的 `pidAlive`。可见性判断工具函数 `isVisibleToTags` 位于 `internal/app/controlplane/visibility.go`。

## Plan of Work

先引入 `VisibilityResolver` 以集中处理 tags/server 到 spec keys 的解析与规范化，再引入 `CallerProbe` 以封装 pid 存活检测。随后将 `ClientRegistry` 拆成多个文件以降低单文件复杂度，并将可见性与探测逻辑改为调用新组件。最后删除旧的 `registry.go`，运行 gofmt 与测试。

## Concrete Steps

1) 新增 `VisibilityResolver` 与 `CallerProbe`。

   - 文件：`internal/app/controlplane/visibility_resolver.go`
     - 定义 `VisibilityResolver` 与 `NewVisibilityResolver`，实现 `VisibleSpecKeys`、`VisibleSpecKeysForCatalog`、`NormalizeTags`、`NormalizeServerName`、`TagsEqual`。
   - 文件：`internal/app/controlplane/caller_probe.go`
     - 定义 `CallerProbe` 接口与默认实现 `OSCallerProbe`，封装 `pidAlive`。

2) 拆分 registry 实现并删除旧文件。

   - 新文件建议：

       internal/app/controlplane/client_registry.go
       internal/app/controlplane/client_registry_resolver.go
       internal/app/controlplane/client_registry_activation.go
       internal/app/controlplane/client_registry_monitor.go
       internal/app/controlplane/client_registry_catalog.go
       internal/app/controlplane/client_registry_snapshot.go
       internal/app/controlplane/client_registry_helpers.go

   - `ClientRegistry` 持有 `resolver *VisibilityResolver` 与 `probe CallerProbe`，构造时注入默认实现。
   - 将原 `registry.go` 中的方法按职责移动至对应文件。
   - 保持对外方法签名不变（例如 `resolveClientTags`、`resolveVisibleSpecKeys` 等）。

3) 删除 `internal/app/controlplane/registry.go` 并执行 gofmt。

4) 运行 `go test ./internal/app/controlplane`。

## Validation and Acceptance

    cd /Users/wibus/dev/mcpd
    go test ./internal/app/controlplane

期望结果：测试全部通过，注册/监控/可见性逻辑行为与重构前一致。

## Idempotence and Recovery

本次重构仅为代码拆分与内部协作组件引入，可重复执行 gofmt 与测试。如果出现编译错误，优先检查 `ClientRegistry` 的新字段与构造函数是否一致，或是 `VisibilityResolver` 的方法调用未更新。

## Artifacts and Notes

关键新增文件：

    internal/app/controlplane/visibility_resolver.go
    internal/app/controlplane/caller_probe.go
    internal/app/controlplane/client_registry.go
    internal/app/controlplane/client_registry_resolver.go
    internal/app/controlplane/client_registry_activation.go
    internal/app/controlplane/client_registry_monitor.go
    internal/app/controlplane/client_registry_catalog.go
    internal/app/controlplane/client_registry_snapshot.go
    internal/app/controlplane/client_registry_helpers.go

## Interfaces and Dependencies

- `ClientRegistry` 仍为核心服务类型，对外 API 保持不变。
- `VisibilityResolver` 提供：

    VisibleSpecKeys(tags []string, server string) ([]string, int)
    VisibleSpecKeysForCatalog(catalog domain.Catalog, serverSpecKeys map[string]string, tags []string, server string) ([]string, int)
    NormalizeTags(tags []string) []string
    NormalizeServerName(server string) string
    TagsEqual(a []string, b []string) bool

- `CallerProbe` 提供：

    Alive(pid int) bool

Plan change note: 2026-02-07 / Codex - 创建本 ExecPlan，准备执行 registry 拆分。
Plan change note: 2026-02-08 / Codex - 完成拆分与测试，并回填 Progress/Outcomes。
