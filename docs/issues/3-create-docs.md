---
type: docs
tags: ["docs"]
---

# Create Docs

## Implementation Plan

1. Clarify product scope and information architecture
   - Normalize docs terminology to `kpv` (current repository) and separate any future `cast` references into their own project docs.
   - Define top-level sections: Getting Started, CLI Commands, Secret Workflows, Vault Path/Password Resolution, Formats/Integrations, Troubleshooting, and Release Notes.
   - Add a versioning strategy so docs can evolve with new command behavior over time.

2. Bootstrap an Astro + Starlight docs site inside `docs/`
   - Scaffold `docs/site` with Astro and Starlight.
   - Enable built-in search and RSS feeds.
   - Configure navigation, sidebar groups, edit links, and API reference pages per command.

3. Author core docs from existing CLI behavior
   - Convert current `README.md` command content into structured pages.
   - Add examples for: `init`, `secrets set/get/get-string/ls/rm/ensure/import/export/sync`, and `config get/set/rm`.
   - Include format-specific examples (`json`, `dotenv`, shell export, CI-oriented output).

4. Add advanced workflow guides and recipes
   - Create task-oriented guides for bootstrapping a vault, rotating credentials, syncing from JSON, and using environment-driven automation.
   - Add OS-specific notes for keyring behavior and default vault locations.
   - Include common failure diagnostics (missing password, vault path resolution, malformed import/sync data).

5. Automate docs quality and publishing
   - Add docs build and link-check steps in CI.
   - Add deployment workflow for Cloudflare Pages (or equivalent static host) with preview deployments for PRs.
   - Add a release process to update docs with each tagged binary release.

## Acceptance Criteria

- A working Astro + Starlight site exists under `docs/` and builds in CI.
- Search and RSS are enabled and validated.
- All current user-facing commands are documented with examples.
- Docs deployment is automated and versioning strategy is documented for future releases.
