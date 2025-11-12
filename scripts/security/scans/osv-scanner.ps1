#!/usr/bin/env pwsh
<#
.SYNOPSIS
    OSV-Scanner dependency vulnerability scanning script

.DESCRIPTION
    Runs osv-scanner to scan for vulnerabilities in dependencies

.PARAMETER Format
    Output format: table (default), json, or markdown

.PARAMETER Lockfile
    Scan specific lockfile instead of directory

.EXAMPLE
    .\osv-scanner.ps1
    Scan repository with table output

.EXAMPLE
    .\osv-scanner.ps1 -Format json
    Scan with JSON output

.EXAMPLE
    .\osv-scanner.ps1 -Lockfile go.sum
    Scan specific lockfile
#>

[CmdletBinding()]
param(
    [ValidateSet('table', 'json', 'markdown')]
    [string]$Format = 'table',
    
    [string]$Lockfile
)

$ErrorActionPreference = "Continue"

# Get absolute paths
$THIS_DIR = $PSScriptRoot
$REPO_ROOT = Resolve-Path (Join-Path $THIS_DIR "..\..\..") | Select-Object -ExpandProperty Path
$ORIGINAL_DIR = Get-Location

# Ensure we return to original directory on exit
try {
    # Check if osv-scanner is installed
    if (-not (Get-Command osv-scanner -ErrorAction SilentlyContinue)) {
        Write-Host "Error: osv-scanner is not installed." -ForegroundColor Red
        Write-Host "Install it with: $REPO_ROOT\scripts\security\tools\install-osv-scanner.ps1" -ForegroundColor Yellow
        exit 1
    }

    # Build osv-scanner command arguments
    $OsvScannerArgs = @("--format=$Format")

    if ($Lockfile) {
        $OsvScannerArgs += "--lockfile=$Lockfile"
        $Target = $Lockfile
    } else {
        $OsvScannerArgs += "."
        $Target = "repository"
    }

    # Show command being run
    Write-Host "Running osv-scanner" -ForegroundColor Cyan
    Write-Host "Format: $Format" -ForegroundColor Blue
    Write-Host "Target: $Target" -ForegroundColor Blue
    Write-Host "Repository: $REPO_ROOT" -ForegroundColor Cyan
    Write-Host "Command: osv-scanner $($OsvScannerArgs -join ' ')" -ForegroundColor Yellow
    Write-Host ""

    # Change to repo root and run osv-scanner
    Set-Location $REPO_ROOT
    & osv-scanner @OsvScannerArgs
    $ExitCode = $LASTEXITCODE

    # Summary
    Write-Host "`n═══════════════════════════════════════" -ForegroundColor Blue
    if ($ExitCode -eq 0) {
        Write-Host "No vulnerabilities found" -ForegroundColor Green
    } else {
        Write-Host "Vulnerabilities detected (exit code: $ExitCode)" -ForegroundColor Yellow
    }
    Write-Host "═══════════════════════════════════════" -ForegroundColor Blue

    exit $ExitCode
} finally {
    # Always return to original directory
    Set-Location $ORIGINAL_DIR
}
