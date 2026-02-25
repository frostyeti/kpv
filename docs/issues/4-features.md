---
type: feature
tags: ["feature", "proposal"]
---

# Feature Proposal Backlog

## 1) Vault path aliases and OS secret management commands

### Why
`README.md` documents `kpv config aliases` and `kpv config secret` flows, but current command files only expose `config get/set/rm`. Adding first-class commands closes the docs/runtime gap and improves day-to-day usability.

### Implementation Plan
- Add `kpv config aliases {set|get|ls|rm}` command group backed by existing Viper alias storage.
- Add `kpv config secret {set|get|rm}` command group for keyring-backed vault password management.
- Support path normalization (`kpv:///absolute/path`) and explicit target vault selectors.
- Add unit and integration tests for alias resolution and keyring fallback behavior.

### Acceptance Criteria
- Alias commands are available in `--help` and resolve vault names consistently.
- Secret commands store/read/remove passwords from supported OS keyring backends.
- README command coverage matches implemented commands.

## 2) Structured exit codes and non-interactive UX hardening

### Why
Several command paths call `utils.Failf` directly, which exits immediately and makes automation behavior harder to test and reason about. Standardized error handling improves scripting reliability.

### Implementation Plan
- Refactor command `Run` handlers to return errors through a shared execution helper.
- Define stable exit code categories (usage error, auth/password error, I/O error, not found, validation error).
- Add `--no-color` support for logs intended for CI and scripts.
- Add regression tests for exit codes and stderr output contracts.

### Acceptance Criteria
- Core commands emit deterministic exit codes.
- Error text remains human-readable while scripts can branch on exit code.
- Command tests can validate failures without process-level flakiness.

## 3) Secret metadata and rotation workflows

### Why
`kpv` already supports ensuring/generated secrets and sync semantics. Extending this with metadata-aware rotation enables safer long-lived secret management.

### Implementation Plan
- Add metadata fields (owner, environment, rotated_at, expires_at, tags) to managed entry strings/tags.
- Add `kpv secrets rotate` with filters (`--tag`, `--older-than`, `--key`) and dry-run mode.
- Integrate rotation with import/sync schemas so policy can be declared in JSON.
- Provide export visibility for metadata to support reporting.

### Acceptance Criteria
- Rotation command updates only matching entries and records rotation metadata.
- Dry-run shows a clear change plan.
- Import/export/sync round-trip metadata reliably.

## 4) Performance and path lookup correctness improvements

### Why
Path traversal and delimiter handling in `internal/keepass` are central to command correctness. A focused internal refactor will improve reliability for nested groups and large vaults.

### Implementation Plan
- Fix delimiter split behavior and add path parsing tests for `/`, `\`, `.`, `:` forms.
- Build an internal index map for group/entry lookup to reduce repeated tree scans.
- Add benchmarks for `FindEntry`, `UpsertEntry`, and sync-heavy workloads.
- Validate behavior with nested group fixtures and case-insensitive lookup tests.

### Acceptance Criteria
- Nested path operations are correct across supported delimiters.
- Lookup-heavy operations show measurable speedup in benchmarks.
- Existing command behavior remains backward-compatible.
