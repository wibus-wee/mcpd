# Gateway Selector Routing for mcpvmcp

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

This plan must be maintained in accordance with `/.agents/PLANS.md` from the repository root.

## Purpose / Big Picture

完成后，一个单独的 `mcpvmcp` 实例可以根据 HTTP URL 路径（主方式）或 Header（备用方式）动态选择不同的 MCP server selector，并为每个 selector 启动独立的 Gateway runtime。使用者可以用不同的 endpoint 连接不同 MCP server（例如 `http://127.0.0.1:8090/mcp/server/context7`），或在同一 endpoint 上通过 Header 切换 selector。可观察的行为是：请求缺少 selector 会得到明确的 400 错误；带有 selector 的请求会连到对应 server 的 tool 列表，且不同 selector 之间的 tools/resources/prompts 完全隔离。

## Progress

- [x] (2026-02-10 00:00Z) Drafted initial ExecPlan and recorded key decisions.
- [x] (2026-02-10 00:30Z) Implemented selector parsing (URL + Header) with conflict handling and normalization.
- [x] (2026-02-10 00:30Z) Implemented GatewayPool with per-selector runtime lifecycle and idle eviction.
- [x] (2026-02-10 00:30Z) Refactored gateway HTTP server to route requests to pooled runtimes.
- [x] (2026-02-10 00:30Z) Updated CLI flags, gateway settings, and docs to match breaking routing model.
- [x] (2026-02-10 00:30Z) Added tests for selector parsing and pool eviction.
- [x] (2026-02-10 01:20Z) Ran `make lint-fix` and `make test` successfully after adding register-ready gating and runtime start test.

## Surprises & Discoveries

- Observation: The new StartRuntime gating required resetting runtime state when startup fails to satisfy tests.
  Evidence: `TestGatewayStartRuntime_FailsWhenRegisterCallerFails` passed after clearing `runtimeStarted` and register-ready state on failure.

## Decision Log

- Decision: Drop backward compatibility and require every HTTP request to include an explicit selector, with URL as the primary mechanism and headers as a secondary mechanism. Missing selector returns HTTP 400.
  Rationale: Keeps routing deterministic and aligns with MCP client configuration reality where endpoints are distinct per server.
  Date/Author: 2026-02-10 / Codex.

- Decision: Use per-selector Gateway runtimes with per-selector callers (e.g., `mcpvmcp:server:context7`) instead of multiplexing a single caller.
  Rationale: Core visibility and registry sync depend on the caller selector; sharing a caller would cause conflicting registrations and tool lists.
  Date/Author: 2026-02-10 / Codex.

- Decision: Do not introduce explicit allowlists for selector values.
  Rationale: User requested trusted, open routing without allowlist overhead.
  Date/Author: 2026-02-10 / Codex.

- Decision: URL selector formats are `/mcp/server/{name}` and `/mcp/tags/{tag1,tag2}`; headers are `X-Mcp-Server` and `X-Mcp-Tags`. URL has priority; conflicting URL/header values return HTTP 400.
  Rationale: URL endpoints map cleanly to MCP client configuration; headers remain available for future authorization and special cases.
  Date/Author: 2026-02-10 / Codex.

- Decision: Keep stdio transport support via new `--selector-server` and `--selector-tag` flags, and default transport to streamable HTTP for gateway usage.
  Rationale: Preserves stdio for advanced use while aligning default behavior with path-based HTTP routing.
  Date/Author: 2026-02-10 / Codex.

## Outcomes & Retrospective

Not started.

## Context and Orientation

`mcpvmcp` is the MCP Gateway binary defined in `cmd/mcpvmcp/main.go`. It currently constructs a single `Gateway` using `internal/infra/gateway.NewGateway`, then calls `RunStreamableHTTP` to expose a single Streamable HTTP endpoint. The current HTTP handler in `internal/infra/gateway/http_server.go` calls `mcp.NewStreamableHTTPHandler` with a `getServer` function that ignores the request and always returns one `*mcp.Server`, so it cannot route per request.

