#!/bin/bash
# Local vulnerability scanning script
# Mimics the GitHub Actions workflow for local testing

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

echo "================================================"
echo "Running Vulnerability Scan"
echo "================================================"
echo ""

# Check if govulncheck is installed
if ! command -v govulncheck &> /dev/null; then
    echo "govulncheck not found. Installing..."
    go install golang.org/x/vuln/cmd/govulncheck@latest
    echo ""
fi

# Check if gosec is installed
if ! command -v gosec &> /dev/null; then
    echo "gosec not found. Installing..."
    go install github.com/securego/gosec/v2/cmd/gosec@latest
    echo ""
fi

# Create reports directory
REPORTS_DIR="$PROJECT_ROOT/vulnerability-reports"
mkdir -p "$REPORTS_DIR"

echo "================================================"
echo "1. Running govulncheck..."
echo "================================================"
echo ""

# Run govulncheck with human-readable output
govulncheck ./... 2>&1 | tee "$REPORTS_DIR/govulncheck-report.txt"
GOVULNCHECK_EXIT=$?

echo ""
echo "================================================"
echo "2. Generating SARIF report..."
echo "================================================"
echo ""

# Generate SARIF format
govulncheck -format sarif ./... > "$REPORTS_DIR/govulncheck-report.sarif" 2>&1 || true

echo ""
echo "================================================"
echo "3. Running gosec..."
echo "================================================"
echo ""

# Run gosec
gosec -fmt=json -out="$REPORTS_DIR/gosec-report.json" ./... || true
gosec -fmt=text ./... 2>&1 | tee "$REPORTS_DIR/gosec-report.txt" || true

echo ""
echo "================================================"
echo "Scan Complete!"
echo "================================================"
echo ""
echo "Reports saved to: $REPORTS_DIR"
echo ""
echo "Files generated:"
echo "  - govulncheck-report.txt  (human-readable govulncheck results)"
echo "  - govulncheck-report.sarif (SARIF format for GitHub)"
echo "  - gosec-report.json       (gosec JSON format)"
echo "  - gosec-report.txt        (human-readable gosec results)"
echo ""

if [ $GOVULNCHECK_EXIT -ne 0 ]; then
    echo "⚠️  WARNING: Vulnerabilities found!"
    echo ""
    echo "View the reports above for details."
    echo "Main findings are in: $REPORTS_DIR/govulncheck-report.txt"
    exit $GOVULNCHECK_EXIT
else
    echo "✅ No vulnerabilities found!"
fi
