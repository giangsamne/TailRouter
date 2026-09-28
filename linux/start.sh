#!/bin/bash
# ==============================================================================
# start.sh - Khởi động TailRouter trên cổng 65534
# ==============================================================================

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PID_FILE="$DIR/server.pid"
LOG_FILE="$DIR/server.log"
PORT=65534

# Kiểm tra xem cổng 65534 có đang bị chiếm dụng không
EXISTING_PID=$(ss -tulpn 2>/dev/null | grep ":$PORT " | grep -o 'pid=[0-9]*' | head -n 1 | cut -d= -f2)
if [ -z "$EXISTING_PID" ]; then
    EXISTING_PID=$(lsof -ti :$PORT 2>/dev/null)
fi

if [ -n "$EXISTING_PID" ]; then
    echo "⚠️  Cổng $PORT đã có tiến trình đang chạy (PID: $EXISTING_PID)."
    echo "👉 Truy cập ngay tại: http://localhost:$PORT/router"
    exit 0
fi

# Chạy foreground nếu có tham số -f hoặc --foreground
if [ "$1" = "-f" ] || [ "$1" = "--foreground" ]; then
    echo "🚀 Đang khởi động TailRouter ở chế độ Foreground (Port $PORT)..."
    python3 "$DIR/server.py"
    exit $?
fi

# Chạy ngầm (Daemon / Background)
echo "🚀 Đang khởi động TailRouter chạy nền trên cổng $PORT..."
if command -v setsid >/dev/null 2>&1; then
    setsid python3 "$DIR/server.py" </dev/null > "$LOG_FILE" 2>&1 &
elif command -v nohup >/dev/null 2>&1; then
    nohup python3 "$DIR/server.py" </dev/null > "$LOG_FILE" 2>&1 &
else
    python3 "$DIR/server.py" </dev/null > "$LOG_FILE" 2>&1 &
fi
SERVER_PID=$!
echo "$SERVER_PID" > "$PID_FILE"

sleep 1

# Kiểm tra xem server đã khởi động thành công chưa
if kill -0 $SERVER_PID 2>/dev/null; then
    # Lấy IP Tailscale nếu có
    TS_IP=$(tailscale ip -4 2>/dev/null || echo "")
    echo "============================================================"
    echo "✅ Khởi động thành công! (PID: $SERVER_PID)"
    echo "👉 Bảng Quản Trị Local:     http://localhost:$PORT/router"
    if [ -n "$TS_IP" ]; then
        echo "👉 Bảng Quản Trị Tailscale: http://$TS_IP:$PORT/router"
    fi

    # Tự động kích hoạt Tailscale Serve cho /router
    tailscale serve --bg --set-path /router http://127.0.0.1:$PORT/router 2>/dev/null || true

    echo "📄 File Log hoạt động:      $LOG_FILE"
    echo "🛑 Để dừng dịch vụ, chạy:   ./stop.sh"
    echo "============================================================"
else
    echo "❌ Không thể khởi động máy chủ. Xem chi tiết log tại $LOG_FILE:"
    cat "$LOG_FILE"
    exit 1
fi
