@echo off
if exist "%~dp0main\server.py" (
  cd /d "%~dp0main"
  python server.py %*
) else (
  python "%~dp0server.py" %*
)
