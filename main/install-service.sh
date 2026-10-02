#!/bin/sh
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "=========================================================="
echo "⚡ TailRouter - Installing Service & Autostart..."
echo "=========================================================="

if [ -f "$SCRIPT_DIR/install.sh" ]; then
  exec sh "$SCRIPT_DIR/install.sh" "$@"
elif [ -x "$HOME/.local/bin/tailrouter" ]; then
  exec "$HOME/.local/bin/tailrouter" service install
elif command -v tailrouter >/dev/null 2>&1; then
  exec tailrouter service install
fi
