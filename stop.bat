@echo off
REM ============================================================================
REM TailRouter - Windows Stop Script
REM ============================================================================
cd /d "%~dp0"
set PORT=65534

echo Stopping TailRouter on port %PORT%...

for /f "tokens=5" %%a in ('netstat -aon ^| findstr ":%PORT% " ^| findstr "LISTENING"') do (
    echo Terminating PID: %%a
    taskkill /F /PID %%a >nul 2>nul
)

if exist server.pid (
    set /p PID=<server.pid
    taskkill /F /PID %PID% >nul 2>nul
    del /f /q server.pid
)

echo [OK] Service stopped.
pause
