# TailRouter Native Windows System Tray Agent
Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName System.Drawing

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RootDir = Split-Path -Parent (Split-Path -Parent $ScriptDir)
$ServerPy = Join-Path $RootDir "server.py"

# 1. Khởi động server.py nếu chưa chạy
$existing = Get-Process python -ErrorAction SilentlyContinue | Where-Object { $_.CommandLine -like "*server.py*" }
if (-not $existing) {
    Start-Process python -ArgumentList "`"$ServerPy`"" -WorkingDirectory "$RootDir" -WindowStyle Hidden
}

# 2. Tạo NotifyIcon trên khay hệ thống Windows
$notify = New-Object System.Windows.Forms.NotifyIcon
$notify.Icon = [System.Drawing.SystemIcons]::Application
$notify.Text = "TailRouter - Tailscale Port Gateway (Port 65534)"
$notify.Visible = $true

$contextMenu = New-Object System.Windows.Forms.ContextMenuStrip

$headerItem = $contextMenu.Items.Add("🟢 TailRouter: Đang chạy (Cổng 65534)")
$headerItem.Enabled = $false

$contextMenu.Items.Add("-") | Out-Null

$openItem = $contextMenu.Items.Add("🚀 Mở Bảng Quản Trị (/router)")
$openItem.Add_Click({
    Start-Process "http://localhost:65534/router"
})

$restartItem = $contextMenu.Items.Add("🔄 Khởi động lại Gateway")
$restartItem.Add_Click({
    Get-Process python -ErrorAction SilentlyContinue | Where-Object { $_.CommandLine -like "*server.py*" } | Stop-Process -Force
    Start-Sleep -Seconds 1
    Start-Process python -ArgumentList "`"$ServerPy`"" -WorkingDirectory "$RootDir" -WindowStyle Hidden
})

$contextMenu.Items.Add("-") | Out-Null

$exitItem = $contextMenu.Items.Add("❌ Thoát TailRouter")
$exitItem.Add_Click({
    $notify.Visible = $false
    $notify.Dispose()
    Get-Process python -ErrorAction SilentlyContinue | Where-Object { $_.CommandLine -like "*server.py*" } | Stop-Process -Force
    [System.Windows.Forms.Application]::Exit()
})

$notify.ContextMenuStrip = $contextMenu

# Double click vào tray icon để mở Dashboard
$notify.Add_DoubleClick({
    Start-Process "http://localhost:65534/router"
})

# Chạy vòng lặp ứng dụng
[System.Windows.Forms.Application]::Run()
