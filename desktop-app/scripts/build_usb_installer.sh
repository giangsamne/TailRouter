#!/bin/bash
set -e

# ==============================================================================
# ⚡ TailRouter - Build Offline USB Installer Package
# Bundles all offline platform binaries and cross-platform smart launchers.
# ==============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
RELEASE_DIR="$ROOT_DIR/release"
BUILD_DIR="$RELEASE_DIR/TailRouter-USB-Installer"
OUTPUT_ZIP="$RELEASE_DIR/TailRouter-USB-Offline-Installer.zip"

echo "=========================================================="
echo "⚡ Packaging TailRouter Offline USB Installer"
echo "=========================================================="

# Ensure release binaries exist
if [ ! -f "$RELEASE_DIR/TailRouter-macOS.zip" ] || \
   [ ! -f "$RELEASE_DIR/TailRouter-Windows.zip" ] || \
   [ ! -f "$RELEASE_DIR/TailRouter-Linux.tar.gz" ]; then
  echo "==> Compiling latest binaries first..."
  "$SCRIPT_DIR/build_all.sh"
fi

rm -rf "$BUILD_DIR" "$OUTPUT_ZIP"
mkdir -p "$BUILD_DIR/packages"

echo "==> [1/4] Copying offline packages for all 3 OS into USB bundle..."
cp "$RELEASE_DIR/TailRouter-macOS.zip" "$BUILD_DIR/packages/"
cp "$RELEASE_DIR/TailRouter-Windows.zip" "$BUILD_DIR/packages/"
cp "$RELEASE_DIR/TailRouter-Linux.tar.gz" "$BUILD_DIR/packages/"

echo "==> [2/4] Generating Universal Smart Launchers..."

# 1. Setup.sh (Linux & macOS terminal launcher)
cat << 'LAUNCHER_SH' > "$BUILD_DIR/Setup.sh"
#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OS="$(uname -s)"
ARCH="$(uname -m)"

echo "=========================================================="
echo "⚡ TailRouter - USB Offline Installer"
echo "=========================================================="
echo "==> Detected OS: $OS ($ARCH)"

