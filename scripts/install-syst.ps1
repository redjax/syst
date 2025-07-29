[CmdletBinding()]
Param(
    [switch] $Auto
)

## Stop immediately on error
$ErrorActionPreference = 'Stop'

## Determine install directory (default to $env:USERPROFILE\bin)
$installPath = Join-Path $env:USERPROFILE 'bin'
if ( -not ( Test-Path $installPath ) ) {
    New-Item -ItemType Directory -Path $installPath -Force | Out-Null
}

## Check if 'syst.exe' already exists in PATH or install path
$existingSyst = Get-Command syst -ErrorAction SilentlyContinue
if ($existingSyst) {
    if (-not $Auto) {
        $confirm = Read-Host "syst is already installed at $($existingSyst.Path). Download and install again? (y/N)"
        if ( -not ( $confirm -in @('y', 'Y', 'yes', 'Yes', 'YES' ) ) ) {
            Read-Prompt "Cancelling installation, press a key to exit ..."
            exit 0
        }
    }
}

## Get latest release tag from GitHub
try {
    $releaseApi = 'https://api.github.com/repos/redjax/syst/releases/latest'
    $release = Invoke-RestMethod -Uri $releaseApi -UseBasicParsing
} catch {
    Write-Error "Failed to fetch latest release info: $($_.Exception.Message)"
    throw $_.Exception
}

$version = $release.tag_name.TrimStart('v')
Write-Host "Installing syst version $version"

## Detect CPU architecture
try {
    $archCode = (Get-CimInstance Win32_Processor | Select-Object -First 1).Architecture
} catch {
    ## Fallback to environment variable string if CIM fails
    $envArch = $env:PROCESSOR_ARCHITECTURE
    if ($envArch -match '^(AMD64|x86_64)$') {
        $archNorm = 'amd64'
    } elseif ($envArch -match '^ARM64$') {
        $archNorm = 'arm64'
    } else {
        Write-Error "Unsupported architecture: $envArch"
        throw $_.Exception
    }
}
## If CPU was not set by CIM, try to normalize
if ( -not $archNorm ) {
    switch ($archCode) {
        ## x64
        9 { $archNorm = 'amd64' }
        ## ARM64
        12 { $archNorm = 'arm64' }
        default {
            Write-Error "Unsupported architecture code: $archCode"
            throw $_.Exception
        }
    }
}

## Build asset file name and download URL
$fileName = "syst-windows-$archNorm-$version.zip"
$downloadUrl = "https://github.com/redjax/syst/releases/download/$($release.tag_name)/$fileName"

## Create temp folder
$tempDir = Join-Path -Path ([System.IO.Path]::GetTempPath()) -ChildPath ("syst_install_" + [Guid]::NewGuid())
Write-Debug "Using temp dir: $($tempDir)"
try {
    New-Item -ItemType Directory -Path $tempDir | Out-Null
} catch {
    Write-Error "Failed to create temp dir: $($_.Exception.Message)"
    throw $_.Exception
}

$zipPath = Join-Path $tempDir $fileName

Write-Host "Downloading $fileName from GitHub..."
try {
    Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath -UseBasicParsing
} catch {
    Write-Error "Failed to download asset: $($_.Exception.Message)"
    Remove-Item -Recurse -Force $tempDir
    
    throw $_.Exception
}

Write-Host "Extracting package..."
try {
    Expand-Archive -Path $zipPath -DestinationPath $tempDir -Force
} catch {
    Write-Error "Failed to extract archive: $_"
    Remove-Item -Recurse -Force $tempDir
    
    throw $_.Exception
}

## Expect binary named "syst.exe" in root of zip
$systExePath = Join-Path $tempDir 'syst.exe'
if ( -not ( Test-Path $systExePath ) ) {
    Write-Error "Extracted package missing expected syst.exe"
    Remove-Item -Recurse -Force $tempDir
    
    return
}

## Copy executable to install path
$destExePath = Join-Path $installPath 'syst.exe'
try {
    Copy-Item -Path $systExePath -Destination $destExePath -Force
} catch {
    Write-Error "Failed to install syst.exe: $($_.Exception.Message)"
    Remove-Item -Recurse -Force $tempDir
    
    throw $_.Exception
}

Remove-Item -Recurse -Force $tempDir

Write-Host "syst installed successfully to $destExePath"

## Check if $installPath is already in user's PATH
$userPath = [Environment]::GetEnvironmentVariable('PATH', [EnvironmentVariableTarget]::User)
if ( -not ( $userPath -split ';' | Where-Object { $_ -eq $installPath } ) ) {
    Write-Host "`nNOTE: '$installPath' is not in your user PATH environment variable."
    Write-Host "Add it by running this once in PowerShell:"
    Write-Host "[Environment]::SetEnvironmentVariable('PATH',`"$userPath;$installPath`",'User')"
    Write-Host "Then restart your PowerShell session for changes to take effect."
}
