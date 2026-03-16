<#
    .SYNOPSIS
    Run Go tests for syst.

    .DESCRIPTION
    Runs Go tests with optional verbose output and coverage reporting.

    .PARAMETER Verbose
    Enable verbose test output (-v).

    .PARAMETER Coverage
    Generate a coverage report.

    .PARAMETER CoverageOutput
    Path for the coverage output file.

    .PARAMETER Package
    Package pattern to test.

    .EXAMPLE
    .\run-tests.ps1
    .\run-tests.ps1 -Verbose
    .\run-tests.ps1 -Coverage
    .\run-tests.ps1 -Verbose -Coverage -Package "./internal/services/..."
#>
Param(
    [Parameter(Mandatory = $false, HelpMessage = "Enable verbose test output.")]
    [switch]$Verbose,
    [Parameter(Mandatory = $false, HelpMessage = "Generate a coverage report.")]
    [switch]$Coverage,
    [Parameter(Mandatory = $false, HelpMessage = "Path for the coverage output file.")]
    [string]$CoverageOutput = "coverage.out",
    [Parameter(Mandatory = $false, HelpMessage = "Package pattern to test.")]
    [string]$Package = "./..."
)

## Check Go is installed
try {
    $null = Get-Command go -ErrorAction Stop
} catch {
    Write-Error "Go is not installed."
    exit 1
}

## Build test command arguments
$TestArgs = @("-count=1")

if ($Verbose) {
    $TestArgs += "-v"
}

if ($Coverage) {
    $TestArgs += "-coverprofile=$CoverageOutput"
}

$TestArgs += $Package

Write-Host "==> Running Go tests..."
Write-Host "    Package: $Package"
Write-Host "    Verbose: $Verbose"
Write-Host "    Coverage: $Coverage"
Write-Host ""

go test @TestArgs
$TestExitCode = $LASTEXITCODE

if ($Coverage -and (Test-Path $CoverageOutput)) {
    Write-Host ""
    Write-Host "==> Coverage summary:"
    go tool cover -func="$CoverageOutput" | Select-Object -Last 1
    Write-Host ""
    Write-Host "    Full report: $CoverageOutput"
    Write-Host "    View in browser: go tool cover -html=$CoverageOutput"
}

exit $TestExitCode