case "$OS" in
  Darwin)
    echo "==> [1/3] Preparing macOS installation..."
    PKG_ZIP="$SCRIPT_DIR/packages/TailRouter-macOS.zip"
    APP_NAME="TailRouter.app"
    INSTALL_DIR="/Applications"

    if [ ! -f "$PKG_ZIP" ]; then
      echo "❌ Error: $PKG_ZIP not found on USB!"
      exit 1
    fi

    TMP_DIR="/tmp/tailrouter_usb_$$"
    mkdir -p "$TMP_DIR"
    echo "==> [2/3] Extracting $APP_NAME from USB package..."
    unzip -q -o "$PKG_ZIP" -d "$TMP_DIR"

    pkill -f "TailRouter.app" 2>/dev/null || true
    pkill -f "tailrouter-server" 2>/dev/null || true
    sleep 1

    echo "==> [3/3] Installing $APP_NAME into $INSTALL_DIR..."
    cp -R "$TMP_DIR/$APP_NAME" "$INSTALL_DIR/"
    xattr -dr com.apple.quarantine "$INSTALL_DIR/$APP_NAME" 2>/dev/null || true
    rm -rf "$TMP_DIR"

    echo "=========================================================="
    echo "✅ TailRouter installed successfully from USB to $INSTALL_DIR/$APP_NAME!"
    echo "==> Launching TailRouter..."
    open "$INSTALL_DIR/$APP_NAME"
    echo "💡 TailRouter is now running in your macOS Menu Bar."
    echo "=========================================================="
    ;;

  Linux)
    echo "==> [1/3] Preparing Linux installation..."
    PKG_TAR="$SCRIPT_DIR/packages/TailRouter-Linux.tar.gz"

    if [ ! -f "$PKG_TAR" ]; then
      echo "❌ Error: $PKG_TAR not found on USB!"
      exit 1
    fi

    TMP_DIR="/tmp/tailrouter_usb_$$"
    mkdir -p "$TMP_DIR"
    echo "==> [2/3] Extracting Linux binaries from USB package..."
    tar -xzf "$PKG_TAR" -C "$TMP_DIR"

    BIN_NAME="tailrouter-server"
    if [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
      BIN_NAME="tailrouter-server-arm64"
    fi

    TARGET_BIN="/usr/local/bin/tailrouter"
    if [ -w "/usr/local/bin" ] || [ "$(id -u)" -eq 0 ]; then
      cp "$TMP_DIR/$BIN_NAME" "$TARGET_BIN"
      chmod +x "$TARGET_BIN"
    elif command -v sudo >/dev/null 2>&1 && sudo -n true 2>/dev/null; then
      sudo cp "$TMP_DIR/$BIN_NAME" "$TARGET_BIN"
      sudo chmod +x "$TARGET_BIN"
    else
      TARGET_BIN="$HOME/.local/bin/tailrouter"
      mkdir -p "$HOME/.local/bin"
      cp "$TMP_DIR/$BIN_NAME" "$TARGET_BIN"
      chmod +x "$TARGET_BIN"
      if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
        export PATH="$HOME/.local/bin:$PATH"
      fi
    fi

    rm -rf "$TMP_DIR"

    echo "==> [3/3] Setting up system service..."
    "$TARGET_BIN" service install 2>/dev/null || true

    echo "=========================================================="
    echo "✅ TailRouter installed successfully from USB at: $TARGET_BIN"
    echo "==> Gateway status:"
    "$TARGET_BIN" status || true
    echo "=========================================================="
    echo "💡 Web Dashboard accessible at: http://localhost:65534/router"
    ;;

  *)
    echo "❌ Unsupported OS: $OS"
    echo "If on Windows, please double-click Setup.cmd"
    exit 1
    ;;
esac
LAUNCHER_SH
chmod +x "$BUILD_DIR/Setup.sh"

# 2. Setup.command (macOS Finder double-clickable launcher)
cat << 'LAUNCHER_CMD' > "$BUILD_DIR/Setup.command"
#!/bin/bash
DIR="$(cd "$(dirname "$0")" && pwd)"
bash "$DIR/Setup.sh"
echo ""
echo "Nhấn phím bất kỳ để đóng cửa sổ này / Press any key to close..."
read -n 1
LAUNCHER_CMD
chmod +x "$BUILD_DIR/Setup.command"

# 3. Setup.cmd (Polyglot: Double click on Windows, or run in Linux/macOS)
cat << 'LAUNCHER_POLY' > "$BUILD_DIR/Setup.cmd"
:; # ==============================================================================
:; # ⚡ TailRouter - USB Offline Universal Launcher
:; # On Linux/macOS: Run with 'sh Setup.cmd' or './Setup.cmd'
:; # On Windows:     Double-click 'Setup.cmd'
:; # ==============================================================================
:; SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
:; if [ -f "$SCRIPT_DIR/Setup.sh" ]; then
:;   bash "$SCRIPT_DIR/Setup.sh" "$@"
:; fi
:; exit 0

@echo off
setlocal
title TailRouter - USB Offline Installer
echo ==========================================================
echo ⚡ TailRouter - USB Offline Installer (Windows)
echo ==========================================================
set SCRIPT_DIR=%~dp0
set PKG_ZIP=%SCRIPT_DIR%packages\TailRouter-Windows.zip
set INSTALL_DIR=%LOCALAPPDATA%\TailRouter

if not exist "%PKG_ZIP%" (
    echo [ERROR] Khong tim thay goi %PKG_ZIP% tren USB!
    pause
    exit /b 1
)

echo ==> [1/3] Dung cac tien trinh TailRouter neu co...
taskkill /f /im TailRouter.exe 2>nul
taskkill /f /im tailrouter-server.exe 2>nul
timeout /t 1 /nobreak >nul

