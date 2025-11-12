#!/usr/bin/env pwsh
# Install gosec - Go security analyzer

$ErrorActionPreference = "Stop"

Write-Host "Installing gosec" -ForegroundColor Cyan

# Check if Go is installed
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "Error: Go is not installed. Please install Go first." -ForegroundColor Red
    exit 1
}

# Install gosec
Write-Host "Running: go install github.com/securego/gosec/v2/cmd/gosec@latest" -ForegroundColor Yellow
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Verify installation
if (Get-Command gosec -ErrorAction SilentlyContinue) {
    Write-Host "gosec installed successfully!" -ForegroundColor Green
    gosec -version
} else {
    Write-Host "gosec installation failed" -ForegroundColor Red
    exit 1
}
