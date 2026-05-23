#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="$ROOT_DIR/dist/windows"
GOARCH="${GOARCH:-amd64}"
OUT="$OUT_DIR/BillOfFare.exe"

cd "$ROOT_DIR"
mkdir -p "$OUT_DIR"

GOOS=windows GOARCH="$GOARCH" go build -trimpath -ldflags="-s -w" -o "$OUT" ./cmd/desktop

cat > "$OUT_DIR/README.txt" <<'README'
Bill of Fare for Windows

Double-click BillOfFare.exe to start the local POS.
It will open the app in your default browser automatically.

Data is stored per Windows user in:
%AppData%\Bill of Fare\bill_of_fare.db

To stop the app, close the BillOfFare.exe console window.
README

echo "Built $OUT"
