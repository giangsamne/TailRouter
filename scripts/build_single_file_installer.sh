#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
RELEASE_DIR="$ROOT_DIR/release"
OUTPUT_FILE="$RELEASE_DIR/TailRouter-Setup.cmd"

echo "=========================================================="
echo "⚡ Packaging 1-File Universal Offline Installer: $OUTPUT_FILE"
echo "=========================================================="

TMP_DIR="/tmp/tailrouter_payload_pack_$$"
rm -rf "$TMP_DIR"
mkdir -p "$TMP_DIR"

cp "$RELEASE_DIR/TailRouter-Windows.zip" "$TMP_DIR/"
cp "$RELEASE_DIR/TailRouter-macOS.zip" "$TMP_DIR/"
cp "$RELEASE_DIR/TailRouter-Linux.tar.gz" "$TMP_DIR/"

PAYLOAD_ZIP="/tmp/tailrouter_payload_bundle_$$.zip"
rm -f "$PAYLOAD_ZIP"
cd "$TMP_DIR"
zip -q -0 -r "$PAYLOAD_ZIP" .
rm -rf "$TMP_DIR"

HEADER_FILE="/tmp/tailrouter_header_$$.cmd"
cat << 'POLYGLOT_HEADER' > "$HEADER_FILE"
:; # ==============================================================================
:; # ⚡ TailRouter - 1-File Universal Offline Installer (Linux, macOS, Windows)
:; # Carry this SINGLE file on a USB drive and run on any operating system!
:; # On Linux & macOS: Run with 'sh TailRouter-Setup.cmd' or './TailRouter-Setup.cmd'
:; # On Windows:       Double-click 'TailRouter-Setup.cmd' in File Explorer
:; # ==============================================================================
:; set -e
:; OS="$(uname -s)"
:; ARCH="$(uname -m)"
:; echo "=========================================================="
:; echo "⚡ TailRouter - Universal 1-File Offline Installer"
:; echo "=========================================================="
:; echo "==> Detected OS: $OS ($ARCH)"
:; 
:; TMP_WORK="/tmp/tailrouter_1file_$$"
:; mkdir -p "$TMP_WORK"
:; PKG_BUNDLE="$TMP_WORK/bundle.zip"
:; 
:; echo "==> [1/3] Extracting embedded packages..."
:; MARKER_LINE=$(grep -a -n -m 1 '^__ARCHIVE_DATA__$' "$0" | cut -d: -f1)
:; if [ -n "$MARKER_LINE" ]; then
:;   tail -n +$((MARKER_LINE + 1)) "$0" > "$PKG_BUNDLE"
:; else
:;   cp "$0" "$PKG_BUNDLE"
:; fi
:; 
:; unzip -q -o "$PKG_BUNDLE" -d "$TMP_WORK"
:; rm -f "$PKG_BUNDLE"
:; 
:; if [ "$OS" = "Darwin" ]; then
:;   echo "==> [2/3] Installing TailRouter on macOS..."
:;   INSTALL_DIR="/Applications"
:;   APP_NAME="TailRouter.app"
:;   unzip -q -o "$TMP_WORK/TailRouter-macOS.zip" -d "$TMP_WORK"
:;   pkill -f "TailRouter.app" 2>/dev/null || true
:;   pkill -f "tailrouter-server" 2>/dev/null || true
:;   sleep 1
:;   cp -R "$TMP_WORK/$APP_NAME" "$INSTALL_DIR/"
:;   xattr -dr com.apple.quarantine "$INSTALL_DIR/$APP_NAME" 2>/dev/null || true
:;   rm -rf "$TMP_WORK"
:;   echo "==> [3/3] Launching TailRouter..."
:;   open "$INSTALL_DIR/$APP_NAME"
:;   echo "=========================================================="
:;   echo "✅ TailRouter installed successfully to $INSTALL_DIR/$APP_NAME!"
:;   echo "💡 TailRouter is running in your macOS Menu Bar."
:;   echo "=========================================================="
:; elif [ "$OS" = "Linux" ]; then
:;   echo "==> [2/3] Installing TailRouter on Linux..."
:;   tar -xzf "$TMP_WORK/TailRouter-Linux.tar.gz" -C "$TMP_WORK"
:;   BIN_NAME="tailrouter-server"
:;   if [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
:;     BIN_NAME="tailrouter-server-arm64"
:;   fi
:;   TARGET_BIN="/usr/local/bin/tailrouter"
:;   if [ -w "/usr/local/bin" ] || [ "$(id -u)" -eq 0 ]; then
:;     cp "$TMP_WORK/$BIN_NAME" "$TARGET_BIN"
:;     chmod +x "$TARGET_BIN"
:;   elif command -v sudo >/dev/null 2>&1 && sudo -n true 2>/dev/null; then
:;     sudo cp "$TMP_WORK/$BIN_NAME" "$TARGET_BIN"
:;     sudo chmod +x "$TARGET_BIN"
:;   else
:;     TARGET_BIN="$HOME/.local/bin/tailrouter"
:;     mkdir -p "$HOME/.local/bin"
:;     cp "$TMP_WORK/$BIN_NAME" "$TARGET_BIN"
:;     chmod +x "$TARGET_BIN"
:;     if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
:;       export PATH="$HOME/.local/bin:$PATH"
:;     fi
:;   fi
:;   rm -rf "$TMP_WORK"
:;   echo "==> [3/3] Setting up system service..."
:;   "$TARGET_BIN" service install 2>/dev/null || true
:;   echo "=========================================================="
:;   echo "✅ TailRouter installed successfully at: $TARGET_BIN"
:;   echo "==> Gateway status:"
:;   "$TARGET_BIN" status || true
:;   echo "=========================================================="
:;   echo "💡 Web Dashboard: http://localhost:65534/router"
:; else
:;   echo "❌ Unsupported OS: $OS"
:;   exit 1
:; fi
:; exit 0

@echo off
setlocal enabledelayedexpansion
title TailRouter - 1-File Universal Offline Installer
echo ==========================================================
echo ⚡ TailRouter - 1-File Universal Offline Installer (Windows)
echo ==========================================================
echo ==> [1/3] Dang trich xuat goi cai dat...

set THIS_FILE=%~f0
set INSTALL_DIR=%LOCALAPPDATA%\TailRouter
set TEMP_ZIP=%TEMP%\tailrouter_bundle_%RANDOM%.zip

powershell -NoProfile -ExecutionPolicy Bypass -Command ^
    "$bytes = [System.IO.File]::ReadAllBytes('%THIS_FILE%');" ^
    "$found = -1;" ^
    "for ($i = 0; $i -lt 8192; $i++) {" ^
    "    if ($bytes[$i] -eq 0x50 -and $bytes[$i+1] -eq 0x4B -and $bytes[$i+2] -eq 0x03 -and $bytes[$i+3] -eq 0x04) {" ^
    "        $found = $i; break;" ^
    "    }" ^
    "};" ^
    "if ($found -ge 0) {" ^
    "    $zipBytes = New-Object byte[] ($bytes.Length - $found);" ^
    "    [Array]::Copy($bytes, $found, $zipBytes, 0, $zipBytes.Length);" ^
    "    [System.IO.File]::WriteAllBytes('%TEMP_ZIP%', $zipBytes);" ^
    "    $tmpDir = [System.IO.Path]::Combine($env:TEMP, 'tailrouter_unpacked_' + [System.Guid]::NewGuid().ToString());" ^
    "    Expand-Archive -Path '%TEMP_ZIP%' -DestinationPath $tmpDir -Force;" ^
    "    Remove-Item -Path '%TEMP_ZIP%' -Force;" ^
    "    if (Test-Path \"$tmpDir\TailRouter-Windows.zip\") {" ^
    "        if (!(Test-Path '%INSTALL_DIR%')) { New-Item -ItemType Directory -Path '%INSTALL_DIR%' -Force | Out-Null };" ^
    "        Expand-Archive -Path \"$tmpDir\TailRouter-Windows.zip\" -DestinationPath '%INSTALL_DIR%' -Force;" ^
    "    };" ^
    "    Remove-Item -Path $tmpDir -Recurse -Force;" ^
    "} else { Write-Error 'Cannot find embedded zip payload'; exit 1 }"

if errorlevel 1 (
    echo [ERROR] Trich xuat du lieu that bai!
    pause
    exit /b 1
)

echo ==> [2/3] Dung tien trinh cu neu co...
taskkill /f /im TailRouter.exe 2>nul
taskkill /f /im tailrouter-server.exe 2>nul
timeout /t 1 /nobreak >nul

echo ==> [3/3] Tao Shortcut Desktop va Autostart...
powershell -NoProfile -ExecutionPolicy Bypass -Command ^
    "$ws = New-Object -ComObject WScript.Shell;" ^
    "$s = $ws.CreateShortcut([Environment]::GetFolderPath('Desktop') + '\TailRouter.lnk');" ^
    "$s.TargetPath = '%INSTALL_DIR%\TailRouter.exe';" ^
    "$s.WorkingDirectory = '%INSTALL_DIR%';" ^
    "$s.Description = 'TailRouter - Tailscale Port Router Gateway';" ^
    "$s.Save()"

reg add "HKCU\Software\Microsoft\Windows\CurrentVersion\Run" /v "TailRouter" /t REG_SZ /d "\"%INSTALL_DIR%\TailRouter.exe\"" /f >nul

echo ==========================================================
echo [SUCCESS] Cai dat TailRouter thanh cong vao %INSTALL_DIR%!
echo ==> Dang khoi dong TailRouter...
start "" "%INSTALL_DIR%\TailRouter.exe"
echo Bieu tuong app da xuat hien duoi khay he thong (System Tray).
echo Bang dieu khien Web: http://localhost:65534/router
echo ==========================================================
pause
exit /b 0
__ARCHIVE_DATA__
POLYGLOT_HEADER

cat "$HEADER_FILE" "$PAYLOAD_ZIP" > "$OUTPUT_FILE"
chmod +x "$OUTPUT_FILE"

rm -f "$HEADER_FILE" "$PAYLOAD_ZIP"

echo "=========================================================="
echo "✅ 1-File Universal Offline Installer built successfully!"
echo "   File: $OUTPUT_FILE"
echo "   Size: $(du -h "$OUTPUT_FILE" | cut -f1)"
echo "=========================================================="
