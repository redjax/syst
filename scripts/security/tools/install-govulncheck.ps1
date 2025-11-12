#!/usr/bin/env pwsh
# Install govulncheck - Go vulnerability scanner

$ErrorActionPreference = "Stop"

Write-Host "Installing govulncheck" -ForegroundColor Cyan

# Check if Go is installed
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "Error: Go is not installed. Please install Go first." -ForegroundColor Red
    exit 1
}

# Install govulncheck
Write-Host "Running: go install golang.org/x/vuln/cmd/govulncheck@latest" -ForegroundColor Yellow
go install golang.org/x/vuln/cmd/govulncheck@latest

# Verify installation
if (Get-Command govulncheck -ErrorAction SilentlyContinue) {
    Write-Host "govulncheck installed successfully!" -ForegroundColor Green
    govulncheck -version
} else {
    Write-Host "govulncheck installation failed" -ForegroundColor Red
    exit 1
}