A `Gateway` in `internal/infra/gateway/core.go` creates one `mcp.Server` plus tool/resource/prompt registries, registers a caller with the core via RPC, and starts synchronization loops. The caller registration in `internal/app/controlplane/registry/client_registry.go` uses either server or tags (mutually exclusive) to compute visible server specs; therefore a caller cannot safely switch selectors.

The Wails UI configures the gateway using Go settings in `internal/ui/gateway_settings.go` and React UI files under `frontend/src/modules/settings`. These currently expose server/tags fields and build `mcpvmcp` CLI args for a fixed selector, which will become invalid after this change.

Definitions used in this plan:

A selector is the routing choice for the gateway, defined as either `server=<name>` or `tags=[tag1,tag2]`. A Gateway runtime is one in-memory instance of `Gateway` plus its `mcp.Server`, registries, RPC client, and sync loops for exactly one selector. A GatewayPool is a component that creates and reuses Gateway runtimes keyed by selector, and can evict idle runtimes.

## Plan of Work

Refactor the gateway so that per-selector runtime creation is separated from the HTTP server. Introduce a selector parser that extracts selector information from URL and headers, validates conflicts, and normalizes tags. Add a GatewayPool that lazily creates a new `Gateway` per selector, starts its runtime with a long-lived context, and stores `lastUsed` for eviction. The Streamable HTTP handler will first resolve the selector, then fetch the correct `mcp.Server` from the pool, and then delegate to `mcp.NewStreamableHTTPHandler`. The handler should return clear HTTP 400 errors for missing or invalid selectors and HTTP 503 errors when a runtime cannot be created.

Update `cmd/mcpvmcp/main.go` to remove fixed selector flags (`--server`, `--tag`, `--allow-all`), add stdio-only selector flags (`--selector-server`, `--selector-tag`), and to adopt a fixed base path (`--http-path`, default `/mcp`) for routing. The caller name should default to a stable base value (for example `mcpvmcp`) and be expanded per selector by the pool. Update `internal/ui/gateway_settings.go` and frontend settings UI to remove server/tags fields and to describe the new URL format so the UI doesn’t generate invalid CLI args. Update documentation describing gateway usage (for example `docs/core/gateway.mdx` if present) to show the new URL and header routing rules.

Add tests in `internal/infra/gateway` for selector parsing (URL only, header only, both matching, both conflicting, missing selector) and for GatewayPool behavior (new selector creates runtime, reuse returns same runtime, idle eviction removes runtime). Tests should avoid external RPC dependencies by using fakes or by validating only the selector and pool bookkeeping.

## Concrete Steps

All commands run from `/Users/wibus/dev/mcpd` unless noted.

1) Inspect existing gateway HTTP handler and CLI wiring.

    rg -n "RunStreamableHTTP|buildStreamableHTTPHandler|NewGateway" internal/infra/gateway cmd/mcpvmcp

2) Add selector parsing and pool types in `internal/infra/gateway`.

    - Create `internal/infra/gateway/selector.go` with selector parsing helpers and normalization.
    - Create `internal/infra/gateway/pool.go` with `GatewayPool` implementation and eviction logic.

3) Refactor gateway runtime lifecycle in `internal/infra/gateway/core.go`.

    - Add `StartRuntime(ctx)` and `StopRuntime(ctx)` so a Gateway can run without owning the HTTP server.
    - Ensure `StartRuntime` returns quickly after starting sync loops.

4) Rewrite `internal/infra/gateway/http_server.go` to use the selector router and pool.

    - Wrap `mcp.NewStreamableHTTPHandler` with a handler that resolves selector and injects the chosen `*mcp.Server` into the request context.
    - Update CORS allowed headers to include `X-Mcp-Server` and `X-Mcp-Tags`.

5) Update CLI flags and config.

    - Remove `--server`, `--tag`, `--allow-all` usage from `cmd/mcpvmcp/main.go` and `internal/ui/gateway_settings.go`.
    - Ensure `--caller` defaults to `mcpvmcp` if empty.
    - Keep `--http-path` as the routing base path only (not a complete endpoint).

