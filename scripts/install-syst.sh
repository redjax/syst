#!/usr/bin/env bash

set -e

AUTO_MODE=false

if ! command -v curl &>/dev/null; then
  echo "curl is not installed."
  exit 1
fi

if ! command -v unzip &>/dev/null; then
  echo "unzip is not installed."
  exit 1
fi

## Parse arguments
for arg in "$@"; do
  if [[ "$arg" == "--auto" ]]; then
    AUTO_MODE=true
  fi
done

if command -v syst &>/dev/null; then
  if [[ "$AUTO_MODE" == true ]]; then
    CONFIRM="y"
  else
    read -p "syst is already installed. Download anyway? (y/N) " CONFIRM
  fi

  if [[ "$CONFIRM" != "y" ]]; then
    echo "Aborting."
    exit 0
  fi
fi

## Detect OS
OS="$(uname -s)"
ARCH="$(uname -m)"

## Detect distro using /etc/os-release
if [ -f /etc/os-release ]; then
  . /etc/os-release
  DISTRO_ID=$ID
else
  echo "Cannot detect Linux distribution."
  exit 1
fi

## Get latest version from GitHub API
SYST_VERSION=$(curl -s https://api.github.com/repos/redjax/syst/releases/latest | grep -Po '"tag_name": "\K.*?(?=")')
## Remove leading 'v' if present (e.g., 'v${SYST_VERSION}' -> '${SYST_VERSION}')
SYST_VERSION="${SYST_VERSION#v}"

echo "Installing syst v${SYST_VERSION}"

## Map to GitHub asset names
case "$OS" in
Linux)
  case "$ARCH" in
  x86_64)
    FILE="linux-amd64-${SYST_VERSION}.zip"
    ;;
  aarch64 | arm64)
    FILE="linux-arm64-${SYST_VERSION}.zip"
    ;;
  *)
    echo "Unsupported Linux architecture: $ARCH"
    exit 1
    ;;
  esac
  ;;
Darwin)
  case "$ARCH" in
  x86_64)
    FILE="macOS-${SYST_VERSION}.zip"
    ;;
  arm64)
    FILE="macOS-${SYST_VERSION}.zip"
    ;;
  *)
    echo "Unsupported macOS architecture: $ARCH"
    exit 1
    ;;
  esac
  ;;
*)
  echo "Unsupported OS: $OS"
  exit 1
  ;;
esac

## Create a temporary directory
TMPDIR=$(mktemp -d)

## Download the release
URL="https://github.com/redjax/syst/releases/download/${SYST_VERSION}/$FILE"
ARCHIVE="$TMPDIR/syst.zip"

## Download the archive to the temp directory
echo "Downloading $FILE from $URL"
curl -L -o "$ARCHIVE" "$URL"

## Extract the archive into the temp directory
unzip "$ARCHIVE" -d "$TMPDIR"
if [ $? -ne 0 ]; then
  echo "Failed to extract $ARCHIVE to $TMPDIR"
  exit 1
fi

if [[ ! -d "$HOME/.local/bin" ]]; then
  mkdir -p "$HOME/.local/bin"
fi

if [ "$OS" = "Darwin" ]; then
  ## macOS: install to $HOME/.local/bin/
  install -m 755 "$TMPDIR/syst" $HOME/.local/bin/
else
  ## Linux: install to $HOME/.local/bin/
  sudo install -m 755 "$TMPDIR/syst" $HOME/.local/bin/
fi

echo "syst installed successfully!"
echo "You may need to add this to your ~/.bashrc:"
echo "export PATH=\$PATH:\$HOME/.local/bin"

exit 0
