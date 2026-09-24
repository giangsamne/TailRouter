#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
APP_NAME="TailRouter"
BUILD_DIR="$ROOT_DIR/dist/macos"
APP_BUNDLE="$BUILD_DIR/$APP_NAME.app"

echo "==> Đang đóng gói $APP_NAME.app (Native Go Engine & Swift Menu Bar)..."

rm -rf "$APP_BUNDLE"
mkdir -p "$APP_BUNDLE/Contents/MacOS"
mkdir -p "$APP_BUNDLE/Contents/Resources"

# 1. Biên dịch Swift Menu Bar App
echo "==> Đang biên dịch Swift Menu Bar status item..."
swiftc "$SCRIPT_DIR/TailRouterMenuApp.swift" -o "$APP_BUNDLE/Contents/MacOS/$APP_NAME" -target arm64-apple-macos11.0

# 2. Copy Native Go Engine vào bundle
echo "==> Đang tích hợp Go Engine..."
if [ -f "$BUILD_DIR/tailrouter-server-arm64" ]; then
    cp "$BUILD_DIR/tailrouter-server-arm64" "$APP_BUNDLE/Contents/MacOS/tailrouter-server"
    chmod +x "$APP_BUNDLE/Contents/MacOS/tailrouter-server"
elif [ -f "$ROOT_DIR/dist/macos/tailrouter-server" ]; then
    cp "$ROOT_DIR/dist/macos/tailrouter-server" "$APP_BUNDLE/Contents/MacOS/tailrouter-server"
    chmod +x "$APP_BUNDLE/Contents/MacOS/tailrouter-server"
fi

# 3. Tạo Info.plist với LSUIElement = true (ẩn Dock icon)
cat << 'PLIST' > "$APP_BUNDLE/Contents/Info.plist"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>TailRouter</string>
    <key>CFBundleIdentifier</key>
    <string>com.tailrouter.menubar</string>
    <key>CFBundleName</key>
    <string>TailRouter</string>
    <key>CFBundleDisplayName</key>
    <string>TailRouter</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>2.0.0</string>
    <key>CFBundleVersion</key>
    <string>2.0.0</string>
    <key>LSMinimumSystemVersion</key>
    <string>11.0</string>
    <key>LSUIElement</key>
    <true/>
    <key>NSHighResolutionCapable</key>
    <true/>
</dict>
</plist>
PLIST

echo "==> Đóng gói hoàn tất tại: $APP_BUNDLE"
