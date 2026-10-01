#!/bin/bash
# ==============================================================================
# install-service.sh - Cài đặt TailRouter tự khởi động cùng hệ thống và tạo lệnh CLI
# Hỗ trợ:
#   - macOS: Apple launchd (LaunchAgent ~/Library/LaunchAgents/com.tailrouter.gateway.plist)
#   - Linux: systemd user service + linger (hoặc crontab @reboot trên Alpine/OpenRC)
#   - Toàn hệ thống: Cài đặt lệnh 'tailrouter' vào /usr/local/bin hoặc ~/.local/bin
# ==============================================================================

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
USER_NAME="$(whoami)"
OS="$(uname -s)"
PYTHON_BIN="$(which python3 2>/dev/null || echo "/usr/bin/python3")"

# Cấp quyền thực thi cho tailrouter CLI và các script
chmod +x "$DIR/tailrouter" "$DIR/start.sh" "$DIR/stop.sh" 2>/dev/null || true

# Chế độ gỡ cài đặt (Uninstall)
if [ "$1" = "--uninstall" ] || [ "$1" = "-u" ]; then
    echo "🛑 Đang gỡ bỏ dịch vụ tự khởi động TailRouter..."
    if [ -f "$DIR/service_helper.py" ]; then
        "$PYTHON_BIN" "$DIR/service_helper.py" disable
    fi
    if [ "$OS" = "Darwin" ]; then
        PLIST="$HOME/Library/LaunchAgents/com.tailrouter.gateway.plist"
        if [ -f "$PLIST" ]; then
            launchctl unload -w "$PLIST" 2>/dev/null || true
            rm -f "$PLIST"
        fi
    else
        SERVICE_NAME="tailscale-port-router"
        if command -v systemctl >/dev/null 2>&1; then
            systemctl --user disable "$SERVICE_NAME.service" 2>/dev/null || true
            systemctl --user stop "$SERVICE_NAME.service" 2>/dev/null || true
            rm -f "$HOME/.config/systemd/user/$SERVICE_NAME.service"
            systemctl --user daemon-reload 2>/dev/null || true
        fi
    fi
    echo "✅ Đã gỡ bỏ tự khởi động TailRouter."
    exit 0
fi

echo "🚀 Đang thiết lập TailRouter CLI và tự khởi động ($OS)..."

# 1. Cài đặt lệnh CLI 'tailrouter' vào PATH
if [ -w "/usr/local/bin" ] || [ "$(id -u)" = "0" ]; then
    ln -sf "$DIR/tailrouter" /usr/local/bin/tailrouter
    echo "✅ Đã tạo lệnh CLI toàn hệ thống: /usr/local/bin/tailrouter"
else
    mkdir -p "$HOME/.local/bin"
    ln -sf "$DIR/tailrouter" "$HOME/.local/bin/tailrouter"
    echo "✅ Đã tạo lệnh CLI người dùng:    $HOME/.local/bin/tailrouter"
fi

# 2. Thiết lập tự khởi động qua service_helper.py (đa nền tảng)
if [ -f "$DIR/service_helper.py" ]; then
    "$PYTHON_BIN" "$DIR/service_helper.py" enable
fi

echo "============================================================"
echo "🎉 Đã hoàn tất cài đặt TailRouter v1.0.1!"
echo "👉 Từ bây giờ bạn có thể gõ 'tailrouter' trong terminal bất kỳ lúc nào:"
echo "   ● tailrouter           : Bật chạy nhanh hoặc xem thông tin Gateway"
echo "   ● tailrouter status    : Xem trạng thái và link Bảng điều khiển"
echo "   ● tailrouter stop      : Dừng dịch vụ"
echo "   ● tailrouter restart   : Khởi động lại dịch vụ"
echo "   ● tailrouter scan      : Dò quét cổng máy thật (sắp xếp tăng dần)"
echo "   ● tailrouter autostart : Quản lý tự khởi động khi mở máy"
echo "============================================================"
