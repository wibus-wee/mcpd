# Server Init Retry Backoff And Cache Refresh Tightening

本 ExecPlan 是一个持续更新的活文档，`Progress`、`Surprises & Discoveries`、`Decision Log`、`Outcomes & Retrospective` 四个章节必须在执行过程中保持最新。

本计划遵循仓库根目录 `.agent/PLANS.md` 的要求，执行与更新必须严格符合该文档。

## Purpose / Big Picture

完成本计划后，MCP Server 启动失败不会进入无休止的短周期重试。系统会对启动失败进行指数退避与最大重试限制，遇到不可恢复错误会进入挂起状态，等待显式的人工重试。与此同时，工具/资源/提示词索引的刷新将避免在单次刷新周期内对全量快照反复重建，降低高负载场景的 CPU 与内存压力。该变化使异常路径可控、可观测，并能在 UI/控制面层进行明确的恢复动作。

## Progress

- [x] (2025-03-09T10:25:00Z) 阅读现有 ServerInitManager、GenericIndex、transport 相关实现，确认可改动范围与依赖。
- [x] (2025-03-09T10:25:00Z) 为 runtime 配置新增 server init 重试策略字段与默认值，完成 schema、loader、UI mapping、示例配置更新。
- [x] (2025-03-09T10:25:00Z) 扩展 ServerInitStatus 状态字段与状态枚举，加入重试计数与挂起状态。
- [x] (2025-03-09T10:25:00Z) 实现 ServerInitManager 的重试上限、指数退避、致命错误判定与手动重试入口。
- [x] (2025-03-09T10:25:00Z) 为 ServerInitManager 增加测试覆盖：重试上限触发挂起、手动重试复位。
- [x] (2025-03-09T10:25:00Z) 优化 GenericIndex 刷新流程为“单次刷新仅重建一次快照”，并在缓存未变化时跳过重建。
- [x] (2025-03-09T10:25:00Z) 更新相关测试，补充刷新行为的新断言。
- [x] (2025-03-09T10:25:00Z) 运行 fmt/vet/test，记录结果。

## Surprises & Discoveries

- Observation: `make fmt` 与 `go vet` 需要显式指定 workspace 内的绝对 `GOCACHE`，否则会触发系统路径权限问题。
  Evidence: `go vet ./...` 报错 `GOCACHE is not an absolute path`。
- Observation: `go test` 在 `internal/ui` 链接阶段出现 macOS 版本警告，但测试仍通过。
  Evidence: `ld: warning: object file ... was built for newer 'macOS' version (26.0) than being linked (11.0)`

## Decision Log

- Decision: Server init 失败进入“挂起”状态时，不再自动重试，仅通过显式重试入口恢复。
  Rationale: 避免错误配置导致的无限重启风暴，符合“熔断”预期并可被用户明确感知。
  Date/Author: 2025-03-09 / Codex
- Decision: GenericIndex 在一次 Refresh 周期内仅重建一次快照，并在缓存未变化时跳过重建。
  Rationale: 显著降低聚合刷新时的重复计算成本，且不改变对外快照语义。
  Date/Author: 2025-03-09 / Codex
- Decision: server init 重试策略进入 runtime 配置，默认值集中于 domain constants。
  Rationale: 避免硬编码并允许不同部署对重试节奏进行调优。
  Date/Author: 2025-03-09 / Codex
- Decision: 引入 `suspended` 状态并输出 `RetryCount` / `NextRetryAt`，作为 UI 与控制面的可见反馈。
  Rationale: 让异常路径可观测且可人工恢复，避免“静默卡死”。
  Date/Author: 2025-03-09 / Codex
- Decision: transport 层不引入 JSON-RPC 类型到 domain，暂不做 RawMessage -> Message 的全量重构。
  Rationale: domain 层保持纯净是当前约束，重构将牵动大量 API 与协议层，需独立评估。
  Date/Author: 2025-03-09 / Codex

## Outcomes & Retrospective

