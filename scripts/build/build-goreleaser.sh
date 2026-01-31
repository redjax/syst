#!/usr/bin/env bash
set -uo pipefail

if ! command -v goreleaser &>/dev/null; then
  echo "[ERROR] GoReleaser is not installed."
  exit 1
fi

goreleaser release --snapshot --clean
