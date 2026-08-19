# Changelog

All notable changes to this project are documented in this file.

## v0.3.0 - 2026-08-19

### Added

- Link-root collection: a plain directory entry in `include` (e.g.
  `docs/team/`) now pulls in only the documents referenced by internal links,
  without browsing the directory. The `transitive_links` config option
  (default true) controls whether pulled documents are followed further.
- `exclude` config key to drop specific documents, folders or masks after
  include and link-root collection; a pattern without a `.md` suffix also
  matches the file with `.md` appended.
- `md-book-builder expand` command that prints the collected documents as an
  include block, so the current page order can be frozen and reordered by
  editing the generated block.
- `-o` shorthand for `--out` on the build command.

### Fixed

- Cover and TOC text now inherit the GitHub sans-serif font stack used by the
  markdown body instead of falling back to the browser default serif.

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
