#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Govulncheck vulnerability scanning script

.DESCRIPTION
    Runs govulncheck to scan for known vulnerabilities in Go code and dependencies

.PARAMETER Mode
    Scan mode: source (default), binary, or convert

.PARAMETER Format
    Output format: text (default), json, or sarif

.EXAMPLE
    .\govulncheck.ps1
    Scan source code with text output

.EXAMPLE
    .\govulncheck.ps1 -Format json
    Scan with JSON output

.EXAMPLE
    .\govulncheck.ps1 -Mode binary
    Scan binary instead of source
#>

[CmdletBinding()]
param(
    [ValidateSet('source', 'binary', 'convert')]
    [string]$Mode = 'source',
    
    [ValidateSet('text', 'json', 'sarif')]
    [string]$Format = 'text'
)

$ErrorActionPreference = "Continue"

# Get absolute paths
$THIS_DIR = $PSScriptRoot
$REPO_ROOT = Resolve-Path (Join-Path $THIS_DIR "..\..\..") | Select-Object -ExpandProperty Path
$ORIGINAL_DIR = Get-Location

# Ensure we return to original directory on exit
try {
    # Check if govulncheck is installed
    if (-not (Get-Command govulncheck -ErrorAction SilentlyContinue)) {
        Write-Host "Error: govulncheck is not installed." -ForegroundColor Red
        Write-Host "Install it with: $REPO_ROOT\scripts\security\tools\install-govulncheck.ps1" -ForegroundColor Yellow
        exit 1
    }

    # Build govulncheck command arguments
    $GovulncheckArgs = @()

    if ($Mode -ne 'source') {
        $GovulncheckArgs += "-mode=$Mode"
    }

    if ($Format -ne 'text') {
        $GovulncheckArgs += "-format=$Format"
    }

    $GovulncheckArgs += ".\..."

    # Show command being run
    Write-Host "Running govulncheck" -ForegroundColor Cyan
    Write-Host "Mode: $Mode" -ForegroundColor Blue
    Write-Host "Format: $Format" -ForegroundColor Blue
    Write-Host "Repository: $REPO_ROOT" -ForegroundColor Cyan
    Write-Host "Command: govulncheck $($GovulncheckArgs -join ' ')" -ForegroundColor Yellow
    Write-Host ""

    # Change to repo root and run govulncheck
    Set-Location $REPO_ROOT
    & govulncheck @GovulncheckArgs
    $ExitCode = $LASTEXITCODE

    # Summary
    if ($ExitCode -eq 0) {
        Write-Host "No vulnerabilities found" -ForegroundColor Green
    } else {
        Write-Host "Vulnerabilities detected (exit code: $ExitCode)" -ForegroundColor Yellow
    }
    
    exit $ExitCode
} finally {
    # Always return to original directory
    Set-Location $ORIGINAL_DIR
}
