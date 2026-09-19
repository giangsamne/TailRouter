#!/bin/bash
# ==============================================================================
# stop.sh - Dừng Tailscale Port Router trên cổng 65534
# ==============================================================================

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PID_FILE="$DIR/server.pid"
PORT=65534

echo "🛑 Đang kiểm tra tiến trình Tailscale Port Router..."

STOPPED=0

# 1. Dừng theo PID file
if [ -f "$PID_FILE" ]; then
    PID=$(cat "$PID_FILE")
    if [ -n "$PID" ] && kill -0 "$PID" 2>/dev/null; then
        echo "➡️  Đang dừng tiến trình PID: $PID..."
        kill "$PID"
        sleep 1
        if kill -0 "$PID" 2>/dev/null; then
            kill -9 "$PID"
        fi
        STOPPED=1
    fi
    rm -f "$PID_FILE"
fi

# 2. Quét thêm cổng 65534 bằng ss nếu PID file không khớp
PORT_PID=$(ss -tulpn 2>/dev/null | grep ":$PORT " | grep -o 'pid=[0-9]*' | head -n 1 | cut -d= -f2)
if [ -z "$PORT_PID" ]; then
    PORT_PID=$(lsof -ti :$PORT 2>/dev/null)
fi

if [ -n "$PORT_PID" ]; then
    echo "➡️  Dừng tiến trình đang chiếm cổng $PORT (PID: $PORT_PID)..."
    kill "$PORT_PID" 2>/dev/null
    sleep 1
    if kill -0 "$PORT_PID" 2>/dev/null; then
        kill -9 "$PORT_PID" 2>/dev/null
    fi
    STOPPED=1
fi

if [ $STOPPED -eq 1 ]; then
    echo "✅ Đã dừng thành công dịch vụ Tailscale Port Router!"
else
    echo "ℹ️  Không có tiến trình nào đang chạy trên cổng $PORT."
fi
