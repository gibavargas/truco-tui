#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
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

cat <<'EOF' | sort >"$tmp_expected"
./apple-touch-icon.png
./assets/app.css
./assets/app.js
./favicon.ico
./favicon.png
./favicon.svg
./index.html
EOF
(cd "$DIST_DIR" && find . -type f ! -name 'truco-api' ! -name 'truco-api.exe' | sed -E 's/\.[a-f0-9]{8}\.css/.css/' | sed -E 's/\.[a-f0-9]{8}\.js/.js/' | sort) >"$tmp_actual"

if ! diff -u "$tmp_expected" "$tmp_actual"; then
  echo "browser dist contents do not match the static browser client layout" >&2
  exit 1
fi

api_bin="$DIST_DIR/truco-api"
if [[ ! -x "$api_bin" && -x "$DIST_DIR/truco-api.exe" ]]; then
  api_bin="$DIST_DIR/truco-api.exe"
fi

if command -v curl >/dev/null 2>&1 && [[ -x "$api_bin" ]]; then
  port="${TRUCO_BROWSER_SMOKE_PORT:-19090}"
  smoke_log="$(mktemp)"
  TRUCO_API_HOST=127.0.0.1 TRUCO_API_PORT="$port" "$api_bin" >"$smoke_log" 2>&1 &
  smoke_pid=$!
  cleanup() {
    kill "$smoke_pid" >/dev/null 2>&1 || true
    wait "$smoke_pid" >/dev/null 2>&1 || true
    rm -f "$smoke_log"
  }
  trap cleanup EXIT

  ready=0
  for _ in {1..40}; do
    if curl -fsS "http://127.0.0.1:${port}/" >/dev/null 2>&1; then
      ready=1
      break
    fi
    sleep 0.25
  done

  if [[ "$ready" -ne 1 ]]; then
    if grep -q "bind: operation not permitted" "$smoke_log"; then
      echo "skipping browser dist smoke server startup because local bind is not permitted in this environment" >&2
      cleanup
      trap 'rm -f "$tmp_expected" "$tmp_actual"' EXIT
      exit 0
    fi
    echo "browser dist smoke server did not become ready on port ${port}" >&2
    cat "$smoke_log" >&2
    exit 1
  fi

  curl -fsS "http://127.0.0.1:${port}/" >/dev/null
  curl -fsS "http://127.0.0.1:${port}/favicon.ico" >/dev/null
  curl -fsS "http://127.0.0.1:${port}/assets/app.css" >/dev/null
  curl -fsS "http://127.0.0.1:${port}/assets/app.js" >/dev/null

  cleanup
  trap 'rm -f "$tmp_expected" "$tmp_actual"' EXIT
fi
