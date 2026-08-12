# Changelog

All notable changes to this project are documented in this file.

## v0.2.0 - 2026-08-12

### Added

- GitHub Actions release workflow (`.github/workflows/release.yml`): on a `v*`
  tag push it runs vet and tests, cross-compiles the binaries, writes
  `SHA256SUMS` and publishes a GitHub Release with the artifacts.
- `linux/arm64` build target in `scripts/build.sh`.

## v0.1.0 - 2026-08-10

### Added

- Initial Go implementation of the markdown-to-PDF book builder, replacing
  the earlier Node.js prototype (kept only in git history on a legacy branch).
- `md-book-builder build` renders a PDF from a `book.yaml` config with page
  numbers (`Page X of Y`), a cover page, a table of contents and a document
  outline.
- `md-book-builder verify` checks the book structure (documents, TOC entries,
  links, broken anchors) without launching Chrome.
- GitHub-flavored rendering: tables, footnotes, task lists, `> [!NOTE]`
  admonitions, and multilingual heading anchors with per-book unique ids.
- Link resolution: internal cross-references become anchors, references to
  excluded/unknown documents render as plain text.
- `--open` flag for Automator/Shortcuts workflows, `--html` to dump the
  assembled HTML for debugging.
- Unit tests for slugging, config loading, source ordering and link fixing.

### Changed

- Documentation and sample config are sanitized for public distribution (no
  company or personal data).

### Removed

- Node.js source tree (`src/`, `scripts/`, `package.json`) that shipped the
  original generator.
