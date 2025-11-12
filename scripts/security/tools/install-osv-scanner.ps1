#!/usr/bin/env pwsh
# Install osv-scanner - Dependency vulnerability scanner

$ErrorActionPreference = "Stop"

Write-Host "Installing osv-scanner" -ForegroundColor Cyan

# Check if Go is installed
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "Error: Go is not installed. Please install Go first." -ForegroundColor Red
    exit 1
}

# Install osv-scanner
Write-Host "Running: go install github.com/google/osv-scanner/cmd/osv-scanner@latest" -ForegroundColor Yellow
go install github.com/google/osv-scanner/cmd/osv-scanner@latest

# Verify installation
if (Get-Command osv-scanner -ErrorAction SilentlyContinue) {
    Write-Host "osv-scanner installed successfully!" -ForegroundColor Green
    osv-scanner --version
} else {
    Write-Host "osv-scanner installation failed" -ForegroundColor Red
    exit 1
}