本次改动完成了 server init 重试策略的可配置化、挂起状态与手动重试入口，并在前端补齐了状态展示与重试操作。GenericIndex 的刷新已收敛为单次周期仅重建一次快照，降低了高频刷新时的重复计算。后续如需进一步优化 transport 的内存分配，应在不破坏 domain 纯净性的前提下单独规划迁移。

## Context and Orientation

Server 初始化由 `internal/app/server_init_manager.go` 管理，`runSpec` 会反复调用 `domain.Scheduler.SetDesiredMinReady` 并根据池状态更新 `domain.ServerInitStatus`。这些状态通过 `internal/infra/aggregator/server_init_index.go` 广播给控制面观测层，并由 `internal/ui/mapping.go` 映射给前端。

运行时配置由 `internal/infra/catalog/loader.go` 解析，结构定义在 `internal/domain/types.go` 与 `internal/infra/catalog/schema.json`；默认值集中在 `internal/domain/constants.go`，示例配置位于仓库根目录 `runtime.yaml` 与 `dev/runtime.yaml`。

聚合索引的刷新逻辑在 `internal/infra/aggregator/index_core.go`。`GenericIndex.Refresh` 当前会在单次刷新周期中对每个 server 的结果都调用 `rebuildSnapshot`，导致全量快照重复构建。工具、资源、提示词索引分别实现于 `internal/infra/aggregator/aggregator.go`、`resource_index.go`、`prompt_index.go`。

Domain 层必须保持纯净，不应引入外部库类型；因此对 transport 的内存优化若涉及 JSON-RPC 类型变更，应谨慎评估是否破坏该约束。

## Plan of Work

首先扩展 runtime 配置以承载 server init 重试策略。在 `internal/domain/constants.go` 新增默认值（如 `DefaultServerInitRetryBaseSeconds`、`DefaultServerInitRetryMaxSeconds`、`DefaultServerInitMaxRetries`），在 `internal/domain/types.go` 的 `RuntimeConfig` 中增加对应字段，并在 `internal/infra/catalog/loader.go` 中设置默认值、解析与校验，同时更新 `internal/infra/catalog/schema.json`。为保证 UI 可用性，同步修改 `internal/ui/types.go` 与 `internal/ui/mapping.go`，并更新 `runtime.yaml` 与 `dev/runtime.yaml` 示例值。

其次扩展初始化状态。为 `domain.ServerInitState` 增加挂起状态（例如 `ServerInitSuspended`），并为 `domain.ServerInitStatus` 增加 `RetryCount` 与 `NextRetryAt` 字段，用于呈现当前重试进度与下一次计划时间。更新 UI 结构体与映射逻辑，使前端可以展示挂起状态与重试计数。

随后重构 `internal/app/server_init_manager.go`。引入基于 runtime 配置的指数退避与最大重试次数，避免固定 200ms 的盲目循环。添加致命错误判定：将不可恢复错误（例如命令不存在、权限不足、协议版本不支持）标记为挂起并停止重试。为实现可重试入口，在 `ServerInitializationManager` 增加 `RetrySpec` 或类似方法，并在 `internal/app/control_plane.go` 与 `internal/domain/controlplane.go` 增加对应的控制面入口，供 UI 或后续 RPC 调用触发显式重试。避免在异常路径内做重度计算，锁内仅更新必要状态。

紧接着更新 tests。扩展 `internal/app/server_init_manager_test.go` 覆盖最大重试后挂起、手动重试复位与退避逻辑（必要时缩短重试间隔以控制测试时长）。若新增状态字段影响序列化或 UI 映射，更新相关测试用例。

最后优化 GenericIndex。修改 `internal/infra/aggregator/index_core.go`，使 `Refresh` 在一次刷新周期结束后仅调用一次 `rebuildSnapshot`，并在缓存未变化时跳过重建。为此，在 `GenericIndexOptions` 增加 `CacheETag` 回调，并为 tool/resource/prompt 的 cache 结构体加入 `etag` 字段，在 fetch 阶段计算并保存。更新与之相关的测试，确保更新后快照仅在变化时广播。

