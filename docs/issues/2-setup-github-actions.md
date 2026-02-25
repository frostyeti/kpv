---
tags: ["chore", "ci-cd"]
type: chore
---

# Setup GitHub Actions

## Implementation Plan

1. Create baseline CI workflows
   - Add `.github/workflows/ci.yml` triggered on pull requests and pushes to main branches.
   - Use `mise` + Go `1.25.5` setup to match local tooling.
   - Run: `go mod tidy` check, `go fmt` verification, unit tests, integration tests (`-tags=integration`), `go lint`, and `govulncheck`.

2. Add build validation across OS/architecture targets
   - Add a matrix build job for `linux`, `windows`, `macos` on `amd64` and `arm64`.
   - Ensure the CLI compiles cleanly per matrix entry (`go build ./...` or target binary build).
   - Upload build artifacts for inspection in workflow runs.

3. Configure release automation with non-premium GoReleaser features
   - Add `.goreleaser.yaml` for cross-platform archives and checksums.
   - Publish Homebrew formula and Scoop/Chocolatey-compatible metadata where supported by non-pro features.
   - Gate package targets that require external infrastructure (Docker Hub, distro repos, Snap, Flatpak, AppImage hosting) behind follow-up tasks; track blockers explicitly.

4. Add release workflow and GitHub release announcement
   - Add `.github/workflows/release.yml` triggered on semver tags.
   - Run `goreleaser release --clean` with `GITHUB_TOKEN`.
   - Ensure release notes are generated and published as the GitHub release announcement.

5. Add observability steps for quality signals
   - Add lint and vulnerability jobs as required checks.
   - Add coverage generation (`go test -coverprofile=coverage.out ./...`) but keep the Codecov upload step commented until Codecov setup is complete.
   - Save coverage and test artifacts under workflow artifacts for troubleshooting.

## Delivery Breakdown

- Phase 1: CI quality gates (fmt/tidy/test/lint/vuln).
- Phase 2: multi-platform build matrix.
- Phase 3: GoReleaser config + tag-based releases.
- Phase 4: packaging ecosystem expansion (brew/choco/deb/rpm/snap/flatpak/appimage/docker) as infrastructure becomes available.

## Acceptance Criteria

- PRs run CI and fail on formatting, lint, vuln, or test regressions.
- Tags create GitHub releases with binaries and checksums via GoReleaser.
- Build matrix verifies Linux/Windows/macOS and amd64/arm64 compatibility.
- Coverage command exists in CI and Codecov upload is clearly staged/commented for later enablement.
