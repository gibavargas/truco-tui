#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
WEB_DIR="$ROOT_DIR/browser-edition/web"
PACKAGE_JSON="$ROOT_DIR/browser-edition/package.json"
PACKAGE_LOCK="$ROOT_DIR/browser-edition/package-lock.json"
DIST_DIR="$ROOT_DIR/browser-edition/dist"
API_BIN="$DIST_DIR/truco-api"
API_SRC_DIR="$ROOT_DIR/browser-edition/cmd/httpapi"

missing=()
[[ -d "$WEB_DIR" ]] || missing+=("browser-edition/web")
[[ -f "$PACKAGE_JSON" ]] || missing+=("browser-edition/package.json")
[[ -f "$PACKAGE_LOCK" ]] || missing+=("browser-edition/package-lock.json")
[[ -d "$API_SRC_DIR" ]] || missing+=("browser-edition/cmd/httpapi")

if (( ${#missing[@]} > 0 )); then
  echo "Build aborted: missing required browser edition paths:"
  for path in "${missing[@]}"; do
    echo "  - $path"
  done
  echo ""
  echo "This repository version expects the TypeScript browser client and Go HTTP API."
  exit 1
fi

mkdir -p "$DIST_DIR"
rm -rf "$DIST_DIR"/*

echo "Installing browser edition dependencies from lockfile..."
npm ci --prefix "$ROOT_DIR/browser-edition"

echo "Bundling TypeScript browser client..."
TRUCO_BROWSER_OUTDIR="$DIST_DIR" npm run build --prefix "$ROOT_DIR/browser-edition"

echo "Compiling Go HTTP API server..."
go build -o "$API_BIN" ./browser-edition/cmd/httpapi

"$ROOT_DIR/scripts/validate-browser-dist.sh"

echo "Build complete."
echo "  API binary: $API_BIN"
echo "  Web assets: $DIST_DIR/"
echo ""
echo "To run locally:"
echo "  1. Start Go app:  $API_BIN"
echo "  2. Open:          http://localhost:9090/"
