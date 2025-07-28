param(
    [switch] $Auto
)

$ErrorActionPreference = 'Stop'

# Check for dependencies
if (-not (Get-Command Invoke-RestMethod -ErrorAction SilentlyContinue)) {
    Write-Error "Invoke-RestMethod is not available. Please run this in PowerShell 3.0 or higher."
    exit 1
}

if (-not (Get-Command Expand-Archive -ErrorAction SilentlyContinue)) {
    Write-Error "Expand-Archive is not available. Please ensure PowerShell 5.0 or higher."
    exit 1
}

# Determine install directory (default to $env:USERPROFILE\bin)
$installPath = Join-Path $env:USERPROFILE 'bin'
if (-not (Test-Path $installPath)) {
    New-Item -ItemType Directory -Path $installPath -Force | Out-Null
}

# Check if 'syst.exe' already exists in PATH or install path
$existingSyst = Get-Command syst -ErrorAction SilentlyContinue
if ($existingSyst) {
    if (-not $Auto) {
        $confirm = Read-Host "syst is already installed at $($existingSyst.Path). Download and install again? (y/N)"
        if ($confirm -ne 'y' -and $confirm -ne 'Y') {
            Write-Host "Aborting installation."
            exit 0
        }
    }
}

# Get latest release tag from GitHub
try {
    $releaseApi = 'https://api.github.com/repos/redjax/syst/releases/latest'
    $release = Invoke-RestMethod -Uri $releaseApi -UseBasicParsing
} catch {
    Write-Error "Failed to fetch latest release info: $_"
    exit 1
}

$version = $release.tag_name.TrimStart('v')

Write-Host "Installing syst version $version"

# Determine architecture for asset naming
$archMap = @{
    'AMD64' = 'amd64'
    'X64' = 'amd64'
    'ARM64' = 'arm64'
}
$machineArch = (Get-CimInstance Win32_Processor).Architecture
# Fallback to environment var if CIM not available
if (-not $machineArch) { $machineArch = $env:PROCESSOR_ARCHITECTURE }

# Normalize architecture string
$archNorm = if ($archMap.ContainsKey($machineArch)) {
    $archMap[$machineArch]
} else {
    Write-Error "Unsupported architecture: $machineArch"
    exit 1
}

# Build asset file name
$fileName = "syst-windows-$archNorm-$version.zip"
$downloadUrl = "https://github.com/redjax/syst/releases/download/$($release.tag_name)/$fileName"

# Create temp folder
$tempDir = Join-Path -Path ([System.IO.Path]::GetTempPath()) -ChildPath ("syst_install_" + [Guid]::NewGuid())
New-Item -ItemType Directory -Path $tempDir | Out-Null
$zipPath = Join-Path $tempDir $fileName

Write-Host "Downloading $fileName from GitHub..."
try {
    Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath -UseBasicParsing
} catch {
    Write-Error "Failed to download asset: $_"
    Remove-Item -Recurse -Force $tempDir
    exit 1
}

Write-Host "Extracting package..."
try {
    Expand-Archive -Path $zipPath -DestinationPath $tempDir -Force
} catch {
    Write-Error "Failed to extract archive: $_"
    Remove-Item -Recurse -Force $tempDir
    exit 1
}

# Expect binary named "syst.exe" in root of zip
$systExePath = Join-Path $tempDir 'syst.exe'
if (-not (Test-Path $systExePath)) {
    Write-Error "Extracted package missing expected syst.exe"
    Remove-Item -Recurse -Force $tempDir
    exit 1
}

# Copy executable to install path
$destExePath = Join-Path $installPath 'syst.exe'
try {
    Copy-Item -Path $systExePath -Destination $destExePath -Force
} catch {
    Write-Error "Failed to install syst.exe: $_"
    Remove-Item -Recurse -Force $tempDir
    exit 1
}

# Clean up temp files
Remove-Item -Recurse -Force $tempDir

Write-Host "syst installed successfully to $destExePath"

# Check if $installPath is already in user's PATH
$userPath = [Environment]::GetEnvironmentVariable('PATH', [EnvironmentVariableTarget]::User)
if (-not $userPath.Split(';') -contains $installPath) {
    Write-Host "`nNOTE: '$installPath' is not in your user PATH environment variable."
    Write-Host "Add it by running this once in PowerShell:"
    Write-Host "[Environment]::SetEnvironmentVariable('PATH', $userPath + ';$installPath', 'User')"
    Write-Host "Then restart your PowerShell session for changes to take effect."
}

exit 0