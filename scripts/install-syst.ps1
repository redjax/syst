<#
    .SYNOPSIS
    This script downloads and installs `syst` on Windows.

    .DESCRIPTION
    Downloads the latest release of `syst` from GitHub, and installs it to the
    current user's %LOCALAPPDATA%\syst directory. If `syst` is already installed,
    the user will be prompted to download and install again. The script will also
    append the install directory to the PATH environment variable.

    .PARAMETER Auto
    If specified, the script will install `syst` without prompting the user.

    .EXAMPLE
    & ([scriptblock]::Create((irm https://raw.githubusercontent.com/redjax/syst/refs/heads/main/scripts/install-syst.ps1))) -Auto
#>

[CmdletBinding()]
Param(
    [switch] $Auto
)

## Stop immediately on error
$ErrorActionPreference = 'Stop'

## Set install directory to $env:LOCALAPPDATA\syst
$InstallPath = Join-Path $env:LOCALAPPDATA 'syst'
if (-not (Test-Path $InstallPath)) {
    New-Item -ItemType Directory -Path $InstallPath -Force | Out-Null
}

## Check if 'syst.exe' already exists in PATH or install path
$ExistingSyst = Get-Command syst -ErrorAction SilentlyContinue
if ($ExistingSyst) {
    if (-not $Auto) {
        $Confirm = Read-Host "syst is already installed at $($ExistingSyst.Path). Download and install again? (y/N)"
        if (-not ( $Confirm -in @('y', 'Y', 'yes', 'Yes', 'YES' ) )) {
            Write-Host "Cancelling installation."
            exit 0
        }
    }
}

## Get latest release tag from GitHub
try {
    $ReleaseApi = 'https://api.github.com/repos/redjax/syst/releases/latest'
    $Release = Invoke-RestMethod -Uri $ReleaseApi -UseBasicParsing
} catch {
    Write-Error "Failed to fetch latest release info: $($_.Exception.Message)"
    throw $_.Exception
}

$Version = $Release.tag_name.TrimStart('v')
Write-Host "Installing syst version $Version"

## Detect CPU architecture
$ArchNorm = $null
try {
    $ArchCode = (Get-CimInstance Win32_Processor | Select-Object -First 1).Architecture
    
    # Convert ArchCode to normalized name
    switch ($ArchCode) {
        9 { $ArchNorm = 'amd64' }
        12 { $ArchNorm = 'arm64' }
        default {
            Write-Error "Unsupported architecture code: $ArchCode"
            throw "Unsupported architecture code: $ArchCode"
        }
    }
} catch {
    # Fallback to environment variable
    $EnvArch = $env:PROCESSOR_ARCHITECTURE
    if ($EnvArch -match '^(AMD64|x86_64)$') {
        $ArchNorm = 'amd64'
    } elseif ($EnvArch -match '^ARM64$') {
        $ArchNorm = 'arm64'
    } else {
        Write-Error "Unsupported architecture: $EnvArch"
        throw "Unsupported architecture: $EnvArch"
    }
}

if (-not $ArchNorm) {
    Write-Error "Failed to detect system architecture"
    throw "Failed to detect system architecture"
}

## Build asset file name and download URL
$FileName = "syst-windows-$ArchNorm-$Version.zip"
$DownloadUrl = "https://github.com/redjax/syst/releases/download/$($Release.tag_name)/$FileName"

## Create temp folder
$TempDir = Join-Path -Path ([System.IO.Path]::GetTempPath()) -ChildPath ("syst_install_" + [Guid]::NewGuid())
Write-Debug "Using temp dir: $($TempDir)"
try {
    New-Item -ItemType Directory -Path $TempDir | Out-Null
} catch {
    Write-Error "Failed to create temp dir: $($_.Exception.Message)"
    throw $_.Exception
}

$ZipPath = Join-Path $TempDir $FileName

Write-Host "Downloading $FileName from GitHub..."
try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $ZipPath -UseBasicParsing
} catch {
    Write-Error "Failed to download asset: $($_.Exception.Message)"
    Remove-Item -Recurse -Force $TempDir
    throw $_.Exception
}

Write-Host "Extracting package..."
try {
    Expand-Archive -Path $ZipPath -DestinationPath $TempDir -Force
} catch {
    Write-Error "Failed to extract archive: $_"
    Remove-Item -Recurse -Force $TempDir
    throw $_.Exception
}

## Expect binary named "syst.exe" in root of zip
$SystExePath = Join-Path $TempDir 'syst.exe'
if (-not (Test-Path $SystExePath)) {
    Write-Error "Extracted package missing expected syst.exe"
    Remove-Item -Recurse -Force $TempDir
    return
}

## Copy executable to install path
$DestExePath = Join-Path $InstallPath 'syst.exe'
try {
    Copy-Item -Path $SystExePath -Destination $DestExePath -Force
} catch {
    Write-Error "Failed to install syst.exe: $($_.Exception.Message)"
    Remove-Item -Recurse -Force $TempDir
    throw $_.Exception
}

Remove-Item -Recurse -Force $TempDir

Write-Host "`nsyst installed successfully to $DestExePath"

## Add to PATH if not present
$UserPath = [Environment]::GetEnvironmentVariable('PATH', [EnvironmentVariableTarget]::User)

if ( -not ( $UserPath -split ';' | Where-Object { $_ -eq $InstallPath } ) ) {
    try {
        [Environment]::SetEnvironmentVariable('PATH', "$UserPath;$InstallPath", 'User')

        Write-Host "Added '$InstallPath' to user PATH environment variable. Close and reopen your shell for changes to take effect."
    } catch {
        Write-Error "Failed to update PATH environment variable: $($_.Exception.Message)"
        
        Write-Warning @"
'$InstallPath' is not in your user PATH environment variable."

"Add it by running this once in PowerShell:"
    $> [Environment]::SetEnvironmentVariable('PATH', "`$UserPath;`$InstallPath", 'User')

Then close & re-open your shell for changes to take effect.
"@

        throw $_.Exception
    }
}
