#!/bin/bash
set -e

# ==============================================================================
# ⚡ TailRouter v1.0.0 (Python Edition) - Universal Installer
# ==============================================================================

REPO="giangsamne/TailRouter"
TAG="v1.0.0"
INSTALL_DIR="$HOME/.tailrouter"

echo "=========================================================="
echo "⚡ TailRouter v1.0.0 (Python Edition) Installer"
echo "=========================================================="

if ! command -v python3 >/dev/null 2>&1; then
  echo "❌ Error: Python 3 is required to run TailRouter v1.0.0."
  exit 1
fi

echo "==> [1/3] Setting up installation in $INSTALL_DIR..."
mkdir -p "$INSTALL_DIR"

if command -v git >/dev/null 2>&1; then
  if [ -d "$INSTALL_DIR/.git" ]; then
    echo "==> Fetching tag $TAG..."
    cd "$INSTALL_DIR" && git fetch --tags && git checkout "$TAG"
  else
    echo "==> Cloning tag $TAG..."
    git clone -b "$TAG" "https://github.com/${REPO}.git" "$INSTALL_DIR"
  fi
else
  echo "==> Downloading source archive for $TAG..."
  TMP_TAR="/tmp/tailrouter_${TAG}.tar.gz"
  curl -fsSL "https://github.com/${REPO}/archive/refs/tags/${TAG}.tar.gz" -o "$TMP_TAR"
  tar -xzf "$TMP_TAR" -C "$INSTALL_DIR" --strip-components=1
  rm -f "$TMP_TAR"
fi

echo "==> [2/3] Setting up launcher symlink..."
mkdir -p "$HOME/.local/bin"
cat << 'EOF' > "$HOME/.local/bin/tailrouter"
#!/bin/sh
INSTALL_DIR="$HOME/.tailrouter"
if [ -d "$INSTALL_DIR/main" ]; then
  exec python3 "$INSTALL_DIR/main/server.py" "$@"
else
  exec python3 "$INSTALL_DIR/server.py" "$@"
fi
EOF
chmod +x "$HOME/.local/bin/tailrouter"

echo "==> [3/3] Starting TailRouter v1.0.0..."
cd "$INSTALL_DIR"
if [ -f "./start.sh" ]; then
  ./start.sh
else
  python3 main/server.py &
fi

echo "=========================================================="
echo "✅ TailRouter v1.0.0 installed and running!"
echo "👉 Web Dashboard: http://localhost:65534/router"
echo "=========================================================="
