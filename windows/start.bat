@echo off
REM ============================================================================
REM TailRouter - Windows Launcher (Port 65534)
REM ============================================================================
title TailRouter (Port 65534)
cd /d "%~dp0"

echo ============================================================
echo   TailRouter ^& Gateway (Port 65534)
echo   Windows Edition
echo ============================================================

where python >nul 2>nul
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Python 3 is not found in PATH!
    echo Please install Python from https://www.python.org/ or Microsoft Store.
    pause
    exit /b 1
)

echo Starting TailRouter...
echo Access Web Dashboard: http://localhost:65534/router
echo.
python server.py
pause
