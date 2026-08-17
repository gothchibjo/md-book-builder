# AGENTS

## PROJECT-SPECIFIC

Project: `md-book-builder` — a small Go CLI that assembles selected
self-referencing Markdown files of a knowledge base into a single PDF. The
source documents are never modified; links to documents outside the selected
set become plain text.

High-level structure:

- `cmd/md-book-builder`: cobra CLI (`build`, `verify`, `expand`, `version`).
- `internal/config`: YAML book configuration (see `book.example.yaml`).
- `internal/source`: knowledge-base scanning, frontmatter parsing, ordering,
  pattern classification and per-document loading.
- `internal/markup`: GitHub-flavored markdown rendering, heading anchors,
  admonitions and link extraction (goldmark + AST transformer).
- `internal/book`: assembly — link-root collection by internal links, link
  fixing, frontmatter tables, TOC, final HTML. `Collect` runs the shared
  scan/follow/exclude pipeline reused by `build`, `verify` and `expand`.
- `internal/pdf`: headless Chrome printing (chromedp) and PDF output.
- `internal/assets`: embedded stylesheets (github-markdown.css + book.css).
- `scripts`: `build.sh` (cross-compiles all targets) and `test.sh` (vet + test).
- `.github/workflows/release.yml`: builds binaries and publishes a GitHub
  Release when a `v*` tag is pushed.

Engineering priorities:

- Match GitHub renderer behavior for headings/anchors and `> [!NOTE]` alerts.
- Respect that source documents are read-only.
- Link-root collection: a plain directory entry in `include` (e.g. `docs/team/`)
  collects only documents referenced by internal links; the directory itself
  is never browsed. `transitive_links` (default true) controls whether pulled
  documents are followed further.
- Keep `verify` runnable without Chrome (pure Go) so CI can validate links.
- Keep the binary self-contained: the only external dependency at runtime is
  a Chrome/Chromium binary.

----- END PROJECT-SPECIFIC -----

Repository guidelines and working agreements for AI agents.

## Language

All code comments and all documentation (including internal docs) must be
written in English.

## Architecture

Keep a simple layering: config → source → markup → book → pdf, where each
package depends only on the ones before it. `cmd/` is the only place that
wires them together.

Rules:

- No business logic inside `cmd/`; commands only parse flags and call packages.
- HTML parsing/manipulation lives in `internal/book` (via `golang.org/x/net/html`).
- Rendering concerns (goldmark, slugging, admonitions) live in `internal/markup`.

## Commit Messages

Use conventional commits. Every commit must include a body with change bullets
and a short purpose paragraph (including merge commits):

- Format: `type(scope): imperative summary`.
- No trailing period in the subject.
- Common types: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`.
- `scope` is required, short and specific.
- Body required: blank line, then bullet list describing key changes.
- After the bullet list, add a short purpose paragraph.
- Breaking changes: use `!` and/or footer `BREAKING CHANGE: ...`.

Example:

```text
fix(book): collect nodes before rewriting links

- Snapshot <a>/<img> nodes before mutating the HTML tree
- Fix iteration that silently skipped links after a flatten

The tree walker invalidated NextSibling during Flatten, so some
references were left unprocessed.
```

## Git Workflow

- The JS prototype lives on the local-only branch `legacy-js` and must not be
  force-pushed or merged back.
- Repo content is sanitized for public distribution; never (re)introduce
  company, customer or personal data into tracked files.

## Version Tags

- Tags use semantic versions (`vMAJOR.MINOR.PATCH`) and must be annotated.
- Update `CHANGELOG.md` before creating a release tag.

## Tests

- Prefer tests for stable behavior: slugging, config defaults, source
  ordering, and link fixing/anchoring.
- Do not write tests that require Chrome (they would fail in CI).

## Markdown Style

- Surround lists with blank lines.
- Surround fenced code blocks with blank lines.
- Always specify a language for fenced code blocks.

## Default Change Policy

- Avoid broad refactors without an explicit request.
- Keep the intro `## PROJECT-SPECIFIC` section current as the project evolves.