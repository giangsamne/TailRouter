#!/bin/bash
set -e

# ==============================================================================
# ⚡ TailRouter - Universal Smart Installer for macOS & Linux
# Auto-detects OS, Architecture, and Init System to install the right package.
# ==============================================================================

REPO="giangsamne/TailRouter"
RELEASE_TAG="v1.1.0"
BASE_URL="https://github.com/${REPO}/releases/download/${RELEASE_TAG}"

echo "=========================================================="
echo "⚡ TailRouter Universal Installer ($RELEASE_TAG)"
echo "=========================================================="

OS="$(uname -s)"
ARCH="$(uname -m)"

echo "==> [1/4] Detecting environment..."
echo "    Operating System: $OS"
echo "    Hardware Arch:    $ARCH"

case "$OS" in
  Darwin)
    echo "==> [2/4] Preparing macOS installation..."
    ARCHIVE_NAME="TailRouter-macOS.zip"
    DOWNLOAD_URL="${BASE_URL}/${ARCHIVE_NAME}"
    TMP_DIR="/tmp/tailrouter_install_$$"

    mkdir -p "$TMP_DIR"
    echo "==> [3/4] Downloading $ARCHIVE_NAME from GitHub Releases..."
    curl -fSL --progress-bar "$DOWNLOAD_URL" -o "$TMP_DIR/$ARCHIVE_NAME"

    echo "==> [4/4] Extracting and installing binary..."
    unzip -q -o "$TMP_DIR/$ARCHIVE_NAME" -d "$TMP_DIR"

    BIN_NAME="tailrouter-arm64"
    if [ "$ARCH" = "x86_64" ]; then
      BIN_NAME="tailrouter-amd64"
    fi

    TARGET_BIN="/usr/local/bin/tailrouter"
    if [ -w "/usr/local/bin" ] || [ "$(id -u)" -eq 0 ]; then
      cp "$TMP_DIR/$BIN_NAME" "$TARGET_BIN"
      chmod +x "$TARGET_BIN"
    elif command -v sudo >/dev/null 2>&1 && sudo -n true 2>/dev/null; then
      sudo cp "$TMP_DIR/$BIN_NAME" "$TARGET_BIN"
      sudo chmod +x "$TARGET_BIN"
    else
      TARGET_BIN="$HOME/.local/bin/tailrouter"
      mkdir -p "$HOME/.local/bin"
      cp "$TMP_DIR/$BIN_NAME" "$TARGET_BIN"
      chmod +x "$TARGET_BIN"
      if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
        export PATH="$HOME/.local/bin:$PATH"
      fi
    fi

    rm -rf "$TMP_DIR"

    echo "=========================================================="
    echo "✅ TailRouter installed successfully at: $TARGET_BIN"
    echo "==> Setting up system service..."
    "$TARGET_BIN" service install 2>/dev/null || true
    echo "==> Checking gateway status..."
    "$TARGET_BIN" status || true
    echo "=========================================================="
    echo "💡 Web Dashboard accessible at: http://localhost:65534/router"
    open "http://localhost:65534/router" 2>/dev/null || true
    ;;

  Linux)
    echo "==> [2/4] Preparing Linux installation..."
    ARCHIVE_NAME="TailRouter-Linux.tar.gz"
    DOWNLOAD_URL="${BASE_URL}/${ARCHIVE_NAME}"
    TMP_DIR="/tmp/tailrouter_install_$$"

    mkdir -p "$TMP_DIR"
    echo "==> [3/4] Downloading $ARCHIVE_NAME from GitHub Releases..."
    curl -fSL --progress-bar "$DOWNLOAD_URL" -o "$TMP_DIR/$ARCHIVE_NAME"

    echo "==> [4/4] Extracting and installing binaries..."
    tar -xzf "$TMP_DIR/$ARCHIVE_NAME" -C "$TMP_DIR"

    # Determine binary architecture
    BIN_NAME="tailrouter"
    if [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
      BIN_NAME="tailrouter-arm64"
    fi

    # Determine install destination (prefer /usr/local/bin if writable/root, else ~/.local/bin)
    TARGET_BIN="/usr/local/bin/tailrouter"
    if [ -w "/usr/local/bin" ] || [ "$(id -u)" -eq 0 ]; then
      cp "$TMP_DIR/$BIN_NAME" "$TARGET_BIN"
      chmod +x "$TARGET_BIN"
    elif command -v sudo >/dev/null 2>&1 && sudo -n true 2>/dev/null; then
      sudo cp "$TMP_DIR/$BIN_NAME" "$TARGET_BIN"
      sudo chmod +x "$TARGET_BIN"
    else
      TARGET_BIN="$HOME/.local/bin/tailrouter"
      mkdir -p "$HOME/.local/bin"
      cp "$TMP_DIR/$BIN_NAME" "$TARGET_BIN"
      chmod +x "$TARGET_BIN"
      if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
        export PATH="$HOME/.local/bin:$PATH"
      fi
    fi

    rm -rf "$TMP_DIR"

    echo "=========================================================="
    echo "✅ TailRouter installed at: $TARGET_BIN"
    echo "==> Setting up system service..."
    "$TARGET_BIN" service install 2>/dev/null || true
    echo "==> Checking gateway status..."
    "$TARGET_BIN" status || true
    echo "=========================================================="
    echo "💡 Web Dashboard accessible at: http://localhost:65534/router"
    ;;

  *)
    echo "❌ Unsupported operating system: $OS"
    echo "For Windows, please run install.ps1 in PowerShell"
    exit 1
    ;;
esac
