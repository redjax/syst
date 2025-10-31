#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

if ! command -v go &>/dev/null; then
    echo "[ERROR] Go is not installed."
    exit 1
fi

echo "[ Go Project Vulnerability Scan ]"
echo " -------------------------------"
echo ""

## Check if govulncheck is installed
if ! command -v govulncheck &> /dev/null; then
    echo "govulncheck not found. Installing"
    
    go install golang.org/x/vuln/cmd/govulncheck@latest
    echo ""
fi

## Check if gosec is installed
if ! command -v gosec &> /dev/null; then
    echo "gosec not found. Installing"
    go install github.com/securego/gosec/v2/cmd/gosec@latest
    echo ""
fi

## Default inputs
REPORTS_DIR="$PROJECT_ROOT/vulnerability-reports"

while [[ $# -gt 0 ]]; do
    case "$1" in
        -o|--output-dir)
            if [[ -z "$2" ]]; then
                echo "[ERROR] --output-dir provided but no directory path given" >&2

                exit 1
            fi

            REPORTS_DIR="$2"
            shift 2
            ;;
        *)
            echo "Unknown argument: $1" >&2
            exit 1
            ;;
    esac
done

if [[ "$REPORTS_DIR" == "" ]]; then
    REPORTS_DIR="$PROJECT_ROOT/vulnerability-reports"
fi

## Add timestamp to reports directory
REPORTS_DIR="$REPORTS_DIR/$(date +%Y-%m-%d-%H-%M-%S)"

if [[ ! -d "$REPORTS_DIR" ]]; then
    mkdir -p "$REPORTS_DIR"
fi

echo ""
echo "--[ Run govulncheck"
echo "   ----------------"
echo ""

## Run govulncheck with human-readable output
govulncheck ./... 2>&1 | tee "$REPORTS_DIR/govulncheck-report.txt"
GOVULNCHECK_EXIT=$?

echo ""
echo "(govulncheck) Generating SARIF report"
echo "-------------------------------------"
echo ""

## Generate SARIF format
govulncheck -format sarif ./... > "$REPORTS_DIR/govulncheck-report.sarif" 2>&1 || true

echo ""
echo "--[ Run gosec"
echo "   ----------"
echo ""

## Run gosec
gosec -fmt=json -out="$REPORTS_DIR/gosec-report.json" ./... || true
gosec -fmt=text ./... 2>&1 | tee "$REPORTS_DIR/gosec-report.txt" || true

echo ""
echo "-- [ Finished scanning project"

echo ""
echo "[ Summary ]"
echo " ---------"
echo ""
echo "Reports saved to: $REPORTS_DIR"
echo ""
echo "Files generated:"
echo "  - govulncheck-report.txt   (govulncheck results)"
echo "  - govulncheck-report.sarif (SARIF format for GitHub)"
echo "  - gosec-report.json        (gosec JSON format)"
echo "  - gosec-report.txt         (gosec results)"
echo ""

if [ $GOVULNCHECK_EXIT -ne 0 ]; then
    echo "[WARNING] Vulnerabilities found"
    echo ""
    echo "View the reports above for details."
    echo "Main findings are in: $REPORTS_DIR/govulncheck-report.txt"
    exit $GOVULNCHECK_EXIT
else
    echo "No vulnerabilities found"
fi
