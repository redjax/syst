#!/usr/bin/env bash
# Gosec security scanning script with selective rule filtering
# Usage: ./gosec.sh [OPTIONS]
#   --rule RULE     Scan for specific rule (e.g., G304, G104). Can be specified multiple times.
#   --format FORMAT Output format: text, json, csv, junit-xml, html, yaml (default: text)
#   --severity SEV  Filter by severity: low, medium, high (can be specified multiple times)
#   --help          Show this help message
#
# Examples:
#   ./gosec.sh                           # Scan all rules
#   ./gosec.sh --rule G304               # Scan only G304
#   ./gosec.sh --rule G304 --rule G301   # Scan G304 and G301
#   ./gosec.sh --format json             # Output as JSON
#   ./gosec.sh --severity high           # Only high severity issues

set -e

# Get absolute paths
THIS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${THIS_DIR}/../../.." && pwd)"
ORIGINAL_DIR="$(pwd)"

# Exit trap to return to original directory
trap "cd '${ORIGINAL_DIR}'" EXIT

# Default values
RULES=()
FORMAT="text"
SEVERITIES=()

# Help message
show_help() {
    cat << EOF
Gosec security scanning script with selective rule filtering

Usage: $(basename "$0") [OPTIONS]

Options:
  --rule RULE       Scan for specific gosec rule (e.g., G304, G104)
                    Can be specified multiple times for multiple rules
  --format FORMAT   Output format: text, json, csv, junit-xml, html, yaml
                    Default: text
  --severity SEV    Filter by severity: low, medium, high
                    Can be specified multiple times
  --help           Show this help message

Examples:
  $(basename "$0")                           # Scan all rules
  $(basename "$0") --rule G304               # Scan only G304
  $(basename "$0") --rule G304 --rule G301   # Scan G304 and G301
  $(basename "$0") --format json             # Output as JSON
  $(basename "$0") --severity high           # Only high severity issues

EOF
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --rule)
            RULES+=("$2")
            shift 2
            ;;
        --format)
            FORMAT="$2"
            shift 2
            ;;
        --severity)
            SEVERITIES+=("$2")
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

# Check if gosec is installed
if ! command -v gosec &> /dev/null; then
    echo "Error: gosec is not installed."
    echo "Install it with: ${REPO_ROOT}/scripts/security/tools/install-gosec.sh"
    exit 1
fi

# Build gosec command
CMD="gosec -fmt=${FORMAT}"

# Add rule filters if specified
if [ ${#RULES[@]} -gt 0 ]; then
    INCLUDE_RULES=$(IFS=,; echo "${RULES[*]}")
    CMD="${CMD} -include=${INCLUDE_RULES}"
    echo "Scanning for rules: ${INCLUDE_RULES}"
else
    echo "Scanning for all security issues"
fi

# Add severity filters if specified
if [ ${#SEVERITIES[@]} -gt 0 ]; then
    SEVERITY_FILTER=$(IFS=,; echo "${SEVERITIES[*]}")
    CMD="${CMD} -severity=${SEVERITY_FILTER}"
    echo "Filtering by severity: ${SEVERITY_FILTER}"
fi

# Add target
CMD="${CMD} ./..."

# Show command being run
echo "Running: ${CMD}"
echo "Repository: ${REPO_ROOT}"
echo ""

# Change to repo root and run gosec
cd "$REPO_ROOT"
eval "$CMD"
EXIT_CODE=$?

# Summary
echo ""
if [ $EXIT_CODE -eq 0 ]; then
    echo "Security scan completed successfully"
else
    echo "Security scan found issues (exit code: $EXIT_CODE)"
fi

exit $EXIT_CODE
