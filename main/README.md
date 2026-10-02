<div align="center">

# ⚡ TailRouter

**High-performance, zero-dependency bare-metal gateway & web dashboard for Tailscale Serve and Funnel.**

[![Release](https://img.shields.io/badge/Release-v2.1.0-blue.svg?logo=github)](https://github.com/giangsamne/TailRouter/releases)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8.svg?logo=go&logoColor=white)](https://golang.org)
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-brightgreen.svg)](https://github.com/giangsamne/TailRouter)
[![Tailscale](https://img.shields.io/badge/Tailscale-Serve%20%26%20Funnel-5056EC.svg?logo=tailscale&logoColor=white)](https://tailscale.com)
[![Dependencies](https://img.shields.io/badge/dependencies-0%20(Pure%20Native%20Binaries)-success.svg)](https://github.com/giangsamne/TailRouter)
[![Memory Footprint](https://img.shields.io/badge/RAM-%3C%208MB%20RSS-blueviolet.svg)](https://github.com/giangsamne/TailRouter)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

</div>

---

### ⚡ 1-Line Universal Auto-Installer (Recommended)

Run a single command in your terminal or PowerShell — it automatically detects your Operating System, hardware architecture (Apple Silicon / Intel / ARM64 / x86_64), and init system (`systemd`, `openrc`, `launchd`), then installs and starts TailRouter in one shot:

* **🐧 Linux & 🍏 macOS** (Terminal):
  ```bash
  curl -fsSL https://raw.githubusercontent.com/giangsamne/TailRouter/v2.1.0/install.sh | bash
  ```

* **🪟 Windows** (PowerShell):
  ```powershell
  irm https://raw.githubusercontent.com/giangsamne/TailRouter/v2.1.0/install.ps1 | iex
  ```

---

## 🆕 What's New in v2.1.0 & v1.1.0

Version **v2.1.0** (Go Native Bare-Metal) and **v1.1.0** (Python Edition) bring a completely redesigned management workflow, cleaner UI architecture, and robust system lifecycle controls:

### 1. ⚙️ 5th Tab: System Settings & Control (Settings Tab)
The Web Dashboard has been standardized with a dedicated 5th tab navigation. All configuration panels are neatly organized inside **⚙️ Settings**:
- **Clutter-Free Main Dashboard**: Relocated all persistent banners (`Tailscale Serve`, `Auto-Start on Boot`, `Operator Warning`) inside the Settings tab. The main overview dashboard remains sleek, uncluttered, and modern.
- **🌐 Remote Access via Tailscale Serve**: One-click enable/disable for `/router` mapped to your Tailscale MagicDNS domain over standard HTTPS (port 443 without typing `:65534`), with direct live link generation.
- **🚀 Auto-Start on Boot Management**: Easily configure background daemon autostart across Linux (`systemd`, `openrc`), macOS (`launchd`), and Windows (`Registry`).
- **⚠️ Danger Zone & System Reset**:
  - 🗑️ **Reset Router Configuration**: Clear all registered routes back to 0 and automatically turn off Tailscale Serve mappings.
  - 🧹 **Clear Proxy Logs**: Purge in-memory live traffic inspection logs.
  - 💥 **Complete App Uninstall & Data Wipe**: 
    - Automatically resets and turns off Tailscale Serve / Funnel.
    - Unregisters and cleans all autostart services across all operating systems.
    - Wipes configuration directories (`~/.config/tailrouter`), `routes.json`, and all log files.
    - Unlinks and removes the `tailrouter` binary executable.
    - Gracefully stops the Gateway process, frees port `65534`, and presents a confirmation screen on the browser.
- **ℹ️ Project & Repository Info**: Displays engine type, version, author **Giang3DLab / Giang Sam (MIT License)**, and direct link to the GitHub repository.

### 2. 🔍 Ascending Numerical Port Order (Auto-Scan)
- Active host TCP listening ports and Docker containers are **automatically sorted in ascending numeric order** (`22 -> 80 -> 443 -> 3000 -> 8080...`).
- Makes browsing, locating, and mapping host services fast and effortless.

### 3. ⌨️ Expanded CLI Commands
- `tailrouter stop`: Stop running gateway process occupying port 65534.
- `tailrouter uninstall [-y]`: Completely uninstall TailRouter and purge all data and binaries directly from the terminal.

### 4. 📡 New REST API Endpoints
- `POST /api/routes/reset`: Reset all routes to 0.
- `POST /api/logs/clear`: Purge proxy traffic logs.
- `POST /api/system/uninstall`: Full system uninstallation and graceful shutdown.

---

### 🌿 Dedicated OS Branches

TailRouter provides dedicated platform branches with optimized guides, raw data, and service setups:

| Branch | Platform | Features | Direct Link |
| :--- | :--- | :--- | :--- |
| **`main`** | 🌐 Multi-Platform (Hub) | Go Native v2.1.0, multi-OS installers, release artifacts | [**View `main` branch 👉**](https://github.com/giangsamne/TailRouter/tree/main) |
| **`linux`** | 🐧 Linux | Raw data & Linux scripts, systemd & OpenRC (Alpine/Ubuntu/Arch) | [**View `linux` branch 👉**](https://github.com/giangsamne/TailRouter/tree/linux) |
| **`windows`** | 🪟 Windows | Raw data & Windows scripts, PowerShell installer, Startup autostart | [**View `windows` branch 👉**](https://github.com/giangsamne/TailRouter/tree/windows) |
| **`macos`** | 🍏 macOS | Raw data & macOS scripts, launchd agent, Apple Silicon & Intel | [**View `macos` branch 👉**](https://github.com/giangsamne/TailRouter/tree/macos) |

---

## 💡 What is TailRouter?

**TailRouter** is an ultra-fast, zero-dependency bare-metal gateway running directly on port **65534** of your host machine. It eliminates the need for heavy reverse proxies (Nginx, Traefik, NPM) or virtual machines to publish local homelab services (3D printers, home automation, web apps, media servers, dev tools) to your **Tailnet** or the **public Internet**.

### 🌟 Key Highlights

- ⚡ **Zero Dependencies & Blazing Fast**:
  - Written in 100% pure static Go.
  - **No Python, no Node.js, and no external runtimes required**.
  - Starts in **< 10ms**, consumes **< 8 MB RAM**, and uses **0.0% idle CPU**.
- 🎛️ **Seamless Tailscale Serve & Funnel Integration**:
  - **Serve**: Encrypted, authenticated access strictly within your private Tailnet (`https://<node-name>.ts.net/<path>`).
  - **Funnel**: Fully public HTTPS route accessible across the Internet without port forwarding or public IPv4.
  - Switch between Serve and Funnel in 1 click or via CLI!
- 🖥️ **Embedded 5-Tab Web Dashboard**:
  - Responsive dark-mode SPA embedded inside the binary via `embed.FS`.
  - Accessible locally at `http://localhost:65534` (or `http://localhost:65534/router`) and over Tailnet at `https://<node-name>.ts.net/router`.
  - Full multi-language support: Tiếng Việt, English, 中文, 日本語.
- 🔍 **Host Port Discovery Scanner**:
  - Scans active TCP listening ports across the host system and Docker containers.
  - Automatically sorts detected ports in ascending numerical order (`22 -> 80 -> 443 -> ...`).
  - Probes HTTP titles and signatures automatically to pre-fill route names and paths.
- 🔁 **Persistent Route Storage**:
  - Automatically saves routes to standard system locations (`~/.config/tailrouter/routes.json` or `%APPDATA%\TailRouter\routes.json`).
  - Survives updates, binary moves, and reboots seamlessly.
- 🔄 **Native Boot Autostart**:
  - Supports `systemd` (Ubuntu, Debian, Fedora), `openrc` (Alpine Linux), `launchd` (macOS), and Windows Registry autostart via `tailrouter service install`.
- 💥 **Complete Clean Uninstall (One-Click / One-Command)**:
  - Both Web Dashboard and CLI support complete uninstallation: removes system services, resets Tailscale Serve, wipes configuration folders, removes logs, deletes the binary, and frees port 65534.

---

## 🖥️ System Architecture

```text
Internet / Tailnet
         │
         ▼
┌──────────────────────────────────────────────────────────────┐
│                  Tailscale WireGuard Mesh                     │
│         (MagicDNS: https://<your-node>.ts.net)               │
└──────────────────────────────┬───────────────────────────────┘
                               │
                ┌──────────────┴───────────────┐
                ▼ (Serve / Funnel)              ▼ (Web Dashboard /router)
┌──────────────────────────────────────────────────────────────┐
│ TailRouter Native Engine (Port 65534)                        │
│  - Dual-stack IPv4/IPv6 reverse proxy                         │
│  - REST API & Embedded Web Dashboard (5 Tabs + Settings)     │
│  - Host TCP port scanner (Sorted ascending)                  │
│  - Full self-uninstall & service manager                     │
└──────────────────────────────┬───────────────────────────────┘
                               │
    ┌──────────────────────────┼───────────────────────────────┐
    ▼                          ▼                               ▼
Port 22                     Port 443                        Port 8080
(SSH / Remote)             (Web App / Next.js)             (Bambu 3D / OctoPrint)
```

---

## 🚀 Quick Start Guide

### 1. Launching TailRouter
Simply double-click `tailrouter` (or `tailrouter.exe` on Windows), or run it in your terminal:
```bash
tailrouter
```
* The gateway automatically starts on port `65534`.
* Your default web browser immediately opens the Web Dashboard at:
  👉 **`http://localhost:65534/router`**

### 2. Running as a Permanent Background Service
To make TailRouter start automatically every time your computer boots:
```bash
tailrouter service install
```
* **Linux**: Configured via `systemd` user service or `openrc`.
* **macOS**: Configured via `launchd` user agent.
* **Windows**: Configured via User Run registry key.

### 3. Uninstalling TailRouter Completely
If you ever want to completely remove TailRouter from your machine:
- **Via Web Dashboard**: Navigate to **⚙️ Settings** -> **Danger Zone** -> Click **💥 Uninstall & Completely Remove TailRouter from Machine**.
- **Via CLI Terminal**:
  ```bash
  tailrouter uninstall -y
  ```

---

## ⌨️ CLI Command Reference

`tailrouter` provides an intuitive command-line interface:

```bash
# Run gateway (starts server and opens browser if no args provided)
tailrouter
tailrouter run [--port 65534] [--open]

# Stop running gateway daemon on port 65534
tailrouter stop

# Inspect running gateway status and Tailscale connection
tailrouter status

# Scan active TCP ports on the host (sorted ascending)
tailrouter scan

# List all configured routes
tailrouter routes list

# Add a new route
tailrouter routes add <name> <path> <port> [serve|funnel]
# Example:
tailrouter routes add "Bambu 3D Printer" /bambu 8080 serve

# Delete a route by ID or path
tailrouter routes delete <id|path>
# Example:
tailrouter routes delete /bambu

# Automatic Tailscale Serve configuration
tailrouter serve router     # Expose TailRouter web interface over Tailscale Serve (/router)
tailrouter serve gateway    # Forward entire host gateway via Tailscale Serve
tailrouter serve reset      # Reset Tailscale Serve configuration

# System service management
tailrouter service install    # Enable automatic start on boot (systemd/openrc/launchd)
tailrouter service status     # Check background service health
tailrouter service uninstall  # Remove system service autostart

# Complete app uninstallation & data wipe
tailrouter uninstall [-y]     # Stop gateway, reset serve, remove autostart, wipe configs & remove binary
```

---

## 📡 REST API Reference

TailRouter exposes a high-performance REST API on port `65534`:

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/api/status` | Node status, Tailscale state, gateway uptime, autostart info. |
| `GET` | `/api/routes` | List all configured routes with latency and hit counts. |
| `POST` | `/api/routes` | Register a new reverse-proxy route. |
| `GET` | `/api/routes/{id}` | Inspect details of a specific route. |
| `PUT` | `/api/routes/{id}` | Update route target, name, or settings. |
| `DELETE` | `/api/routes/{id}` | Delete a route and unmap it from Tailscale. |
| `POST` | `/api/routes/{id}/ping` | Health-check backend target and return latency in ms. |
| `POST` | `/api/routes/{id}/toggle` | Enable or disable a route without deleting it. |
| `POST` | `/api/routes/{id}/toggle-mode` | Switch route between `serve` (private) and `funnel` (public). |
| `POST` | `/api/routes/reset` | 🗑️ Reset all routes to 0 and reset Tailscale Serve. |
| `GET` | `/api/logs` | Fetch recent live proxy traffic logs. |
| `POST` | `/api/logs/clear` | 🧹 Clear all proxy traffic logs. |
| `GET` | `/api/scan` | Trigger immediate host TCP port discovery scan (sorted ascending). |
| `GET` | `/api/autostart` | Inspect operating system autostart configuration. |
| `POST` | `/api/autostart` | Enable or disable autostart on boot. |
| `POST` | `/api/system/uninstall` | 💥 Full app uninstall, wipe all data/configs/binary, and exit gracefully. |

---

## 📁 Repository Structure

```text
TailRouter/
├── cmd/
│   └── tailrouter/           # Single-binary engine: Daemon + Embedded Web UI + CLI (Go)
├── internal/
│   ├── gateway/              # Reverse proxy, REST API & embedded web server (5 tabs + settings)
│   ├── scanner/              # Host TCP port discovery engine (sorted ascending)
│   ├── config/               # JSON route persistence with standard system path resolution
│   ├── service/              # Systemd, OpenRC, Launchd, and Windows autostart & uninstall
│   ├── tailscale/            # Tailscale Serve & Funnel CLI integration
│   └── browser/              # Cross-platform browser launcher
├── web/
│   ├── embed.go              # Go binary asset packaging via embed.FS
│   └── index.html            # Responsive SPA Web Dashboard (5 Tabs, i18n, Danger Zone)
├── scripts/
│   ├── build_all.sh          # Multi-platform static compilation tool
│   └── snapshot_tag.sh       # Release snapshot packaging tool
├── install.sh                # 1-Line universal installer for Linux & macOS (v2.1.0)
├── install.ps1               # 1-Line installer for Windows (PowerShell v2.1.0)
├── LICENSE                   # MIT License
└── README.md                 # Project documentation & upgrade guides
```

---

## 🤝 Contributing

Contributions, issues, and feature requests are welcome!
Feel free to open an issue or submit a pull request on the [GitHub Issues page](https://github.com/giangsamne/TailRouter/issues).

---

## 📜 License

Distributed under the **MIT License**. Author: **Giang3DLab / Giang Sam**. See [LICENSE](LICENSE) for more details.
