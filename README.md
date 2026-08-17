# md-book-builder <!-- omit in toc -->

- [Purpose](#purpose)
- [How it works](#how-it-works)
- [Prerequisites](#prerequisites)
- [Quick start](#quick-start)
- [Configuration](#configuration)
- [Commands](#commands)
- [FAQ](#faq)
- [Build](#build)

## Purpose

`md-book-builder` assembles a selected set of self-referencing Markdown files
(a knowledge base, wiki, or documentation tree) into one clean PDF:

- cross-references between documents become clickable links,
- links to documents outside the selected set become plain text,
- page numbers (`Page X of Y`) and a PDF outline are added automatically,
- the GitHub look is preserved (headings, tables, checklists, admonitions),
- source files are only ever read, never modified.

The result is a single binary: build the book from the command line, or from
Automator/`Shortcuts` for an as-you-like click-to-build workflow.

## How it works

1. A YAML config lists the documents (globs) that form the book.
2. Markdown is rendered to HTML with GitHub-flavored tables, footnotes,
   task lists and `> [!NOTE]` / `> [!WARNING]` admonitions.
3. Links are resolved against the selected set: internal references point to
   the matching section anchor, external ones keep their label as text.
4. A table of contents (documents and/or their H2 sections) is generated.
5. Google Chrome (headless) exports the page as PDF with a page-number footer
   and a navigable document outline.

## Prerequisites

- Go 1.25+ (to build)
- A Chrome or Chromium binary (only external dependency at runtime)

## Quick start

```bash
go build -o md-book-builder ./cmd/md-book-builder
cp book.example.yaml book.yaml   # edit: point kb_root to your docs
./md-book-builder verify book.yaml   # check all links before rendering
./md-book-builder build book.yaml    # produce the PDF
```

Run `verify` after changing the knowledge base or the config. It reports
structural metrics (documents, TOC entries, links) and fails on any broken
internal anchor.

## Configuration

All relative paths in the config are resolved against the config file's
directory, so the config can live next to the binary while the knowledge base
stays elsewhere.

| key           | description                                                                                                                |
| ------------- | -------------------------------------------------------------------------------------------------------------------------- |
| `title`       | Book title shown on the cover page. Required.                                                                              |
| `subtitle`    | Optional subtitle on the cover page.                                                                                       |
| `kb_root`     | Knowledge base root directory. Required.                                                                                   |
| `out`         | Output PDF path. Required.                                                                                                 |
| `toc_levels`  | `[1]` documents only (default) or `[1, 2]` plus H2 sections.                                                               |
| `include`     | `**` globs of documents to assemble (`**/*.md` matches `README.md` too). Document order follows the order of the patterns. |
| `chrome_path` | Optional explicit Chrome/Chromium binary.                                                                                  |
| `locale`      | Footer language: `en` (default) or `ru`.                                                                                    |

See `book.example.yaml` for a commented template.

## Commands

```
md-book-builder build CONFIG.yaml [--out path] [--html path] [--open]
md-book-builder verify CONFIG.yaml [--html path]
md-book-builder expand CONFIG.yaml
md-book-builder version
```

- `build` renders the PDF. `--open` reveals it in the default viewer (handy
  for an Automator action). `--html` additionally writes the assembled HTML.
- `verify` validates the structure without launching Chrome.
- `expand` prints the ordered document list of the book as a ready-to-paste
  `include:` block. It collects documents with the same algorithms as `build`
  (globs, link-roots, excludes) but does not render anything, so the list can
  be reordered manually and pasted back into the config to control page order.
- The Chrome binary is auto-detected; override with the `chrome_path` config
  key or the `MD_BOOK_BUILDER_CHROME` environment variable.

## FAQ

**Why not Pandoc?** The knowledge base is a dense web of cross-references in
the form `[CODE]: ../path/file.md`. Pandoc does not turn those into working
in-PDF links without editing the source, which was not allowed.

**Are the source files modified?** No. Files are read-only; per-file reference
blocks are resolved in memory, and nothing is written back into the tree.

**I use a different wiki renderer.** Only GitHub-style anchors and admonitions
are emulated, so content using other constructs still renders — just without
the GitHub-specific flourishes.

## Build

See [BUILD.md](BUILD.md) for release/versioning instructions.
