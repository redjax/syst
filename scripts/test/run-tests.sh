#!/usr/bin/env bash
set -euo pipefail

## Default param values
Verbose=false
Coverage=false
CoverageOutput="coverage.out"
Package="./..."

function print_help {
  echo "-- | Run Go tests | --"
  echo ""
  echo "[ Flags ]"
  echo "  --help: Print this help menu"
  echo ""
  echo "  --verbose: Enable verbose test output (-v)"
  echo "  --coverage: Generate coverage report"
  echo "  --coverage-output (default: coverage.out): Coverage output file"
  echo "  --package (default: ./...): Package pattern to test"
}

if ! command -v go >/dev/null 2>&1; then
  echo "[ERROR] Go is not installed."
  exit 1
fi

## Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --help)
      print_help
      exit 0
      ;;
    --verbose)
      Verbose=true
      shift
      ;;
    --coverage)
      Coverage=true
      shift
      ;;
    --coverage-output)
      CoverageOutput="$2"
      shift 2
      ;;
    --package)
      Package="$2"
      shift 2
      ;;
    *)
      echo "[ERROR] Unknown option: $1"
      print_help
      exit 1
      ;;
  esac
done

## Build test command arguments
TestArgs=("-count=1")

if [ "$Verbose" = true ]; then
  TestArgs+=("-v")
fi

if [ "$Coverage" = true ]; then
  TestArgs+=("-coverprofile=${CoverageOutput}")
fi

TestArgs+=("${Package}")

echo "==> Running Go tests..."
echo "    Package: ${Package}"
echo "    Verbose: ${Verbose}"
echo "    Coverage: ${Coverage}"
echo ""

go test "${TestArgs[@]}"
TestExitCode=$?

if [ "$Coverage" = true ] && [ -f "$CoverageOutput" ]; then
  echo ""
  echo "==> Coverage summary:"
  go tool cover -func="$CoverageOutput" | tail -1
  echo ""
  echo "    Full report: ${CoverageOutput}"
  echo "    View in browser: go tool cover -html=${CoverageOutput}"
fi

exit $TestExitCode
