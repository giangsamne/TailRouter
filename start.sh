#!/bin/bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -f "$SCRIPT_DIR/main/server.py" ]; then
  cd "$SCRIPT_DIR/main" && exec python3 server.py "$@"
elif [ -f "$SCRIPT_DIR/server.py" ]; then
  exec python3 "$SCRIPT_DIR/server.py" "$@"
fi
