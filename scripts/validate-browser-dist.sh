#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PHP_DIR="$ROOT_DIR/browser-edition/php"
DIST_DIR="$ROOT_DIR/browser-edition/dist"

if [[ ! -d "$DIST_DIR" ]]; then
  echo "browser dist missing: $DIST_DIR" >&2
  exit 1
fi

if [[ ! -f "$DIST_DIR/truco-api" && ! -f "$DIST_DIR/truco-api.exe" ]]; then
  echo "browser dist missing compiled API binary (truco-api)" >&2
  exit 1
fi

if find "$DIST_DIR" -type f | grep -E '/[^/]* 2(\.[^/]+)?$' >/dev/null 2>&1; then
  echo "browser dist contains duplicate-looking files:" >&2
  find "$DIST_DIR" -type f | grep -E '/[^/]* 2(\.[^/]+)?$' >&2
  exit 1
fi

tmp_expected="$(mktemp)"
tmp_actual="$(mktemp)"
trap 'rm -f "$tmp_expected" "$tmp_actual"' EXIT

(cd "$PHP_DIR" && find . -type f | sort) >"$tmp_expected"
(cd "$DIST_DIR" && find . -type f ! -name 'truco-api' ! -name 'truco-api.exe' | sort) >"$tmp_actual"

if ! diff -u "$tmp_expected" "$tmp_actual"; then
  echo "browser dist contents do not match browser-edition/php" >&2
  exit 1
fi
