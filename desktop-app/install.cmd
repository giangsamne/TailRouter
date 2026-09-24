:; # ==============================================================================
:; # ⚡ TailRouter - Universal Polyglot Installer (Linux, macOS, Windows)
:; # On Linux/macOS: Run with 'sh install.cmd' or './install.cmd'
:; # On Windows:     Double-click 'install.cmd' or run from Command Prompt
:; # ==============================================================================
:; SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
:; if [ -f "$SCRIPT_DIR/install.sh" ]; then
:;   bash "$SCRIPT_DIR/install.sh" "$@"
:; else
:;   curl -fsSL https://raw.githubusercontent.com/giangsamne/TailRouter/desktop-app/install.sh | bash
:; fi
:; exit 0

@echo off
setlocal
title TailRouter Universal Installer
echo ==========================================================
echo ⚡ TailRouter Universal Installer (Windows)
echo ==========================================================
set SCRIPT_DIR=%~dp0
if exist "%SCRIPT_DIR%install.ps1" (
    powershell -NoProfile -ExecutionPolicy Bypass -File "%SCRIPT_DIR%install.ps1"
) else (
    powershell -NoProfile -ExecutionPolicy Bypass -Command "irm https://raw.githubusercontent.com/giangsamne/TailRouter/desktop-app/install.ps1 | iex"
)
pause
exit /b 0
