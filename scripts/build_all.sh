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
echo "==> [1/4] Đang biên dịch Linux (x86_64 & ARM64)..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/linux/tailrouter ./cmd/tailrouter
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o dist/linux/tailrouter-arm64 ./cmd/tailrouter

# 2. Windows Binaries
echo "==> [2/4] Đang biên dịch Windows (x86_64)..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/windows/tailrouter.exe ./cmd/tailrouter

# 3. macOS Binaries
echo "==> [3/4] Đang biên dịch macOS (Apple Silicon & Intel)..."
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/macos/tailrouter-arm64 ./cmd/tailrouter
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/macos/tailrouter-amd64 ./cmd/tailrouter

# 4. Packaging Release Archives
echo "==> [4/4] Đóng gói các tệp phát hành (Releases)..."

# Linux Package
tar -czf release/TailRouter-Linux.tar.gz -C dist/linux tailrouter tailrouter-arm64

# Windows Package
cd dist/windows && zip -r ../../release/TailRouter-Windows.zip tailrouter.exe && cd ../..

# macOS Package
cd dist/macos && zip -r ../../release/TailRouter-macOS.zip tailrouter-arm64 tailrouter-amd64 && cd ../..

echo "==> Đóng gói hoàn tất!"
ls -lh release/
