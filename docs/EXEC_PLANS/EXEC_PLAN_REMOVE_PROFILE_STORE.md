# Remove Deprecated Profile Store Artifacts

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

Follow the requirements in `/.agents/PLANS.md` from the repository root. This document must be maintained in accordance with that file.

## Purpose / Big Picture

The goal is to fully remove the obsolete “profile store” configuration model (profiles directory plus callers mapping) from the codebase, and to align user-facing guidance with the single-file configuration model that the current runtime already uses. After this change, the core no longer ships any profile-store loaders or types, CLI help no longer suggests a directory-based profile store, and documentation clearly indicates that a single YAML config file is the only supported configuration format.

## Progress

- [x] (2026-02-17 00:35Z) Inventory and remove profile store code paths and domain types that are unused by the single-file loader.
- [x] (2026-02-17 00:40Z) Update CLI defaults and in-code comments to describe a single config file instead of a profile store directory.
- [x] (2026-02-17 00:50Z) Update or deprecate documentation that still describes profiles/callers and profile store layout.
- [ ] (2026-02-17 01:05Z) Run `make lint-fix` and `make test` and record results (completed: commands executed; remaining: rerun in an environment that allows Go build cache access and local test listeners).

## Surprises & Discoveries

- Observation: `make lint-fix` fails because golangci-lint reports “no go files to analyze,” suggesting a context loading issue in the current environment.
  Evidence: `golangci-lint run --config .golangci.yml --fix` → `context loading failed: no go files to analyze`.

- Observation: `make test` fails in this environment due to permission errors when Go tries to access the build cache and bind HTTP test listeners.
  Evidence: errors such as `open /Users/wibus/Library/Caches/go-build/...: operation not permitted` and `httptest: failed to listen on a port: bind: operation not permitted`.

## Decision Log

- Decision: Treat the profile store implementation as dead code and remove it entirely, rather than keeping a compatibility shim.
  Rationale: The active loader already parses a single YAML config file, and there are no code references to the profile store types or loader. Removing the dead code reduces confusion and aligns with the current product direction.
  Date/Author: 2026-02-17 / Codex

- Decision: Add deprecation notices to historical documentation instead of rewriting every profile-store-specific section.
  Rationale: The profile store content is extensive and no longer representative of the running system. Deprecation notices prevent users from following outdated guidance while keeping historical design context available.
  Date/Author: 2026-02-17 / Codex

## Outcomes & Retrospective

Not completed yet.

## Context and Orientation

The current runtime loads configuration from a single YAML file using `internal/infra/catalog/loader/loader.go` and related normalizer/validator code. The obsolete profile store implementation lives in `internal/infra/catalog/store/` and defines a `ProfileStore` type in `internal/domain/profile.go`. These files describe and load a directory layout (`profiles/*.yaml`, `callers.yaml`, optional `runtime.yaml`) that is no longer used. The CLI entry point in `cmd/mcpv/main.go` still describes `--config` as a profile store directory and defaults to `.` even though the loader expects a file path. Several documentation files (for example `docs/PRD.md` and `docs/UX_REDUCTION.md`) still describe the profile store model and need to be updated or clearly marked as deprecated.

In this plan, “single-file configuration” means one YAML file that contains runtime settings (top-level fields like `routeTimeoutSeconds`, `rpc`, `observability`, `subAgent`, etc.) plus `servers:` and optional `plugins:`. An example lives at `dev/catalog.example.yaml`.

## Plan of Work

First, remove the unused profile store code and tests by deleting `internal/domain/profile.go` and the entire `internal/infra/catalog/store/` package. Then update CLI defaults and in-code comments to refer to a config file rather than a profile store directory, and to default `--config` to `runtime.yaml` in the current working directory. Next, update the most user-facing docs to either reflect the single-file config or clearly mark the profile-store content as deprecated. At minimum, update `docs/PRD.md` to show a single-file config example, remove profile-store references from `docs/WAILS_BINDINGS.md`, and add a deprecation notice to documents that remain historically useful but obsolete (such as `docs/CONFIG_VISUALIZATION_DESIGN.md`, `docs/UX_REDUCTION.md`, `docs/APP_SCOPE.md`, and `docs/config_migration/profile_store_migration.md`). Finally, run `make lint-fix` and `make test` to ensure the repo still builds cleanly.

## Concrete Steps

From the repository root, remove the unused profile store code and adjust comments and CLI help. Then update documentation and run lint/tests. Example commands (to be run from `/Users/wibus/dev/mcpd`) are:

  rg -n "ProfileStore|profilesDirName|callers.yaml" internal cmd docs
  # Delete profile store code and tests.
  # Update CLI and comments in place.
  make lint-fix
  make test

Record any failing command output in this plan and retry after fixes.

## Validation and Acceptance

Acceptance is reached when:

The project builds and tests pass after `make lint-fix` and `make test`.

Running `mcpv --help` shows `--config` described as a config file path (not a profile store directory), and the default is `runtime.yaml` in the current directory unless overridden.

The documentation no longer instructs users to create `profiles/` or `callers.yaml`, and deprecated design documents are clearly marked as historical-only so they do not mislead current users.

## Idempotence and Recovery

The file deletions and edits in this plan are safe to re-run because the removed files are not referenced by buildable code paths. If a deletion was premature, reintroduce the file from git history and re-run `make test` to confirm. Documentation edits are safe to reapply because they are additive or replace obsolete guidance with current wording.

## Artifacts and Notes

None yet.

## Interfaces and Dependencies

No new external dependencies are introduced. The remaining configuration interface is the single YAML config file parsed by `internal/infra/catalog/loader/loader.go`, which must continue to accept top-level runtime fields plus `servers` and `plugins` arrays. The CLI flag `--config` in `cmd/mcpv/main.go` must accept a file path that is consumed by the loader. No other interfaces should rely on profile store artifacts after this change.

Change note: Recorded lint/test failures in progress and surprises after running `make lint-fix` and `make test`.
