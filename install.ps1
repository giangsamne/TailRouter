# ==============================================================================
# ⚡ TailRouter - Windows Smart Installer
# Auto-downloads and installs TailRouter Native Engine, creates shortcut & autostart.
# ==============================================================================

param(
    [string]$Tag = "v2.3.0"
)

$ErrorActionPreference = "Stop"
[Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12

$Repo = "giangsamne/TailRouter"
$ReleaseTag = $Tag
$Arch = "amd64"
if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
    $Arch = "arm64"
}
$ArchiveName = "tailrouter-windows-$Arch.zip"
$DownloadUrl = "https://github.com/$Repo/releases/download/$ReleaseTag/$ArchiveName"
$InstallDir = "$env:LOCALAPPDATA\TailRouter"
$RandomSuffix = [System.IO.Path]::GetRandomFileName()
$TempDir = Join-Path $env:TEMP "tailrouter_$RandomSuffix"
New-Item -ItemType Directory -Path $TempDir -Force | Out-Null
$TempZip = Join-Path $TempDir "$ArchiveName"

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "⚡ TailRouter Windows Installer ($ReleaseTag)" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

# 1. Stop running processes
Write-Host "==> [1/4] Stopping existing TailRouter processes if any..." -ForegroundColor Yellow
Stop-Process -Name "tailrouter" -ErrorAction SilentlyContinue
Stop-Process -Name "TailRouter" -ErrorAction SilentlyContinue
Start-Sleep -Seconds 1

# 2. Download release
Write-Host "==> [2/4] Downloading $ArchiveName from GitHub Releases..." -ForegroundColor Yellow
try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $TempZip -UseBasicParsing
} catch {
    # Fallback to legacy bundle archive
    $ArchiveName = "TailRouter-Windows.zip"
    $DownloadUrl = "https://github.com/$Repo/releases/download/$ReleaseTag/$ArchiveName"
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $TempZip -UseBasicParsing
}

# 3. Extract and install
Write-Host "==> [3/4] Installing to $InstallDir..." -ForegroundColor Yellow
if (!(Test-Path -Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}
Expand-Archive -Path $TempZip -DestinationPath $TempDir\extract -Force
if (Test-Path "$TempDir\extract\tailrouter-windows-$Arch.exe") {
    Move-Item -Path "$TempDir\extract\tailrouter-windows-$Arch.exe" -Destination "$InstallDir\tailrouter.exe" -Force
} elseif (Test-Path "$TempDir\extract\tailrouter.exe") {
    Move-Item -Path "$TempDir\extract\tailrouter.exe" -Destination "$InstallDir\tailrouter.exe" -Force
}
Remove-Item -Path $TempDir -Recurse -Force -ErrorAction SilentlyContinue

# 4. Create Desktop Shortcut and Startup Registry
Write-Host "==> [4/4] Setting up PATH and Autostart..." -ForegroundColor Yellow
$WshShell = New-Object -ComObject WScript.Shell
$DesktopPath = [Environment]::GetFolderPath("Desktop")
$Shortcut = $WshShell.CreateShortcut("$DesktopPath\TailRouter.lnk")
$Shortcut.TargetPath = "$InstallDir\tailrouter.exe"
$Shortcut.Arguments = ""
$Shortcut.WorkingDirectory = $InstallDir
$Shortcut.Description = "TailRouter - High-Performance Bare-Metal Gateway"
$Shortcut.Save()

# Autostart on Windows Logon
$StartupKey = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Run"
Set-ItemProperty -Path $StartupKey -Name "TailRouter" -Value "`"$InstallDir\tailrouter.exe`" run" -Force

# Add InstallDir to User PATH if not present
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
    Write-Host "✅ Added $InstallDir to User PATH." -ForegroundColor Yellow
}

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host " 🎉 Installation Complete!                                " -ForegroundColor Green
Write-Host " 🚀 Starting TailRouter Gateway in background...          " -ForegroundColor Green
Write-Host "==========================================================" -ForegroundColor Cyan

# 5. Start TailRouter in background
Start-Process -FilePath "$InstallDir\tailrouter.exe" -ArgumentList "run" -WindowStyle Hidden
Start-Sleep -Seconds 2

Write-Host "💡 Web Dashboard accessible at: http://localhost:65534" -ForegroundColor Cyan
Start-Process "http://localhost:65534"
