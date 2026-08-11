# Changelog

All notable changes to this project are documented in this file.

## v0.1.0 - 2026-08-10

### Added

- Initial Go implementation of the markdown-to-PDF book builder, replacing
  the earlier Node.js prototype (kept only in git history on a legacy branch).
- `md-book-builder build` renders a PDF from a `book.yaml` config with page
  numbers (`стр. X из Y`), a cover page, a table of contents and a document
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
