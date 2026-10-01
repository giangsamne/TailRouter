#!/bin/sh
# ==============================================================================
# install-cli.sh - Cài đặt lệnh 'tailrouter' vào biến môi trường PATH
# ==============================================================================

DIR="$(cd "$(dirname "$0")" && pwd)"
chmod +x "$DIR/tailrouter" "$DIR/start.sh" "$DIR/stop.sh" 2>/dev/null || true

if [ -w "/usr/local/bin" ] || [ "$(id -u)" = "0" ]; then
    ln -sf "$DIR/tailrouter" /usr/local/bin/tailrouter
    echo "✅ Đã cài đặt lệnh 'tailrouter' vào /usr/local/bin/tailrouter"
else
    mkdir -p "$HOME/.local/bin"
    ln -sf "$DIR/tailrouter" "$HOME/.local/bin/tailrouter"
    echo "✅ Đã cài đặt lệnh 'tailrouter' vào $HOME/.local/bin/tailrouter"
fi

echo "👉 Thử gõ ngay: tailrouter status"
