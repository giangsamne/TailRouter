#!/bin/bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -f "$SCRIPT_DIR/main/stop.sh" ]; then
  exec "$SCRIPT_DIR/main/stop.sh" "$@"
elif [ -f "$SCRIPT_DIR/stop.sh" ]; then
  exec "$SCRIPT_DIR/stop.sh" "$@"
else
  pkill -f "python3.*server.py" || true
fi
