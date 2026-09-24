#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

export PATH="/home/giang3dlab/.local/go/bin:$PATH"

echo "=========================================================="
echo "⚡ TailRouter Native Engine - Multi-Platform Build System"
echo "=========================================================="

cd "$ROOT_DIR"
mkdir -p dist/linux dist/windows dist/macos release

# 1. Linux Binaries
echo "==> [1/4] Đang biên dịch Linux (Server CLI & Desktop)..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/linux/tailrouter-server ./cmd/tailrouter
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o dist/linux/tailrouter-server-arm64 ./cmd/tailrouter
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/linux/tailrouter-desktop ./cmd/tailrouter-desktop

# 2. Windows Binaries
echo "==> [2/4] Đang biên dịch Windows (Server CLI & Desktop Tray)..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/windows/tailrouter-server.exe ./cmd/tailrouter
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -H=windowsgui" -o dist/windows/TailRouter.exe ./cmd/tailrouter-desktop

# 3. macOS Binaries
echo "==> [3/4] Đang biên dịch macOS Native Engine (Apple Silicon & Intel)..."
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/macos/tailrouter-server-arm64 ./cmd/tailrouter
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/macos/tailrouter-server-amd64 ./cmd/tailrouter

# 4. Packaging Release Archives
echo "==> [4/4] Đóng gói các tệp phát hành (Releases)..."

# Linux Package
tar -czf release/TailRouter-Linux.tar.gz -C dist/linux tailrouter-server tailrouter-server-arm64 tailrouter-desktop

# Windows Package
cd dist/windows && zip -r ../../release/TailRouter-Windows.zip TailRouter.exe tailrouter-server.exe && cd ../..

echo "==> Đóng gói hoàn tất!"
ls -lh release/
