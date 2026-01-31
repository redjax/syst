#!/usr/bin/env bash
set -uo pipefail

echo "Upgrading Go packages"

go get -u all

while true; do
  read -r -n 1 -p "Run go mod tidy? y/n: " choice
  echo ""

  case $choice in
  [Yy])
    go mod tidy
    break
    ;;
  [Nn])
    echo "Exiting."
    exit 1
    ;;
  *)
    echo "[ERROR] Invalid choice: $choice. Please use 'y' or 'n'."
    ;;
  esac
done
