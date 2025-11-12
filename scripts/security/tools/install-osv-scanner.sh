#!/usr/bin/env bash
# Install osv-scanner - Dependency vulnerability scanner

set -e

echo "Installing osv-scanner"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go first."
    exit 1
fi

# Install osv-scanner
go install github.com/google/osv-scanner/cmd/osv-scanner@latest

# Verify installation
if command -v osv-scanner &> /dev/null; then
    echo "osv-scanner installed successfully!"
    osv-scanner --version
else
    echo "osv-scanner installation failed"
    exit 1
fi
