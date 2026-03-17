#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
status=0

check_path_regex() {
  local path="$1"
  local regex="$2"
  local label="$3"
  local base
  base="$(basename "$path")"
  if [[ ! "$base" =~ $regex ]]; then
    echo "invalid $label artifact name: $path" >&2
    status=1
  fi
}

if [[ -d "$ROOT_DIR/bin/tui" ]]; then
  while IFS= read -r path; do
    check_path_regex "$path" '^truco-tui-core-[a-z0-9]+-[a-z0-9]+(-portable)?(\.exe)?$' "TUI"
  done < <(find "$ROOT_DIR/bin/tui" -mindepth 1 -maxdepth 1 \( -type f -o -type d \) | sort)
fi

if [[ -d "$ROOT_DIR/bin/gui" ]]; then
  while IFS= read -r path; do
    check_path_regex "$path" '^truco-gui-[a-z0-9]+-[a-z0-9]+-[a-z0-9]+(-portable)?(\.exe)?$' "GUI"
  done < <(find "$ROOT_DIR/bin/gui" -mindepth 2 -maxdepth 2 \( -type f -o -type d \) | sort)
fi

exit "$status"
