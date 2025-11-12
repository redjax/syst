#!/usr/bin/env bash
# Install govulncheck - Go vulnerability scanner

set -e

echo "Installing govulncheck"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go first."
    exit 1
fi

# Install govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest

# Verify installation
if command -v govulncheck &> /dev/null; then
    echo "govulncheck installed successfully!"
    govulncheck -version
else
    echo "govulncheck installation failed"
    exit 1
fi
