#!/usr/bin/env bash
# OSV-Scanner dependency vulnerability scanning script
# Usage: ./osv-scanner.sh [OPTIONS]
#   --format FORMAT Output format: table (default), json, markdown
#   --lockfile PATH Scan specific lockfile instead of directory
#   --help          Show this help message
#
# Examples:
#   ./osv-scanner.sh                    # Scan repository
#   ./osv-scanner.sh --format json      # JSON output
#   ./osv-scanner.sh --lockfile go.sum  # Scan specific lockfile

set -e

# Get absolute paths
THIS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${THIS_DIR}/../../.." && pwd)"
ORIGINAL_DIR="$(pwd)"

# Exit trap to return to original directory
trap "cd '${ORIGINAL_DIR}'" EXIT

# Default values
FORMAT="table"
LOCKFILE=""

# Help message
show_help() {
    cat << EOF
OSV-Scanner dependency vulnerability scanning script

Usage: $(basename "$0") [OPTIONS]

Options:
  --format FORMAT   Output format: table (default), json, markdown
  --lockfile PATH   Scan specific lockfile instead of directory
  --help           Show this help message

Examples:
  $(basename "$0")                    # Scan repository
  $(basename "$0") --format json      # JSON output
  $(basename "$0") --lockfile go.sum  # Scan specific lockfile

EOF
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --format)
            FORMAT="$2"
            shift 2
            ;;
        --lockfile)
            LOCKFILE="$2"
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

# Check if osv-scanner is installed
if ! command -v osv-scanner &> /dev/null; then
    echo "Error: osv-scanner is not installed."
    echo "Install it with: ${REPO_ROOT}/scripts/security/tools/install-osv-scanner.sh"
    exit 1
fi

# Build osv-scanner command
CMD="osv-scanner --format=${FORMAT}"

if [ -n "$LOCKFILE" ]; then
    CMD="${CMD} --lockfile=${LOCKFILE}"
    TARGET="$LOCKFILE"
else
    CMD="${CMD} ."
    TARGET="repository"
fi

# Show command being run
echo "Running osv-scanner"
echo "Format: ${FORMAT}"
echo "Target: ${TARGET}"
echo "Repository: ${REPO_ROOT}"
echo "Command: ${CMD}"
echo ""

# Change to repo root and run osv-scanner
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
