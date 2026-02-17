# Implement mcpvctl CLI and RPC Authentication

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

Follow the requirements in `/.agents/PLANS.md` from the repository root. This document must be maintained in accordance with that file.

## Purpose / Big Picture

After this change, operators can control the mcpv core through a dedicated CLI client (`mcpvctl`) over gRPC, locally or remotely, with standard authentication. The core continues to run as the `mcpv` service process; the CLI is a pure client. Authentication is optional but standard: token-based metadata or mutual TLS. Users can verify the behavior by starting `mcpv serve`, running `mcpvctl info`, `mcpvctl tools list`, and observing that authenticated RPC endpoints reject unauthenticated requests when enabled.

## Progress

- [x] (2026-02-17 01:40Z) Add RPC auth configuration to domain types, raw config, normalization, and schema.
- [x] (2026-02-17 01:45Z) Enforce RPC auth on the server and attach auth metadata in the client.
- [x] (2026-02-17 02:05Z) Implement `mcpvctl` CLI with full control-plane coverage and register/unregister lifecycle.
- [x] (2026-02-17 02:10Z) Extend gateway CLI flags to pass RPC auth token/mTLS.
- [x] (2026-02-17 02:15Z) Update configuration examples and docs to describe the CLI and auth behavior.
- [x] (2026-02-17 03:05Z) Add tests for auth validation and interceptors; run `make lint-fix` and `make test`.

## Surprises & Discoveries

- gRPC `NewClient` requires a passthrough target (`passthrough:///...`) when used with `bufconn`.

## Decision Log

- Decision: Introduce a dedicated client binary named `mcpvctl` rather than overloading `mcpv` with client subcommands.
  Rationale: Separating server and client binaries keeps responsibilities clear, avoids accidental server/client coupling, and aligns with standard control-plane tooling patterns.
  Date/Author: 2026-02-17 / Codex

- Decision: Implement token authentication via gRPC metadata `authorization: Bearer <token>` and mTLS using existing `rpc.tls.clientAuth`.
  Rationale: Both approaches are standard and interoperate with typical infra setups. Token is simpler to bootstrap; mTLS is stronger for enterprise use.
  Date/Author: 2026-02-17 / Codex

- Decision: Default CLI behavior registers a caller before performing control-plane operations and unregisters on exit.
  Rationale: The control plane expects caller registration to resolve visibility and lifecycle. Automatic registration reduces foot-guns for automation users.
  Date/Author: 2026-02-17 / Codex

## Outcomes & Retrospective

Delivered `mcpvctl`, RPC auth config + enforcement, gateway auth flags, schema/docs updates, and tests. `make lint-fix` and `make test` both pass.

## Context and Orientation

The mcpv core runs from `cmd/mcpv` and exposes a gRPC control plane in `proto/mcpv/control/v1/control.proto`. The control plane server is implemented in `internal/infra/rpc/server.go` and uses request interceptors from `internal/infra/rpc/request_interceptors.go`. Client connections are created in `internal/infra/rpc/client.go`. Configuration is loaded from a single YAML file by `internal/infra/catalog/loader/loader.go`, with runtime types defined in `internal/domain/types.go`, normalization in `internal/infra/catalog/normalizer/runtime.go`, and schema validation in `internal/infra/catalog/validator/schema.json`.

This plan adds a new CLI in `cmd/mcpvctl`, extends RPC configuration with authentication settings, and ensures the gateway `cmd/mcpvmcp` can authenticate when RPC auth is enabled.

## Plan of Work

First, add RPC authentication configuration to the domain and config normalization layers, including schema validation and example config updates. This introduces `rpc.auth` fields and enforces correct combinations of TLS and auth settings.

Second, implement RPC auth enforcement on the server via interceptors, and attach auth metadata on the client when configured. The server rejects unauthenticated calls when auth is enabled. The client attaches a Bearer token header when provided.

Third, build the `mcpvctl` CLI. It should provide full coverage for the existing control-plane RPC methods: info, register/unregister, tools list/watch/call, tasks list/get/result/cancel, resources list/watch/read, prompts list/watch/get, logs stream, runtime status watch, and server-init status watch. The CLI must register a caller (unless explicitly disabled), then perform the requested action, and unregister on exit. Output should be human-readable by default and support `--json` for machine parsing.

Fourth, extend the gateway CLI (`cmd/mcpvmcp`) to accept RPC auth flags and pass them into the RPC client configuration. This ensures the gateway can connect to a secured core.

Fifth, update documentation and examples to describe the new CLI and `rpc.auth` configuration, and add tests that validate auth config normalization and interceptor behavior using in-memory gRPC (`bufconn`) to avoid network dependency.

## Concrete Steps

All commands run from `/Users/wibus/dev/mcpd`.