如发现 transport 的内存优化需要引入外部 JSON-RPC 类型到 domain 层，应记录为设计约束并暂缓，避免破坏 domain 纯净性。若已存在可行的轻量优化（例如减少重复快照构建），优先完成这些改动。

## Concrete Steps

在仓库根目录执行以下步骤：

1) 编辑 runtime 配置与 schema：
   - `internal/domain/constants.go`
   - `internal/domain/types.go`
   - `internal/infra/catalog/loader.go`
   - `internal/infra/catalog/schema.json`
   - `internal/ui/types.go`
   - `internal/ui/mapping.go`
   - `runtime.yaml`
   - `dev/runtime.yaml`

2) 扩展 server init 状态与重试逻辑：
   - `internal/app/server_init_manager.go`
   - `internal/domain/types.go`
   - `internal/app/control_plane.go`
   - `internal/domain/controlplane.go`
   - `internal/ui/service.go`（如需对外暴露重试入口）

3) 更新测试：
   - `internal/app/server_init_manager_test.go`
   - 任何因新增字段导致失败的 UI / mapping / loader tests

4) 优化 GenericIndex 刷新：
   - `internal/infra/aggregator/index_core.go`
   - `internal/infra/aggregator/aggregator.go`
   - `internal/infra/aggregator/resource_index.go`
   - `internal/infra/aggregator/prompt_index.go`
   - 相关测试文件

5) 运行测试与格式化：
   - 在仓库根目录执行 `GOCACHE=./.cache/go-build make fmt`
   - 在仓库根目录执行 `GOCACHE=./.cache/go-build go vet ./...`
   - 在仓库根目录执行 `GOCACHE=./.cache/go-build make test`

## Validation and Acceptance

通过以下方式验证变更：

1) 启动服务并观察 server init 状态：在配置一个故意错误的 `cmd` 时，状态应在达到最大重试次数后进入 `suspended`（或等效命名）的挂起状态，`RetryCount` 达到上限且不再频繁重试。

2) 触发手动重试入口后，状态应重新进入 `starting`，并再次尝试拉起实例。

3) 工具、资源、提示词索引刷新不应在单次 Refresh 中多次重建快照。可以通过单测或日志观察 refresh 频率降低，并确认更新后快照广播次数减少。

4) `make test` 全部通过；如果新增测试，确保“变更前失败、变更后通过”。

## Idempotence and Recovery

本计划的变更为可重复应用的配置与逻辑调整。若重试策略导致启动无法自动恢复，可通过降低 `serverInitMaxRetries` 或在配置中将其设为更高值进行回滚式调整。若 GenericIndex 优化导致更新未广播，恢复到“每次结果都重建”仅需撤回 `CacheETag` 和单次重建逻辑。

## Artifacts and Notes

暂无。

## Interfaces and Dependencies

需要新增或变更的接口与类型如下：

- `internal/domain/types.go`:
  - `RuntimeConfig` 增加 `ServerInitRetryBaseSeconds`, `ServerInitRetryMaxSeconds`, `ServerInitMaxRetries`。
  - `ServerInitState` 增加挂起状态常量（例如 `ServerInitSuspended`）。
  - `ServerInitStatus` 增加 `RetryCount int` 与 `NextRetryAt time.Time`。
- `internal/domain/controlplane.go`:
  - 增加 `RetryServerInit(ctx context.Context, specKey string) error`（若提供手动重试入口）。
- `internal/app/server_init_manager.go`:
  - 新增 `RetrySpec` 方法并应用指数退避与重试上限策略。
- `internal/infra/aggregator/index_core.go`:
  - `GenericIndexOptions` 新增 `CacheETag func(cache Cache) string`。
  - `Refresh` 逻辑在单次周期内仅重建一次快照，并在 cache 未变化时跳过。

Plan Update Note (2025-03-09T09:20:00Z): Initial plan created to address server init retry/backoff and index refresh optimization.
Plan Update Note (2025-03-09T10:25:00Z): Completed implementation, updated tests, and recorded validation output.
