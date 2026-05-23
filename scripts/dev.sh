#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DB_PATH="${DB_PATH:-$ROOT_DIR/bill_of_fare.db}"
MENU_PATH="${MENU_PATH:-$ROOT_DIR/seed/menu.yaml}"
PORT="${PORT:-8080}"
BIN_PATH="${TMPDIR:-/tmp}/bill-of-fare-dev-server"

cd "$ROOT_DIR"

if [[ ! -f "$DB_PATH" ]]; then
  echo "No database found at $DB_PATH; seeding from $MENU_PATH"
  go run ./cmd/seed -db "$DB_PATH" -menu "$MENU_PATH"
fi

server_pid=""

cleanup() {
  if [[ -n "$server_pid" ]] && kill -0 "$server_pid" 2>/dev/null; then
    kill "$server_pid" 2>/dev/null || true
    wait "$server_pid" 2>/dev/null || true
  fi
}

snapshot() {
  find cmd internal web database seed -type f \
    \( -name '*.go' -o -name '*.html' -o -name '*.css' -o -name '*.sql' -o -name '*.yaml' \) \
    -print0 | sort -z | xargs -0 shasum 2>/dev/null || true
}

start_server() {
  cleanup
  go build -o "$BIN_PATH" ./cmd/server
  echo "Starting Bill of Fare on http://localhost:$PORT"
  DB_PATH="$DB_PATH" PORT="$PORT" "$BIN_PATH" &
  server_pid="$!"
}

trap cleanup EXIT INT TERM

last_snapshot="$(snapshot)"
start_server

while true; do
  sleep 1
  current_snapshot="$(snapshot)"
  if [[ "$current_snapshot" != "$last_snapshot" ]]; then
    echo "Change detected; restarting server"
    last_snapshot="$current_snapshot"
    start_server
  fi
done