1) Search for RPC config and client usage to identify touch points.
   rg -n "RPCConfig|rpc\." internal cmd proto docs

2) Implement domain and config changes.
   - Edit `internal/domain/types.go` to add `RPCAuthConfig` and attach it to `RPCConfig`.
   - Edit `internal/infra/catalog/normalizer/raw_types.go` and `internal/infra/catalog/normalizer/runtime.go` to parse and validate `rpc.auth`.
   - Edit `internal/infra/catalog/validator/schema.json` to add `auth` under `rpc`.
   - Update `dev/catalog.example.yaml` with commented `rpc.auth` examples.

3) Implement RPC auth enforcement and client injection.
   - Add auth interceptors in `internal/infra/rpc/request_interceptors.go` (unary + stream).
   - Wire interceptors in `internal/infra/rpc/server.go`.
   - Add auth configuration to `internal/infra/rpc/client.go` and attach Bearer token metadata.

4) Implement CLI client.
   - Create `cmd/mcpvctl/main.go` and supporting files for cobra commands and output formatting.
   - Add command groups for tools, resources, prompts, tasks, logs, runtime, and init status.
   - Use `internal/infra/rpc.Client` for all RPC calls.

5) Update gateway CLI flags.
   - Add `--rpc-token` / `--rpc-token-env` flags in `cmd/mcpvmcp/main.go` and pass into RPC client config.

6) Documentation and tests.
   - Update README to mention `mcpvctl`.
   - Update any relevant docs to show `rpc.auth` usage and the new CLI.
   - Add tests for auth normalization and interceptor enforcement.

7) Validation.
   - Run `make lint-fix` and `make test`.
   - Start core with `mcpv serve --config runtime.yaml`, then verify CLI behavior as described below.

## Validation and Acceptance

The change is accepted when:

- `mcpvctl info` succeeds against a running core without authentication when `rpc.auth.enabled` is false.
- When `rpc.auth.enabled: true` and `rpc.auth.mode: token`, RPC calls without a Bearer token fail with `Unauthenticated`, and the same calls succeed when the token is provided.
- When `rpc.auth.mode: mtls`, RPC calls succeed only with valid client certificates, and fail without them.
- `mcpvmcp` can connect to a secured core when `--rpc-token` (or mTLS flags) are supplied.
- `make lint-fix` and `make test` complete successfully.

Example verification transcript (tokens are illustrative only):

  # Start core with auth enabled
  MCPV_RPC_TOKEN=devtoken mcpv serve --config ./runtime.yaml

  # Fails without token
  mcpvctl --rpc unix:///tmp/mcpv.sock info
  # Expect: Unauthenticated

  # Succeeds with token
  mcpvctl --rpc unix:///tmp/mcpv.sock --rpc-token devtoken info

## Idempotence and Recovery

All code edits are additive and can be safely re-applied. If a change breaks config validation, revert the last edit to `internal/infra/catalog/normalizer/runtime.go` and `internal/infra/catalog/validator/schema.json`, then re-run tests to confirm the failure scope. If CLI behavior is incorrect, the CLI can be isolated by disabling or removing `cmd/mcpvctl` without affecting the core.

## Artifacts and Notes

Keep command outputs from `make lint-fix` and `make test` here if failures occur, along with any temporary workaround applied.

## Interfaces and Dependencies

At the end of this plan, the following interfaces must exist:

- In `internal/domain/types.go`:
  - `type RPCAuthConfig struct { Enabled bool; Mode RPCAuthMode; Token string; TokenEnv string }`
  - `type RPCAuthMode string` with `"token"` and `"mtls"` constants.
  - `RPCConfig` includes `Auth RPCAuthConfig`.

- In `internal/infra/rpc/client.go`:
  - `ClientConfig` includes auth fields (token or full `RPCAuthConfig`).
  - Client attaches `authorization: Bearer <token>` metadata when configured.

- In `internal/infra/rpc/request_interceptors.go`:
  - Auth interceptors enforce token or mTLS requirements when `rpc.auth.enabled` is true.

- In `cmd/mcpvctl`:
  - Root flags: `--rpc`, `--rpc-tls`, `--rpc-tls-cert`, `--rpc-tls-key`, `--rpc-tls-ca`, `--rpc-token`, `--rpc-token-env`, `--caller`, `--tag`, `--server`, `--json`, `--no-register`.
  - Subcommands: `info`, `register`, `unregister`, `tools list/watch/call`, `tasks list/get/result/cancel`, `resources list/watch/read`, `prompts list/watch/get`, `logs`, `runtime watch`, `init watch`.

- In `cmd/mcpvmcp`:
  - New flags for RPC token and token env, wired to client config.

Change note: Marked completed milestones after implementing auth config, server/client interceptors, gateway flags, CLI, docs, and tests. Lint/test still pending.
