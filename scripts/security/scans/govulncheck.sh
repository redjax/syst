#!/usr/bin/env bash
# Govulncheck vulnerability scanning script
# Usage: ./govulncheck.sh [OPTIONS]
#   --mode MODE     Scan mode: source (default), binary, or convert
#   --format FORMAT Output format: text (default), json, sarif
#   --help          Show this help message
#
# Examples:
#   ./govulncheck.sh                  # Scan source code
#   ./govulncheck.sh --format json    # JSON output
#   ./govulncheck.sh --mode binary    # Scan binary

set -e

# Get absolute paths
THIS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${THIS_DIR}/../../.." && pwd)"
ORIGINAL_DIR="$(pwd)"

# Exit trap to return to original directory
trap "cd '${ORIGINAL_DIR}'" EXIT

# Default values
MODE="source"
FORMAT="text"

# Help message
show_help() {
    cat << EOF
Govulncheck vulnerability scanning script

Usage: $(basename "$0") [OPTIONS]

Options:
  --mode MODE      Scan mode: source (default), binary, convert
  --format FORMAT  Output format: text (default), json, sarif
  --help          Show this help message

Examples:
  $(basename "$0")                  # Scan source code
  $(basename "$0") --format json    # JSON output
  $(basename "$0") --mode binary    # Scan binary

EOF
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --mode)
            MODE="$2"
            shift 2
            ;;
        --format)
            FORMAT="$2"
            shift 2
            ;;
        --help)
            show_help
            exit 0
            ;;
        *)
            echo "Error: Unknown option $1"
            show_help
            exit 1
            ;;
    esac
done

# Check if govulncheck is installed
if ! command -v govulncheck &> /dev/null; then
    echo "Error: govulncheck is not installed."
    echo "Install it with: ${REPO_ROOT}/scripts/security/tools/install-govulncheck.sh"
    exit 1
fi

# Build govulncheck command
CMD="govulncheck"

if [ "$MODE" != "source" ]; then
    CMD="${CMD} -mode=${MODE}"
fi

if [ "$FORMAT" != "text" ]; then
    CMD="${CMD} -format=${FORMAT}"
fi

CMD="${CMD} ./..."

# Show command being run
echo "Running govulncheck"
echo "Mode: ${MODE}"
echo "Format: ${FORMAT}"
echo "Repository: ${REPO_ROOT}"
echo "Command: ${CMD}"
echo ""

# Change to repo root and run govulncheck
cd "$REPO_ROOT"
eval "$CMD"
EXIT_CODE=$?

# Summary
echo ""
if [ $EXIT_CODE -eq 0 ]; then
    echo "No vulnerabilities found"
else
    echo "Vulnerabilities detected (exit code: $EXIT_CODE)"
fi

exit $EXIT_CODE
