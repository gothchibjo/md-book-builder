# Build & Release Notes

This document is for maintainers who build and distribute binaries.

## Build

### Quick build

```bash
./scripts/build.sh
```

The script stamps the version from `git describe --tags --dirty --always` and
writes binaries to `dist/`: darwin (amd64/arm64), linux (amd64) and windows
(amd64).

### Manual build

```bash
go build -ldflags "-X github.com/gothchibjo/md-book-builder/internal/buildinfo.Version=v0.1.0" \
  -o dist/md-book-builder ./cmd/md-book-builder
```

Cross-compile for Apple Silicon from an Intel host and vice versa:

```bash
GOOS=darwin GOARCH=arm64 go build -o dist/md-book-builder_darwin_arm64 ./cmd/md-book-builder
GOOS=darwin GOARCH=amd64 go build -o dist/md-book-builder_darwin_amd64 ./cmd/md-book-builder
```

## Runtime prerequisites

The only external component is Chrome/Chromium. Detection order:

1. `MD_BOOK_BUILDER_CHROME` environment variable
2. `chrome_path` in the book config
3. macOS well-known install locations
4. `google-chrome` on PATH

## Versioning

- Tags use semantic versions (`vMAJOR.MINOR.PATCH`) and must be annotated.
- Update `CHANGELOG.md` before creating a release tag.

## CI usage

```
./scripts/build.sh
./scripts/test.sh   # go vet ./... && go test ./...
```

Binaries are static Go executables; only the Chrome binary is fetched at
runtime from the machine that runs the build.