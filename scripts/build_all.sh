#!/usr/bin/env bash
set -e

VERSION="v2.3.0"
APP_NAME="tailrouter"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
DIST_DIR="$ROOT_DIR/dist"
RELEASE_DIR="$ROOT_DIR/release"

export PATH="/home/giang3dlab/.local/go/bin:$PATH"

echo "=========================================================="
echo " Building TailRouter $VERSION for All OS "
echo "=========================================================="

rm -rf "$DIST_DIR" "$RELEASE_DIR"
mkdir -p "$DIST_DIR" "$RELEASE_DIR"

cd "$ROOT_DIR"

# Targets: GOOS/GOARCH
TARGETS=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
)

for TARGET in "${TARGETS[@]}"; do
  OS="${TARGET%%/*}"
  ARCH="${TARGET##*/}"
  
  OUTPUT_NAME="${APP_NAME}-${OS}-${ARCH}"
  if [ "$OS" = "windows" ]; then
    OUTPUT_NAME="${OUTPUT_NAME}.exe"
  fi

  echo "🔨 Compiling for $OS ($ARCH)..."
  CGO_ENABLED=0 GOOS="$OS" GOARCH="$ARCH" go build -ldflags="-s -w" -o "${DIST_DIR}/${OUTPUT_NAME}" main.go

  # Package into individual 1-file archive (git-webui style)
  cd "$DIST_DIR"
  if [ "$OS" = "windows" ]; then
    zip -q "${RELEASE_DIR}/${APP_NAME}-${OS}-${ARCH}.zip" "${OUTPUT_NAME}"
  else
    tar -czf "${RELEASE_DIR}/${APP_NAME}-${OS}-${ARCH}.tar.gz" "${OUTPUT_NAME}"
  fi
  cd "$ROOT_DIR"
done

# Backwards compatibility bundle packages
cd "$DIST_DIR"
cp tailrouter-linux-amd64 tailrouter
cp tailrouter-linux-arm64 tailrouter-arm64
tar -czf "${RELEASE_DIR}/TailRouter-Linux.tar.gz" tailrouter tailrouter-arm64
rm -f tailrouter tailrouter-arm64

cp tailrouter-darwin-amd64 tailrouter-amd64
cp tailrouter-darwin-arm64 tailrouter-arm64
zip -q "${RELEASE_DIR}/TailRouter-macOS.zip" tailrouter-amd64 tailrouter-arm64
rm -f tailrouter-amd64 tailrouter-arm64

cp tailrouter-windows-amd64.exe tailrouter.exe
zip -q "${RELEASE_DIR}/TailRouter-Windows.zip" tailrouter.exe
rm -f tailrouter.exe
cd "$ROOT_DIR"

echo "✅ All builds completed in ${DIST_DIR}/ and ${RELEASE_DIR}/:"
ls -lh "$RELEASE_DIR"
