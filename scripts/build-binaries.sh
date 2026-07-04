#!/usr/bin/env bash
set -euo pipefail

BINARY_DIR="$(dirname "$0")/../binary"
CMD="./cmd/myserv"

mkdir -p "$BINARY_DIR"

  echo "==> Building jsonserv binaries..."

build() {
  local GOOS="$1" GOARCH="$2" SUFFIX="${3:-}"
  local NAME="jsonserv-${GOOS}-${GOARCH}${SUFFIX}"
  echo "    $GOOS/$GOARCH -> binary/$NAME"
  GOOS="$GOOS" GOARCH="$GOARCH" CGO_ENABLED=0 go build -ldflags="-s -w" -o "$BINARY_DIR/$NAME" "$CMD"
}

build linux   amd64
build linux   arm64
build darwin  amd64
build darwin  arm64
build windows amd64 ".exe"
build windows arm64 ".exe"

echo "==> Done. Binaries in binary/:"
ls -lh "$BINARY_DIR"
