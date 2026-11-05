#!/bin/bash
set -euo pipefail

RUNTIME_DIR="${1:-}"
OUTPUT="${2:-}"
if [[ -z "$RUNTIME_DIR" || -z "$OUTPUT" ]]; then
  echo "Usage: $0 <runtime-dir> <output.exe>" >&2
  exit 1
fi

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
BUILD_DIR="$(mktemp -d)"
trap 'rm -rf "$BUILD_DIR"' EXIT

cd "$PROJECT_ROOT"
GOOS=windows CGO_ENABLED=0 go build -ldflags="-s -w" \
  -o "$BUILD_DIR/launcher.exe" ./tools/windows-standalone
go run ./tools/windows-standalone-builder \
  -runtime "$RUNTIME_DIR" -launcher "$BUILD_DIR/launcher.exe" -output "$OUTPUT"
