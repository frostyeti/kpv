---
tags: ["chore", "testing"]
type: chore
---

# Implement tests

## Implementation Plan

1. Establish a test strategy and folder layout
   - Keep fast unit tests close to packages (`cmd`, `internal/keepass`, `internal/utils`).
   - Add integration tests in separate files with `// +build integration` and run them with `go test -tags=integration ./...`.
   - Add end-to-end tests under `test/e2e` that execute the built `kpv` binary against temporary vaults and clean up all test data.

2. Expand unit test coverage for command and utility logic
   - Add table-driven tests for argument/flag validation in command handlers (`secrets set/get/rm/import/sync`, `init`, `config`).
   - Add tests for vault path resolution, default path behavior, and alias handling in `internal/utils`.
   - Refactor small command internals where needed (dependency injection for I/O and KeePass open/save calls) so logic is testable without `os.Exit` side effects.

3. Add integration tests for real KeePass file behavior
   - Use temp directories and real `.kdbx` files to verify create/open/save/read flows across command boundaries.
   - Validate import/sync semantics (create, update, ensure generation, delete, dry-run) against known JSON fixtures.
   - Verify password lookup order (`--vault-password`, `--vault-password-file`, env, key file) and failure modes.

4. Add e2e command workflow tests
   - Cover common user paths: `init`, `secrets set/get/ls/rm`, `secrets export/import/sync`, and `config get/set/rm`.
   - Assert exit codes, stdout/stderr shape, and format outputs (`text`, `json`, `dotenv`, shell exports, GitHub Actions format).
   - Ensure tests are hermetic: isolated temp home/config dirs, deterministic fixtures, and no dependency on a user keyring.

5. Wire test execution into local and CI workflows
   - Add a standard local sequence: `go test ./...` and `go test -tags=integration ./...`.
   - Keep e2e in a dedicated step so failures are easy to triage.
   - Publish coverage output (when CI coverage is enabled) and gate merges on unit + integration success first.

## Acceptance Criteria

- `go test ./...` passes reliably on a clean checkout.
- `go test -tags=integration ./...` passes and exercises real `.kdbx` file operations.
- `test/e2e` covers core user workflows and passes in CI.
- Critical paths (secret CRUD, sync/import/export, path/password resolution) have meaningful assertions rather than smoke-only tests.