echo ==> [2/3] Dang giai nen tu USB vao %INSTALL_DIR%...
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"
powershell -NoProfile -ExecutionPolicy Bypass -Command "Expand-Archive -Path '%PKG_ZIP%' -DestinationPath '%INSTALL_DIR%' -Force"

echo ==> [3/3] Dang tao Shortcut Desktop va cai dat Tu Khoi Dong...
powershell -NoProfile -ExecutionPolicy Bypass -Command "$ws = New-Object -ComObject WScript.Shell; $s = $ws.CreateShortcut([Environment]::GetFolderPath('Desktop') + '\TailRouter.lnk'); $s.TargetPath = '%INSTALL_DIR%\TailRouter.exe'; $s.WorkingDirectory = '%INSTALL_DIR%'; $s.Save()"
reg add "HKCU\Software\Microsoft\Windows\CurrentVersion\Run" /v "TailRouter" /t REG_SZ /d "\"%INSTALL_DIR%\TailRouter.exe\"" /f >nul

echo ==========================================================
echo [SUCCESS] Cai dat TailRouter tu USB thanh cong!
echo ==> Dang khoi dong TailRouter...
start "" "%INSTALL_DIR%\TailRouter.exe"
echo Bieu tuong app da xuat hien duoi khay he thong (System Tray).
echo Bang dieu khien Web: http://localhost:65534/router
echo ==========================================================
pause
exit /b 0
LAUNCHER_POLY
chmod +x "$BUILD_DIR/Setup.cmd"

echo "==> [3/4] Creating Quick Guide file..."
cat << 'GUIDE_TXT' > "$BUILD_DIR/Hướng-Dẫn-Cài-Đặt.txt"
================================================================================
⚡ BỘ CÀI ĐẶT TAILROUTER ĐA NỀN TẢNG QUA USB (OFFLINE INSTALLER)
Không cần kết nối Internet - Tự động nhận diện hệ điều hành và cài đặt
================================================================================

📦 HƯỚNG DẪN CÀI ĐẶT:

1. 🪟 TRÊN WINDOWS:
   - Mở ổ USB trong File Explorer.
   - Nhấp đúp chuột (Double Click) vào file:
     👉 Setup.cmd
   - Bộ cài sẽ tự động giải nén, tạo biểu tượng Desktop và khởi chạy TailRouter.exe.

2. 🍏 TRÊN MACOS:
   - Mở ổ USB trong Finder.
   - Nhấp đúp chuột (Double Click) vào file:
     👉 Setup.command
   - Bộ cài sẽ tự động cài TailRouter.app vào thư mục /Applications và mở app.

3. 🐧 TRÊN LINUX (Ubuntu, Debian, Fedora, Alpine Linux, Raspberry Pi):
   - Mở Terminal tại thư mục USB.
   - Chạy lệnh:
     👉 ./Setup.sh
     (hoặc: bash Setup.sh)
   - Bộ cài sẽ tự nhận diện kiến trúc CPU (x86_64 hoặc ARM64) và hệ thống init
     (systemd hoặc openrc) để cài đặt và kích hoạt dịch vụ chạy ngầm.

================================================================================
🌐 BẢNG ĐIỀU KHIỂN SAU KHI CÀI ĐẶT:
   Truy cập: http://localhost:65534/router
================================================================================
GUIDE_TXT

echo "==> [4/4] Compressing into standalone archive: $OUTPUT_ZIP..."
cd "$RELEASE_DIR"
zip -r -q "$OUTPUT_ZIP" "TailRouter-USB-Installer"
rm -rf "$BUILD_DIR"

echo "=========================================================="
echo "✅ USB Offline Installer package created successfully!"
echo "   File: $OUTPUT_ZIP"
echo "   Size: $(du -h "$OUTPUT_ZIP" | cut -f1)"
echo "=========================================================="
