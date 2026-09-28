<div align="center">

# 🪟 TailRouter for Windows

**High-performance Native CLI Engine & Web Control (:65534) for Windows 10, Windows 11 & Tiny 11.**

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8.svg?logo=go&logoColor=white)](https://golang.org)
[![Windows](https://img.shields.io/badge/Windows-10%20%7C%2011%20%7C%20Tiny11%20%7C%20Server-0078D6.svg?logo=windows&logoColor=white)](https://github.com/giangsamne/TailRouter/tree/windows)
[![Tailscale](https://img.shields.io/badge/Tailscale-Serve%20%26%20Funnel-5056EC.svg?logo=tailscale&logoColor=white)](https://tailscale.com)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

</div>

---

### ⚡ 1-Line Universal Install for Windows (PowerShell)

Run this single command in PowerShell (Administrator or standard user):

```powershell
irm https://raw.githubusercontent.com/giangsamne/TailRouter/v2.2.0/install.ps1 | iex
```

* **Zero Dependencies**: Pure static Go binary (`tailrouter.exe`).
* **Tiny 11 & Low-spec Friendly**: Consumes < 8MB RAM and 0.0% CPU when idle. No .NET runtime, bulky Electron framework, or desktop wrappers.
* **Startup Autostart**: Automatically configures background startup on Windows login.
* **Direct Web Control**: Immediately ready at: **`http://localhost:65534/router`**.

---

### 📦 Manual Download (Standalone Zip)

Download the pre-compiled binary package:
* [**`TailRouter-Windows.zip` (GitHub Release)**](https://github.com/giangsamne/TailRouter/releases/download/v2.2.0/TailRouter-Windows.zip)

To install or run manually:
1. Extract `TailRouter-Windows.zip`.

---

### 🖥️ Dual Control Architecture: CLI + Web Control (:65534)

TailRouter on Windows runs as a single lightweight background process providing both CLI and Web interfaces:

#### 1. 🌐 Web Control Dashboard
* Access locally at: **`http://localhost:65534/router`**
* Or over your Tailnet: **`https://<your-tailscale-node>.ts.net/router`**
* Visually add routes, toggle Serve/Funnel modes, scan ports, and monitor backend latency.

#### 2. ⌨️ Windows CLI Commands (CMD / PowerShell)
```powershell
# Start Gateway in foreground and open browser dashboard
tailrouter.exe

# Inspect Gateway health and Tailscale connection
tailrouter.exe status

# Scan open TCP ports across the Windows machine
tailrouter.exe scan

# List all configured routes
tailrouter.exe routes list

# Add a new reverse-proxy route
tailrouter.exe routes add "Bambu Printer" /bambu 8080 serve

# Delete a route
tailrouter.exe routes delete /bambu

# Setup / Remove startup registry
tailrouter.exe service install    # Adds to HKCU\Software\Microsoft\Windows\CurrentVersion\Run
tailrouter.exe service uninstall  # Removes from Run registry
```

---

### 📂 Persistent Configuration

Route definitions are reliably saved to:
* `%APPDATA%\TailRouter\routes.json`
* Survives updates, binary replacements, and Windows reboots.

---

### 🌿 Related Branches
* [**`main`**](https://github.com/giangsamne/TailRouter/tree/main): Multi-platform hub & documentation.
* [**`linux`**](https://github.com/giangsamne/TailRouter/tree/linux): Linux native CLI & Web Control.
* [**`macos`**](https://github.com/giangsamne/TailRouter/tree/macos): macOS native CLI & Web Control.
