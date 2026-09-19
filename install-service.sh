#!/bin/bash
# ==============================================================================
# install-service.sh - Cài đặt Tailscale Port Router thành systemd service
# Tự động khởi động cùng hệ thống máy chủ thật
# ==============================================================================

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
USER_NAME="$(whoami)"
SERVICE_NAME="tailscale-port-router"
SYSTEMD_USER_DIR="$HOME/.config/systemd/user"
SERVICE_FILE="$SYSTEMD_USER_DIR/$SERVICE_NAME.service"

echo "⚙️  Đang thiết lập systemd user service cho $USER_NAME..."

mkdir -p "$SYSTEMD_USER_DIR"

cat << EOF > "$SERVICE_FILE"
[Unit]
Description=Tailscale Port Router & Gateway (Port 65534)
After=network.target tailscaled.service

[Service]
Type=simple
WorkingDirectory=$DIR
ExecStart=/usr/bin/python3 $DIR/server.py
ExecStartPost=-/snap/bin/tailscale serve --bg --set-path /router http://127.0.0.1:65534/router
Restart=always
RestartSec=5
StandardOutput=append:$DIR/server.log
StandardError=append:$DIR/server.log

[Install]
WantedBy=default.target
EOF

echo "✅ Đã tạo file service tại: $SERVICE_FILE"

# Kích hoạt systemd service
systemctl --user daemon-reload
systemctl --user enable "$SERVICE_NAME.service"
systemctl --user restart "$SERVICE_NAME.service"

echo "============================================================"
echo "🎉 Đã kích hoạt dịch vụ tự khởi động cùng máy chủ!"
echo "👉 Kiểm tra trạng thái: systemctl --user status $SERVICE_NAME"
echo "👉 Xem log:             journalctl --user -u $SERVICE_NAME -f"
echo "============================================================"
