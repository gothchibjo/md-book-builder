#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="$ROOT_DIR/dist"
MODULE="github.com/gothchibjo/md-book-builder"

VERSION=$(git -C "$ROOT_DIR" describe --tags --dirty --always 2>/dev/null || echo "dev")
LDFLAGS="-X $MODULE/internal/buildinfo.Version=$VERSION"

mkdir -p "$DIST_DIR"

build() {
  local goos=$1
  local goarch=$2
  local target=$3

  echo "Building $target for $goos/$goarch..."
  GOOS=$goos GOARCH=$goarch go build -trimpath -ldflags "$LDFLAGS" \
    -o "$DIST_DIR/$target" "$ROOT_DIR/cmd/md-book-builder"
}

build darwin amd64 md-book-builder_darwin_amd64
build darwin arm64 md-book-builder_darwin_arm64
build linux  amd64 md-book-builder_linux_amd64
build linux  arm64 md-book-builder_linux_arm64
build windows amd64 md-book-builder_windows_amd64.exe

printf "\nVersion: %s\n" "$VERSION"