#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Gosec security scanning script with selective rule filtering

.DESCRIPTION
    Runs gosec security scanner with optional rule and severity filtering

.PARAMETER Rule
    Scan for specific gosec rule (e.g., G304, G104). Can be specified multiple times.

.PARAMETER Format
    Output format: text, json, csv, junit-xml, html, yaml (default: text)

.PARAMETER Severity
    Filter by severity: low, medium, high. Can be specified multiple times.

.EXAMPLE
    .\gosec.ps1
    Scan all rules

.EXAMPLE
    .\gosec.ps1 -Rule G304
    Scan only G304

.EXAMPLE
    .\gosec.ps1 -Rule G304,G301
    Scan G304 and G301

.EXAMPLE
    .\gosec.ps1 -Format json
    Output as JSON

.EXAMPLE
    .\gosec.ps1 -Severity high
    Only high severity issues
#>

[CmdletBinding()]
param(
    [Parameter(ValueFromRemainingArguments)]
    [string[]]$Rule,
    
    [ValidateSet('text', 'json', 'csv', 'junit-xml', 'html', 'yaml')]
    [string]$Format = 'text',
    
    [ValidateSet('low', 'medium', 'high')]
    [string[]]$Severity
)

$ErrorActionPreference = "Continue"

# Get absolute paths
$THIS_DIR = $PSScriptRoot
$REPO_ROOT = Resolve-Path (Join-Path $THIS_DIR "..\..\..") | Select-Object -ExpandProperty Path
$ORIGINAL_DIR = Get-Location

# Ensure we return to original directory on exit
try {
    # Check if gosec is installed
    if (-not (Get-Command gosec -ErrorAction SilentlyContinue)) {
        Write-Host "Error: gosec is not installed." -ForegroundColor Red
        Write-Host "Install it with: $REPO_ROOT\scripts\security\tools\install-gosec.ps1" -ForegroundColor Yellow
        exit 1
    }

    # Build gosec command arguments
    $GosecArgs = @("-fmt=$Format")

    # Add rule filters if specified
    if ($Rule -and $Rule.Count -gt 0) {
        $IncludeRules = $Rule -join ','
        $GosecArgs += "-include=$IncludeRules"
        Write-Host "Scanning for rules: $IncludeRules" -ForegroundColor Blue
    } else {
        Write-Host "Scanning for all security issues" -ForegroundColor Blue
    }

    # Add severity filters if specified
    if ($Severity -and $Severity.Count -gt 0) {
        $SeverityFilter = $Severity -join ','
        $GosecArgs += "-severity=$SeverityFilter"
        Write-Host "Filtering by severity: $SeverityFilter" -ForegroundColor Blue
    }

    # Add target
    $GosecArgs += ".\..."

    # Show command being run
    Write-Host "`nRunning: gosec $($GosecArgs -join ' ')" -ForegroundColor Yellow
    Write-Host "Repository: $REPO_ROOT" -ForegroundColor Cyan
    Write-Host ""

    # Change to repo root and run gosec
    Set-Location $REPO_ROOT
    & gosec @GosecArgs
    $ExitCode = $LASTEXITCODE

    # Summary
    if ($ExitCode -eq 0) {
        Write-Host "Security scan completed successfully" -ForegroundColor Green
    } else {
        Write-Host "Security scan found issues (exit code: $ExitCode)" -ForegroundColor Yellow
    }

    exit $ExitCode
} finally {
    # Always return to original directory
    Set-Location $ORIGINAL_DIR
}
