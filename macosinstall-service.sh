#!/bin/bash
# ==============================================================================
# install-service.sh - Cài đặt TailRouter tự khởi động khi bật máy và chạy ngầm
# Hỗ trợ tự động nhận diện:
#   - macOS: Apple launchd (LaunchAgent ~/Library/LaunchAgents/com.tailrouter.gateway.plist)
#   - Linux: systemd user service + linger
# ==============================================================================

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
USER_NAME="$(whoami)"
OS="$(uname -s)"
PYTHON_BIN="$(which python3 2>/dev/null || echo "/usr/bin/python3")"

# Chế độ gỡ cài đặt (Uninstall)
if [ "$1" = "--uninstall" ] || [ "$1" = "-u" ]; then
    echo "🛑 Đang gỡ bỏ dịch vụ tự khởi động TailRouter..."
    if [ "$OS" = "Darwin" ]; then
        PLIST="$HOME/Library/LaunchAgents/com.tailrouter.gateway.plist"
        if [ -f "$PLIST" ]; then
            launchctl unload -w "$PLIST" 2>/dev/null || true
            rm -f "$PLIST"
            echo "✅ Đã gỡ bỏ LaunchAgent khỏi macOS."
        else
            echo "ℹ️  LaunchAgent chưa được cài đặt."
        fi
    else
        SERVICE_NAME="tailscale-port-router"
        systemctl --user disable "$SERVICE_NAME.service" 2>/dev/null || true
        systemctl --user stop "$SERVICE_NAME.service" 2>/dev/null || true
        rm -f "$HOME/.config/systemd/user/$SERVICE_NAME.service"
        systemctl --user daemon-reload 2>/dev/null || true
        echo "✅ Đã gỡ bỏ systemd service khỏi Linux."
    fi
    exit 0
fi

echo "🚀 Đang thiết lập tự khởi động cho TailRouter ($OS)..."

if [ "$OS" = "Darwin" ]; then
    AGENT_DIR="$HOME/Library/LaunchAgents"
    PLIST="$AGENT_DIR/com.tailrouter.gateway.plist"

    mkdir -p "$AGENT_DIR"

    cat << PLIST_EOF > "$PLIST"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.tailrouter.gateway</string>
    <key>ProgramArguments</key>
    <array>
        <string>$PYTHON_BIN</string>
        <string>$DIR/server.py</string>
    </array>
    <key>WorkingDirectory</key>
    <string>$DIR</string>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>$DIR/server.log</string>
    <key>StandardErrorPath</key>
    <string>$DIR/server.log</string>
    <key>EnvironmentVariables</key>
    <dict>
        <key>PATH</key>
        <string>/usr/local/bin:/opt/homebrew/bin:/usr/bin:/bin:/usr/sbin:/sbin</string>
        <key>SHLVL</key>
        <string>1</string>
    </dict>
</dict>
</plist>
PLIST_EOF

    echo "✅ Đã tạo LaunchAgent tại: $PLIST"
    launchctl unload -w "$PLIST" 2>/dev/null || true
    launchctl load -w "$PLIST"

    sleep 1
    echo "============================================================"
    echo "🎉 Đã kích hoạt tự khởi động thành công trên macOS (launchd)!"
    echo "👉 Dịch vụ sẽ tự động chạy ngầm mỗi khi mở máy mà không cần terminal."
    echo "👉 Bảng Quản Trị:          http://localhost:65534/router"
    echo "👉 Xem log hoạt động:      tail -f $DIR/server.log"
    echo "🛑 Để gỡ tự khởi động:     ./install-service.sh --uninstall"
    echo "============================================================"

else
    SERVICE_NAME="tailscale-port-router"
    SYSTEMD_USER_DIR="$HOME/.config/systemd/user"
    SERVICE_FILE="$SYSTEMD_USER_DIR/$SERVICE_NAME.service"

    mkdir -p "$SYSTEMD_USER_DIR"

    cat << SERVICE_EOF > "$SERVICE_FILE"
[Unit]
Description=TailRouter & Gateway (Port 65534)
After=network.target tailscaled.service

[Service]
Type=simple
WorkingDirectory=$DIR
ExecStart=$PYTHON_BIN $DIR/server.py
Restart=always
RestartSec=5
StandardOutput=append:$DIR/server.log
StandardError=append:$DIR/server.log

[Install]
WantedBy=default.target
SERVICE_EOF

    echo "✅ Đã tạo file service tại: $SERVICE_FILE"

    systemctl --user daemon-reload
    systemctl --user enable "$SERVICE_NAME.service"
    systemctl --user restart "$SERVICE_NAME.service"

    loginctl enable-linger "$USER_NAME" 2>/dev/null || true

    echo "============================================================"
    echo "🎉 Đã kích hoạt tự khởi động thành công trên Linux (systemd)!"
    echo "👉 Kiểm tra trạng thái:    systemctl --user status $SERVICE_NAME"
    echo "👉 Xem log:                journalctl --user -u $SERVICE_NAME -f"
    echo "🛑 Để gỡ tự khởi động:     ./install-service.sh --uninstall"
    echo "============================================================"
fi
