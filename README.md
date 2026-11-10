<div align="center">

# ⚡ TailRouter

**High-performance, zero-dependency bare-metal gateway & web dashboard for Tailscale Serve and Funnel.**

[![Release](https://img.shields.io/badge/Release-v2.3.0-blue.svg?logo=github)](https://github.com/giangsamne/TailRouter/releases)
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
  curl -fsSL https://raw.githubusercontent.com/giangsamne/TailRouter/v2.3.0/install.sh | bash
  ```

* **🪟 Windows** (PowerShell):
  ```powershell
  irm https://raw.githubusercontent.com/giangsamne/TailRouter/v2.3.0/install.ps1 | iex
  ```

---

### 🚀 Pre-compiled Single-File Binaries & Archives

Tải trực tiếp bản nhị phân đóng gói độc lập cho từng hệ điều hành và kiến trúc từ [GitHub Releases](https://github.com/giangsamne/TailRouter/releases):

| Nền tảng (Platform) | Kiến trúc CPU (Architecture) | Tệp đóng gói (1-File Package) |
| :--- | :--- | :--- |
| **🐧 Linux** | AMD64 (x86_64 PC, VPS, Server) | [`tailrouter-linux-amd64.tar.gz`](https://github.com/giangsamne/TailRouter/releases/download/v2.3.0/tailrouter-linux-amd64.tar.gz) |
| **🐧 Linux** | ARM64 (Raspberry Pi 4/5, Cloud ARM) | [`tailrouter-linux-arm64.tar.gz`](https://github.com/giangsamne/TailRouter/releases/download/v2.3.0/tailrouter-linux-arm64.tar.gz) |
| **🍏 macOS** | Apple Silicon (M1 / M2 / M3 / M4) | [`tailrouter-darwin-arm64.tar.gz`](https://github.com/giangsamne/TailRouter/releases/download/v2.3.0/tailrouter-darwin-arm64.tar.gz) |
| **🍏 macOS** | Intel (x86_64) | [`tailrouter-darwin-amd64.tar.gz`](https://github.com/giangsamne/TailRouter/releases/download/v2.3.0/tailrouter-darwin-amd64.tar.gz) |
| **🪟 Windows** | AMD64 (x64) | [`tailrouter-windows-amd64.zip`](https://github.com/giangsamne/TailRouter/releases/download/v2.3.0/tailrouter-windows-amd64.zip) |

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
- 🔐 **10-Minute Password & Session Security (v2.3.0)**:
  - Secure master password protected with SHA-256 + 16-byte random salt, stored in `auth.json` with strict `0600` file permissions.
  - Generates 10-minute temporary emergency OTP passwords (`TR-XXXXXX`) via CLI (`tailrouter password temp`) or Settings tab.
  - Strictly enforces 10-minute sessions (`tailrouter_session` HttpOnly SameSite=Strict cookie + `X-TailRouter-Auth` header). Real-time countdown badge (`⏳ Phiên: 10:00`) in header with 10-minute extend button and instant lock.
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

# Password & 10-minute session security management
tailrouter password status           # Check security & temporary OTP status
tailrouter password set <password>   # Set master admin password
tailrouter password temp             # Generate emergency 10-minute temporary OTP (TR-XXXXXX)
tailrouter password clear            # Disable password protection

# Complete app uninstallation & data wipe
tailrouter uninstall [-y]     # Stop gateway, reset serve, remove autostart, wipe configs & remove binary
```

---

## 📡 REST API Reference

TailRouter exposes a high-performance REST API on port `65534`:

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/api/status` | Node status, Tailscale state, gateway uptime, autostart info. |
| `GET` | `/api/auth/status` | Public status check (whether auth is enabled/authenticated/temp OTP active). |
| `POST` | `/api/auth/login` | Public login verifying master password or 10-min OTP; issues 10-min session cookie & token. |
| `POST` | `/api/auth/logout` | Revokes current session token and clears session cookie. |
| `POST` | `/api/auth/extend` | Extends active session duration by another 10 minutes. |
| `POST` | `/api/auth/set_password` | Configures new master password (requires session if auth enabled). |
| `POST` | `/api/auth/temp_password` | Generates a 10-minute OTP temporary password. |
| `POST` | `/api/auth/disable` | Disables password authentication. |
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
│   ├── auth/                 # 10-minute password & session manager (SHA-256+Salt, OTP, cookies)
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
