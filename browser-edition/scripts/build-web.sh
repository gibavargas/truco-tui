#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
WEB_DIR="$ROOT_DIR/browser-edition/web"
DIST_DIR="$ROOT_DIR/browser-edition/dist"
WASM_BIN="$DIST_DIR/main.wasm"

mkdir -p "$DIST_DIR"
rm -f "$DIST_DIR"/*.wasm "$DIST_DIR"/wasm_exec.js "$DIST_DIR"/index.html "$DIST_DIR"/app.js "$DIST_DIR"/style.css

echo "Compilando WASM..."
GOOS=js GOARCH=wasm go build -o "$WASM_BIN" ./browser-edition/cmd/wasm

GOROOT_PATH="$(go env GOROOT)"
if [[ -f "$GOROOT_PATH/lib/wasm/wasm_exec.js" ]]; then
  cp "$GOROOT_PATH/lib/wasm/wasm_exec.js" "$DIST_DIR/wasm_exec.js"
elif [[ -f "$GOROOT_PATH/misc/wasm/wasm_exec.js" ]]; then
  cp "$GOROOT_PATH/misc/wasm/wasm_exec.js" "$DIST_DIR/wasm_exec.js"
else
  echo "wasm_exec.js não encontrado em GOROOT=$GOROOT_PATH" >&2
  exit 1
fi

cp "$WEB_DIR/index.html" "$WEB_DIR/app.js" "$WEB_DIR/style.css" "$DIST_DIR/"

echo "Build concluído em: $DIST_DIR"
