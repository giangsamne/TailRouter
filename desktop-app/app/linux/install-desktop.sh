#!/bin/bash
set -e

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$DIR/../.." && pwd)"

echo "🐧 Đang cài đặt TailRouter Desktop Launcher cho Linux..."

# Cài đặt icon
ICON_DIR="$HOME/.local/share/icons/hicolor/scalable/apps"
mkdir -p "$ICON_DIR"
cp "$ROOT_DIR/app/assets/icon.svg" "$ICON_DIR/tailrouter.svg"

# Cài đặt file .desktop
APPS_DIR="$HOME/.local/share/applications"
mkdir -p "$APPS_DIR"
cp "$DIR/tailrouter.desktop" "$APPS_DIR/tailrouter.desktop"

# Cập nhật desktop database
update-desktop-database "$APPS_DIR" 2>/dev/null || true

echo "============================================================"
echo "🎉 Đã cài đặt thành công TailRouter vào menu ứng dụng Linux!"
echo "👉 Bạn có thể tìm thấy 'TailRouter' trong App Launcher / Dash."
echo "============================================================"
