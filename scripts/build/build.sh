#!/bin/bash

set -euo

## Default param values
BinName="syst"
BuildOS="linux"
BuildArch="amd64"
BuildOutputDir="dist/"
BuildTarget="./cmd/entrypoint/main.go"

function print_help {
  echo "-- | Build Go module | --"
  echo ""
  echo "[ Flags ]"
  echo "  --help: Print this help menu"
  echo ""
  echo "  --bin-name (default: b2cleaner): Name for the binary output"
  echo "  --build-os (default: linux): Target OS to build for"
  echo "  --build-arch (default: amd4): Target CPU architecture to build for"
  echo "  --build-output-dir (default: dist/): Path to binary output directory"
  echo "  --build-target (default: ./main.go): Path to module entrypoint"
}

if ! command -v go --version >/dev/null 2>&1; then
  echo "[ERROR] Go is not installed."
  exit 1
fi

## Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --help)
      print_help
      exit 0
    ;;
    --bin-name)
      BinName="$2"
      shift 2
      ;;
    --build-os)
      BuildOS="$2"
      shift 2
      ;;
    --build-arch)
      BuildArch="$2"
      shift 2
      ;;
    --build-output-dir)
      BuildOutputDir="$2"
      shift 2
      ;;
    --build-target)
      BuildTarget="$2"
      shift 2
      ;;
    *)
      echo "Unknown parameter: $1"
      exit 1
      ;;
  esac
done

## Debug output
echo "BinName: $BinName"
echo "BuildOS: $BuildOS"
echo "BuildArch: $BuildArch"
echo "BuildOutputDir: $BuildOutputDir"
echo "BuildTarget: $BuildTarget"

## Check for BinName presence
if [[ -z "$BinName" ]]; then
  echo "Warning: No bin name provided, pass the name of your executable using the --bin-name flag"
  exit 1
fi

## Append .exe if building for Windows and BinName doesn't end with .exe
if [[ "$BuildOS" == "windows" && "${BinName##*.}" != "exe" ]]; then
  echo "Warning: Building for Windows but bin name does not end with '.exe'. Appending .exe to '$BinName'"
  BinName="${BinName}.exe"
fi

## Set environment variables for Go build
export GOOS="$BuildOS"
export GOARCH="$BuildArch"

## Ensure output directory exists
mkdir -p "$BuildOutputDir"

BuildOutput="${BuildOutputDir%/}/$BinName"

echo -e "Building $BuildTarget, outputting to $BuildOutput"
echo "-- [ Build start"

# Run go build
if go build -o "$BuildOutput" "$BuildTarget"; then
  echo -e "Build successful"
else
  echo "Error building app."
  exit 1
fi

echo "-- [ Build complete"

exit 0
