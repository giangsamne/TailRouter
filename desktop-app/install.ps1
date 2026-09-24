# ==============================================================================
# ⚡ TailRouter - Windows Smart Installer
# Auto-downloads and installs TailRouter Tray & Server, creates shortcut & autostart.
# ==============================================================================

$Repo = "giangsamne/TailRouter"
$ReleaseTag = "v2.0.0"
$ArchiveName = "TailRouter-Windows.zip"
$DownloadUrl = "https://github.com/$Repo/releases/download/$ReleaseTag/$ArchiveName"
$InstallDir = "$env:LOCALAPPDATA\TailRouter"
$TempZip = "$env:TEMP\TailRouter-Windows.zip"

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "⚡ TailRouter Windows Installer ($ReleaseTag)" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

# 1. Stop running processes
Write-Host "==> [1/4] Stopping existing TailRouter processes if any..." -ForegroundColor Yellow
Stop-Process -Name "TailRouter" -ErrorAction SilentlyContinue
Stop-Process -Name "tailrouter-server" -ErrorAction SilentlyContinue
Start-Sleep -Seconds 1

# 2. Download release
Write-Host "==> [2/4] Downloading $ArchiveName from GitHub Releases..." -ForegroundColor Yellow
Invoke-WebRequest -Uri $DownloadUrl -OutFile $TempZip -UseBasicParsing

# 3. Extract and install
Write-Host "==> [3/4] Installing to $InstallDir..." -ForegroundColor Yellow
if (!(Test-Path -Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}
Expand-Archive -Path $TempZip -DestinationPath $InstallDir -Force
Remove-Item -Path $TempZip -Force

# 4. Create Desktop Shortcut and Startup Registry
Write-Host "==> [4/4] Creating Desktop shortcut and setting up Autostart..." -ForegroundColor Yellow
$WshShell = New-Object -ComObject WScript.Shell
$DesktopPath = [Environment]::GetFolderPath("Desktop")
$Shortcut = $WshShell.CreateShortcut("$DesktopPath\TailRouter.lnk")
$Shortcut.TargetPath = "$InstallDir\TailRouter.exe"
$Shortcut.WorkingDirectory = $InstallDir
$Shortcut.Description = "TailRouter - Tailscale Port Router Gateway"
$Shortcut.Save()

# Autostart on Windows Logon
$StartupKey = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Run"
Set-ItemProperty -Path $StartupKey -Name "TailRouter" -Value "`"$InstallDir\TailRouter.exe`"" -Force

# Add InstallDir to User PATH if not present
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
}

Write-Host "==========================================================" -ForegroundColor Green
Write-Host "✅ TailRouter installed successfully to $InstallDir!" -ForegroundColor Green
Write-Host "==> Starting TailRouter Tray App..." -ForegroundColor Cyan
Start-Process -FilePath "$InstallDir\TailRouter.exe"
Write-Host "💡 Web Dashboard accessible at: http://localhost:65534/router" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Green
