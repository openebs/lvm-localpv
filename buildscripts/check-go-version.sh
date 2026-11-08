#!/usr/bin/env bash

# Ensures every pinned Go version in the repo matches the go.mod toolchain directive.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

gomod_version="$(awk '/^toolchain go/ { sub(/^go/, "", $2); print $2 }' "$ROOT_DIR/go.mod")"
if [[ -z "$gomod_version" ]]; then
  gomod_version="$(awk '/^go / { print $2 }' "$ROOT_DIR/go.mod")"
fi

rc=0

check() {
  local what="$1" file="$2" found="$3" expected="$4" source="$5"
  if [[ "$found" != "$expected" ]]; then
    echo "ERROR: $file uses $what '$found', expected '$expected' (from $source)"
    rc=1
  else
    echo "OK: $file uses $what $found"
  fi
}

dockerfile="buildscripts/lvm-driver/Dockerfile.buildx"
check Go "$dockerfile" "$(sed -n 's/^FROM golang:\([0-9.]*\).*/\1/p' "$ROOT_DIR/$dockerfile" | head -n1)" "$gomod_version" go.mod

exit "$rc"
