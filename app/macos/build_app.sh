#!/bin/bash
set -e

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$DIR/../.." && pwd)"
APP_NAME="TailRouter.app"
DIST_DIR="$ROOT_DIR/dist"
APP_DIR="$DIST_DIR/$APP_NAME"

echo "🍏 Đang đóng gói $APP_NAME cho macOS..."

# 1. Tạo cấu trúc thư mục .app bundle
rm -rf "$APP_DIR"
mkdir -p "$APP_DIR/Contents/MacOS"
mkdir -p "$APP_DIR/Contents/Resources"

# 2. Biên dịch Native Swift Menu Bar App
echo "🔨 Đang biên dịch TailRouterMenuApp.swift..."
swiftc -O "$DIR/TailRouterMenuApp.swift" -o "$APP_DIR/Contents/MacOS/TailRouter"

# 3. Tạo Info.plist
cat << PLIST_EOF > "$APP_DIR/Contents/Info.plist"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleName</key>
    <string>TailRouter</string>
    <key>CFBundleDisplayName</key>
    <string>TailRouter</string>
    <key>CFBundleIdentifier</key>
    <string>com.tailrouter.app</string>
    <key>CFBundleVersion</key>
    <string>1.0.0</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0.0</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleExecutable</key>
    <string>TailRouter</string>
    <key>LSMinimumSystemVersion</key>
    <string>11.0</string>
    <key>LSUIElement</key>
    <true/>
    <key>NSHighResolutionCapable</key>
    <true/>
</dict>
</plist>
PLIST_EOF

# 4. Sao chép toàn bộ mã nguồn server và giao diện vào Resources
echo "📦 Đóng gói tài nguyên Core Server & Web UI..."
cp "$ROOT_DIR/server.py" "$APP_DIR/Contents/Resources/"
cp "$ROOT_DIR/routes_manager.py" "$APP_DIR/Contents/Resources/"
cp "$ROOT_DIR/tailscale_helper.py" "$APP_DIR/Contents/Resources/"
cp "$ROOT_DIR/service_helper.py" "$APP_DIR/Contents/Resources/"
cp "$ROOT_DIR/port_scanner.py" "$APP_DIR/Contents/Resources/"
cp -R "$ROOT_DIR/web" "$APP_DIR/Contents/Resources/"
mkdir -p "$APP_DIR/Contents/Resources/config"

chmod +x "$APP_DIR/Contents/MacOS/TailRouter"

echo "============================================================"
echo "🎉 Đã đóng gói thành công: $APP_DIR"
echo "👉 Để cài đặt vào máy Mac: cp -R \"$APP_DIR\" /Applications/"
echo "👉 Hoặc mở trực tiếp: open \"$APP_DIR\""
echo "============================================================"
