#!/usr/bin/env bash
# Install gosec - Go security analyzer

set -e

echo "Installing gosec"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go first."
    exit 1
fi

# Install gosec
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Verify installation
if command -v gosec &> /dev/null; then
    echo "gosec installed successfully!"
    gosec -version
else
    echo "gosec installation failed"
    exit 1
fi
