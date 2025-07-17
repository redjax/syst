<#
    .SYNOPSIS
    My Go build script.

    .DESCRIPTION
    Accepts a common set of parameters to automate (re)building a Go app.

    .PARAMETER BinName
    The name of your executable, i.e. ./`$BinName.

    .PARAMETER BuildOS
    The OS to build for. Full list available at https://github.com/golang/go/blob/master/src/internal/syslist/syslist.go

    .PARAMETER BuildArch
    The CPU architecture to build for. A full list does not seem to be available,
    but more info in this Gist: https://gist.github.com/asukakenji/f15ba7e588ac42795f421b48b8aede63

    .PARAMETER BuildOutputDir
    The build artifact path, where build outputs will be saved.

    .PARAMETER BuildTarget
    The name of the file to build (the entrypoint for your app).

    .EXAMPLE
    .\build.ps1 -BinName "syst" -BuildOS "windows" -BuildArch "amd64" -BuildOutputDir "dist/"
#>
Param(
    [Parameter(Mandatory = $false, HelpMessage = "The name of your executable, i.e. ./`$BinName.")]
    $BinName = "syst",
    [Parameter(Mandatory = $false, HelpMessage = "The OS to build for. Full ist available at https://github.com/golang/go/blob/master/src/internal/syslist/syslist.go")]
    $BuildOS = "windows",
    [Parameter(Mandatory = $false, HelpMessage = "The CPU architecture to build for. A full list does not seem to be available, but more info in this Gist: https://gist.github.com/asukakenji/f15ba7e588ac42795f421b48b8aede63")]
    $BuildArch = "amd64",
    [Parameter(Mandatory = $false, HelpMessage = "The build artifact path, where build outputs will be saved.")]
    $BuildOutputDir = "dist/",
    [Parameter(Mandatory = $false, HelpMessage = "The name of the file to build (the entrypoint for your app).")]
    $BuildTarget = "./cmd/entrypoint",
    [Parameter(Mandatory = $false, HelpMessage = "The name you gave the app during go mod init. Check in go.mod to see the value the app is using.")]
    $ModulePath = "github.com/redjax/syst"
)

## Get Git metadata
try {
    $GitVersion = (git describe --tags --always).Trim()
} catch {
    $GitVersion = "dev"
}

try {
    $GitCommit = (git rev-parse --short HEAD).Trim()
} catch {
    $GitCommit = "none"
}

$BuildDate = (Get-Date -Format "yyyy-MM-ddTHH:mm:ssZ")

## Set up ldflags to inject version metadata
$LdFlags = "-s -w " +
    "-X `"$ModulePath.Version=$GitVersion`" " +
    "-X `"$ModulePath.Commit=$GitCommit`" " +
    "-X `"$ModulePath.Date=$BuildDate`""

Write-Debug "BinName: $BinName"
Write-Debug "BuildOS: $BuildOS"
Write-Debug "BuildArch: $BuildArch"
Write-Debug "BuildOutputDir: $BuildOutputDir"
Write-Debug "BuildTarget: $BuildTarget"
Write-Debug "GitVersion: $GitVersion"
Write-Debug "BuildDate: $BuildDate"

if ( $null -eq $BinName ) {
    Write-Warning "No bin name provided, pass the name of your executable using the -BinName flag"
    exit(1)
}

if ( ($BuildOS -eq "windows") -and ( -not $BinName.EndsWith(".exe") ) ) {
    Write-Warning "Building for Windows but bin name does not end with '.exe'. Appending .exe to '$BinName'"
    $BinName += ".exe"
}

$env:GOOS = $BuildOS
$env:GOARCH = $BuildArch

$BuildOutput = Join-Path -Path $BuildOutputDir -ChildPath $BinName
Write-Debug "Build output: $BuildOutput"

Write-Host "Building $($BuildTarget), outputting to $($BuildOutput)" -ForegroundColor Cyan
Write-Information "-- [ Build start"
try {
    go build -ldflags "$LdFlags" -o $BuildOutput $BuildTarget
    Write-Host "Build successful" -ForegroundColor Green
}
catch {
    Write-Error "Error building app. Details: $($_.Exception.Message)"
    exit(1)
}
finally {
    Write-Information "-- [ Build complete"
}

exit(0)
