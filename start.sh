#!/bin/sh
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "=========================================================="
echo "⚡ TailRouter - Starting Gateway..."
echo "=========================================================="

# 1. Stop existing instances to avoid port 65534 conflict
pkill -f "tailrouter" 2>/dev/null || killall tailrouter 2>/dev/null || true
pkill -f "server.py" 2>/dev/null || true
sleep 1

# 2. Check local Python engine first if present in directory
if [ -f "$SCRIPT_DIR/main/server.py" ] && command -v python3 >/dev/null 2>&1; then
  cd "$SCRIPT_DIR/main" && exec python3 server.py "$@"
elif [ -f "$SCRIPT_DIR/server.py" ] && command -v python3 >/dev/null 2>&1; then
  exec python3 "$SCRIPT_DIR/server.py" "$@"
# 3. Prefer native Go binary if present in directory
elif [ -x "$SCRIPT_DIR/tailrouter" ]; then
  exec "$SCRIPT_DIR/tailrouter" "$@"
elif [ -x "$SCRIPT_DIR/dist/linux/tailrouter" ] && [ "$(uname -m)" = "x86_64" ]; then
  exec "$SCRIPT_DIR/dist/linux/tailrouter" "$@"
elif [ -x "$SCRIPT_DIR/dist/linux/tailrouter-arm64" ] && { [ "$(uname -m)" = "aarch64" ] || [ "$(uname -m)" = "arm64" ]; }; then
  exec "$SCRIPT_DIR/dist/linux/tailrouter-arm64" "$@"
elif [ -x "$HOME/.local/bin/tailrouter" ]; then
  exec "$HOME/.local/bin/tailrouter" "$@"
elif command -v tailrouter >/dev/null 2>&1; then
  exec tailrouter "$@"
# 4. If nothing is built/installed yet, run installer
elif [ -f "$SCRIPT_DIR/install.sh" ]; then
  echo "==> Setting up TailRouter via installer..."
  sh "$SCRIPT_DIR/install.sh"
else
  echo "❌ Could not find TailRouter binary or runtime."
  exit 1
fi
