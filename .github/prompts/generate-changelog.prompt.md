---
agent: agent
---

# Changelog Generation Prompt

Generate a CHANGELOG.md file in Keep a Changelog format (https://keepachangelog.com/en/1.1.0/) for this repository based on Git history using Conventional Commits format (https://www.conventionalcommits.org/en/v1.0.0/). Use semantic versioning with the following rules:

## Structure

- Header with "Changelog" title and format reference links
- `[Unreleased]` section at top (empty for releases)
- Version sections in reverse chronological order: `## [X.Y.Z] - YYYY-MM-DD`
- Bottom section with comparison links for all versions

## Version Detection

- Get versions from Git tags (format: `vX.Y.Z`)
- Extract release date from Git tag timestamp (format: YYYY-MM-DD)
- Group commits between consecutive tags (e.g., commits from v0.1.0 to v0.1.1)
- Sort versions in reverse chronological order

## Version Links (at bottom of file)

- `[Unreleased]`: Link comparing latest tag to HEAD: `https://github.com/{owner}/{repo}/compare/vX.Y.Z...HEAD`
- `[X.Y.Z]`: Link comparing to previous tag: `https://github.com/{owner}/{repo}/compare/vPREV...vX.Y.Z`
- First version: Link to release tag: `https://github.com/{owner}/{repo}/releases/tag/vX.Y.Z`

## Commit Parsing (Conventional Commits Format)

- Format: `<type>[optional scope][!]: <description>`
- Breaking changes indicated by `!` after type/scope (e.g., `refactor!:`, `feat(api)!:`)
- Types: `feat`, `fix`, `refactor`, `chore`, `docs`, `test`, `ci`, `build`, etc.

## Categorization Rules

- **Added**: `feat:` commits that introduce new features/capabilities
- **Changed**: `refactor:` (non-breaking) or behavior-modifying `chore:` commits
- **Fixed**: `fix:` commits
- **Removed**: Commits that remove features/endpoints
- **Breaking changes**: Mark with `**BREAKING:**` prefix when `!` is present after type/scope

## Formatting

- Link all PRs: `([#123](https://github.com/{owner}/{repo}/pull/123))`
- Extract PR numbers from commit messages (format: `(#123)`)
- Group multiple related PRs in one entry if they implement the same feature
- Use concise descriptions, remove type/scope prefixes (e.g., `feat:`, `fix:`, `refactor:`)
- Capitalize first word of each entry

## Exclusions

- Skip dependency updates (`chore(deps):`, `fix(deps):`)
- Skip pure refactoring/internal changes (`refactor:`, `test:`, `ci:`, `build:`) unless user-facing
- For versions with ONLY excluded changes, add note: `**Note:** This release contains only internal refactoring and dependency updates.`
- Skip `release:` commits

## Examples of Correct Entries

- `Add Prometheus metrics support ([#237](https://github.com/{owner}/{repo}/pull/237))` (from `feat: add Prometheus metrics support (#237)`)
- `**BREAKING:** Remove /v1/metrics endpoint (JSON format). Use /metrics (Prometheus format) ([#279]..., [#280]...)` (from `feat!: remove /v1/metrics endpoint (#279)`)
- `Log version during initialization ([#164]...)` (from `feat: log version during initialization (#164)`)
- `Combine request_allowed and request_valid log fields into request_status ([#223]...)` (from `refactor: combine ... (#223)`)

## Task

Generate the complete changelog following this exact format, processing all Git tags from earliest to latest.