6) Update frontend gateway settings UI and docs.

    - Remove server/tags fields from `frontend/src/modules/settings/lib/gateway-config.ts` and `frontend/src/modules/settings/components/gateway-settings-card.tsx`.
    - Update copy to explain new URL formats, and ensure file headers and directory README.md are updated per `frontend/CLAUDE.md`.
    - Update gateway documentation to describe the new routing rules.

7) Add tests.

    - New tests for selector parsing and pool behavior in `internal/infra/gateway`.
    - Use `go test ./internal/infra/gateway` to validate.

8) Run lint fixes and full test suite.

    make lint-fix
    make test

## Validation and Acceptance

Behavioral acceptance:

- Starting `mcpvmcp` with streamable HTTP transport exposes endpoints like:

    http://127.0.0.1:8090/mcp/server/context7
    http://127.0.0.1:8090/mcp/tags/git,db

- Requests to `/mcp` (without selector) return HTTP 400 with a clear error message.
- Requests with conflicting URL and header selector return HTTP 400.
- Requests with a valid selector create a Gateway runtime for that selector and return MCP responses without mixing tools between selectors.

Test acceptance:

- Run `go test ./internal/infra/gateway` and expect all tests to pass, including new selector tests that fail before the change.
- Run `make lint-fix` and `make test` and expect successful completion.

Optional manual validation (requires a running core and at least one server named `context7`):

    curl -i http://127.0.0.1:8090/mcp
    # Expect: HTTP 400 and a message indicating selector is required.

    curl -i http://127.0.0.1:8090/mcp/server/context7 \
      -H 'Mcp-Protocol-Version: 2025-03-26' \
      -H 'Content-Type: application/json' \
      --data '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}'
    # Expect: HTTP 200 with a tool list for context7.

## Idempotence and Recovery

All changes are code edits and can be re-applied safely. If a refactor breaks compilation, revert only the local edits in the affected file and re-run tests. No data migrations are involved. The gateway can be restarted without special cleanup; GatewayPool eviction relies only on in-memory state.

## Artifacts and Notes

Expected selector error response example (HTTP 400):

    HTTP/1.1 400 Bad Request
    Content-Type: text/plain; charset=utf-8
    ...
    selector required: use /mcp/server/{name} or /mcp/tags/{tag1,tag2}

## Interfaces and Dependencies

New or changed interfaces to define:

- In `internal/infra/gateway/selector.go` define:

    type Selector struct {
        Server string
        Tags   []string
    }

    func ParseSelector(r *http.Request, basePath string) (Selector, error)
    func NormalizeTags(tags []string) []string
    func SelectorKey(sel Selector) string

- In `internal/infra/gateway/pool.go` define:

    type gatewayPool struct {
        // owns Gateway runtimes keyed by selector
    }

    func newGatewayPool(ctx context.Context, cfg rpc.ClientConfig, baseCaller string, logger *zap.Logger, opts PoolOptions) *gatewayPool
    func (p *gatewayPool) Get(ctx context.Context, sel Selector) (*mcp.Server, error)
    func (p *gatewayPool) Close(ctx context.Context) error

    type PoolOptions struct {
        IdleTimeout time.Duration
        MaxInstances int
    }

- In `internal/infra/gateway/core.go` add:

    func (g *Gateway) StartRuntime(ctx context.Context) error
    func (g *Gateway) StopRuntime(ctx context.Context) error
    func (g *Gateway) Server() *mcp.Server

- In `internal/infra/gateway/http_server.go` update `buildStreamableHTTPHandler` to:

    - Parse selector from URL + headers.
    - Resolve `*mcp.Server` via `GatewayPool`.
    - Inject the server into request context before calling `mcp.NewStreamableHTTPHandler`.

External dependencies:

- Use `github.com/modelcontextprotocol/go-sdk/mcp` streamable HTTP server and `auth` middleware as currently used.
- No new third-party libraries are required.

Plan Update Note: Marked implementation milestones complete after landing selector parsing, pool lifecycle, HTTP routing changes, CLI/UI/doc updates, and tests, and recorded the decision to keep stdio support via new selector flags with streamable HTTP as the default transport.
