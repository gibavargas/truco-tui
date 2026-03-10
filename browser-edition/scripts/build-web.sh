#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
PHP_DIR="$ROOT_DIR/browser-edition/php"
DIST_DIR="$ROOT_DIR/browser-edition/dist"
API_BIN="$DIST_DIR/truco-api"

mkdir -p "$DIST_DIR"
rm -rf "$DIST_DIR"/*

echo "Compiling Go HTTP API server..."
go build -o "$API_BIN" ./browser-edition/cmd/httpapi

echo "Copying PHP files..."
cp -r "$PHP_DIR"/* "$DIST_DIR/"

echo "Build complete."
echo "  API binary: $API_BIN"
echo "  PHP files:  $DIST_DIR/"
echo ""
echo "To run locally:"
echo "  1. Start Go API:  $API_BIN"
echo "  2. Start PHP:     php -S localhost:8080 -t $DIST_DIR"
echo "  3. Open:          http://localhost:8080/index.php"
